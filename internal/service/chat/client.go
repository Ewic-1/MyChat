package chat

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"mychat_server/internal/config"
	"mychat_server/internal/dao"
	"mychat_server/internal/dto/request"
	"mychat_server/internal/model"
	mykafka "mychat_server/internal/service/kafka"
	"mychat_server/pkg/constants"
	"mychat_server/pkg/enum/contact/contact_status_enum"
	"mychat_server/pkg/utils/random"
	"mychat_server/pkg/utils/zlog"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  2048,
	WriteBufferSize: 2048,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type client struct {
	ID        string
	Conn      *websocket.Conn
	writeMu   sync.Mutex
	closeOnce sync.Once
}

func (c *client) WriteJSON(v any) error {
	c.writeMu.Lock()
	defer c.writeMu.Unlock()
	return c.Conn.WriteJSON(v)
}

func (c *client) Close() error {
	var err error
	c.closeOnce.Do(func() {
		err = c.Conn.Close()
	})
	return err
}

type wsMessage struct {
	SessionId  string `json:"session_id"`
	Type       int8   `json:"type"`
	Content    string `json:"content"`
	Url        string `json:"url"`
	SendId     string `json:"send_id"`
	SendName   string `json:"send_name"`
	SendAvatar string `json:"send_avatar"`
	ReceiveId  string `json:"receive_id"`
	FileSize   string `json:"file_size"`
	FileType   string `json:"file_type"`
	FileName   string `json:"file_name"`
	CreatedAt  string `json:"created_at"`
	AVdata     string `json:"av_data,omitempty"`
}

type chatServer struct {
	Clients      map[string]*client
	clientsMu    sync.RWMutex
	runtimeOnce  sync.Once
	contactDao   dao.ContactInfoDao
	messageDao   dao.MessageDao
	kafkaEnabled bool
}

func newChatServer() *chatServer {
	return &chatServer{Clients: make(map[string]*client)}
}

var ChatServer = newChatServer()

func InitRuntime() {
	ChatServer.InitRuntime()
}

func StopRuntime() {
	ChatServer.StopRuntime()
}

func (s *chatServer) InitRuntime() {
	s.runtimeOnce.Do(func() {
		mode := strings.ToLower(strings.TrimSpace(config.GetConfig().KafkaConfig.MessageMode))
		s.kafkaEnabled = mode == "kafka"
		if !s.kafkaEnabled {
			zlog.Info("chat runtime started with channel mode")
			return
		}

		mykafka.KafkaService.KafkaInit()
		go s.consumeKafkaMessages()
		zlog.Info("chat runtime started with kafka mode")
	})
}

func (s *chatServer) StopRuntime() {
	if s.kafkaEnabled {
		mykafka.KafkaService.KafkaClose()
	}
}

func (s *chatServer) consumeKafkaMessages() {
	for {
		msg, err := mykafka.KafkaService.ReadChatMessage(context.Background())
		if err != nil {
			zlog.Error("read kafka chat message failed: " + err.Error())
			time.Sleep(500 * time.Millisecond)
			continue
		}

		var out wsMessage
		if err := json.Unmarshal(msg.Value, &out); err != nil {
			zlog.Error("unmarshal kafka chat message failed: " + err.Error())
			continue
		}
		s.dispatchMessage(out)
	}
}

func (s *chatServer) addClient(c *client) {
	s.clientsMu.Lock()
	oldClient := s.Clients[c.ID]
	s.Clients[c.ID] = c
	s.clientsMu.Unlock()

	if oldClient != nil {
		_ = oldClient.Close()
	}
}

func (s *chatServer) getClient(clientId string) *client {
	s.clientsMu.RLock()
	defer s.clientsMu.RUnlock()
	return s.Clients[clientId]
}

func (s *chatServer) removeClient(clientId string) *client {
	s.clientsMu.Lock()
	defer s.clientsMu.Unlock()
	oldClient := s.Clients[clientId]
	delete(s.Clients, clientId)
	return oldClient
}

func (s *chatServer) readLoop(c *client) {
	defer func() {
		s.removeClient(c.ID)
		_ = c.Close()
		zlog.Info("ws连接断开: " + c.ID)
	}()

	for {
		var req request.ChatMessageRequest
		if err := c.Conn.ReadJSON(&req); err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				zlog.Error("ws read failed: " + err.Error())
			}
			return
		}

		s.handleIncomingMessage(c.ID, req)
	}
}

func (s *chatServer) handleIncomingMessage(clientId string, req request.ChatMessageRequest) {
	if strings.TrimSpace(req.ReceiveId) == "" {
		zlog.Warn("ignore message with empty receive_id")
		return
	}
	if strings.TrimSpace(req.SendId) == "" {
		req.SendId = clientId
	}
	if req.SendId != clientId {
		zlog.Warn("ws send_id mismatch, reset to authenticated client")
		req.SendId = clientId
	}

	message := buildModelMessage(req)
	msg, ret := s.messageDao.SaveMessage(message)
	if ret != 0 {
		zlog.Error("save message failed: " + msg)
		return
	}

	out := buildWsMessage(message)
	if s.kafkaEnabled {
		payload, err := json.Marshal(out)
		if err != nil {
			zlog.Error("marshal kafka message failed: " + err.Error())
			s.dispatchMessage(out)
			return
		}
		if err := mykafka.KafkaService.PublishChatMessage(out.ReceiveId, payload); err != nil {
			zlog.Error("publish kafka message failed: " + err.Error())
			s.dispatchMessage(out)
			return
		}
		return
	}

	s.dispatchMessage(out)
}

func buildModelMessage(req request.ChatMessageRequest) model.Message {
	now := time.Now()
	return model.Message{
		Uuid:       fmt.Sprintf("M%s", random.GetNowAndLenRandomString(11)),
		SessionId:  req.SessionId,
		Type:       req.Type,
		Content:    req.Content,
		URL:        req.Url,
		SendId:     req.SendId,
		SendName:   req.SendName,
		SendAvatar: req.SendAvatar,
		ReceiveId:  req.ReceiveId,
		FileType:   req.FileType,
		FileName:   req.FileName,
		FileSize:   req.FileSize,
		AVData:     req.AVdata,
		Status:     0,
		CreatedAt:  now,
	}
}

func buildWsMessage(message model.Message) wsMessage {
	return wsMessage{
		SessionId:  message.SessionId,
		Type:       message.Type,
		Content:    message.Content,
		Url:        message.URL,
		SendId:     message.SendId,
		SendName:   message.SendName,
		SendAvatar: message.SendAvatar,
		ReceiveId:  message.ReceiveId,
		FileSize:   message.FileSize,
		FileType:   message.FileType,
		FileName:   message.FileName,
		CreatedAt:  message.CreatedAt.Format("2006-01-02 15:04:05"),
		AVdata:     message.AVData,
	}
}

func (s *chatServer) dispatchMessage(message wsMessage) {
	targets := s.resolveTargets(message)
	for clientId := range targets {
		s.sendToClient(clientId, message)
	}
}

func (s *chatServer) resolveTargets(message wsMessage) map[string]struct{} {
	targets := make(map[string]struct{})

	if strings.HasPrefix(message.ReceiveId, "G") {
		msg, members, ret := s.contactDao.GetContactByContactId(message.ReceiveId)
		if ret != 0 {
			zlog.Error("query group members failed: " + msg)
		} else {
			for _, member := range members {
				if canReceiveGroupMessage(member.Status) {
					targets[member.UserId] = struct{}{}
				}
			}
		}
	} else {
		targets[message.ReceiveId] = struct{}{}
	}

	targets[message.SendId] = struct{}{}
	return targets
}

func canReceiveGroupMessage(status int8) bool {
	return status == contact_status_enum.NORMAL || status == contact_status_enum.SILENCE
}

func (s *chatServer) sendToClient(clientId string, message wsMessage) {
	cli := s.getClient(clientId)
	if cli == nil {
		return
	}

	if err := cli.WriteJSON(message); err != nil {
		zlog.Error("send ws message failed: " + err.Error())
		staleClient := s.removeClient(clientId)
		if staleClient != nil {
			_ = staleClient.Close()
		}
	}
}

// NewClientInit 当接受到前端有登录消息时，会调用该函数
func NewClientInit(c *gin.Context, clientId string) {
	ChatServer.InitRuntime()

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		zlog.Error(err.Error())
		return
	}

	cli := &client{ID: clientId, Conn: conn}
	ChatServer.addClient(cli)
	zlog.Info("ws连接成功: " + clientId)
	ChatServer.readLoop(cli)
}

// ClientLogout 当接受到前端有登出消息时，会调用该函数
func ClientLogout(clientId string) (string, int) {
	client := ChatServer.removeClient(clientId)
	if client != nil {
		if err := client.Close(); err != nil {
			zlog.Error(err.Error())
			return constants.SYSTEM_ERROR, -1
		}
	}
	return "退出成功", 0
}

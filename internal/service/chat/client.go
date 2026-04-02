package chat

import (
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"mychat_server/internal/dto/request"
	"mychat_server/pkg/constants"
	"mychat_server/pkg/utils/zlog"
)

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

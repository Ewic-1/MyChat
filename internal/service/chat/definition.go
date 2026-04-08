package chat

import (
	"context"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"

	"mychat_server/internal/dao"
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
	Clients              map[string]*client
	clientsMu            sync.RWMutex
	runtimeOnce          sync.Once
	runtimeCtx           context.Context
	runtimeCancel        context.CancelFunc
	runtimeWg            sync.WaitGroup
	dispatchQueue        chan wsMessage
	contactDao           dao.ContactInfoDao
	messageDao           dao.MessageDao
	kafkaEnabled         bool
	hybridEnabled        bool
	highWatermarkRatio   float64
	highWatermarkSeconds int
	offloadPercent       int
	offloadActive        int32
	offloadSeq           uint64
}

func newChatServer() *chatServer {
	return &chatServer{Clients: make(map[string]*client)}
}

var ChatServer = newChatServer()

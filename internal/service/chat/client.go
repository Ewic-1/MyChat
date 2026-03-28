package chat

import (
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"mychat_server/internal/config"
	"mychat_server/pkg/constants"
	"mychat_server/pkg/utils/zlog"
	"net/http"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  2048,
	WriteBufferSize: 2048,
	// 检查连接的Origin头
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// NewClientInit 当接受到前端有登录消息时，会调用该函数
func NewClientInit(c *gin.Context, clientId string) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		zlog.Error(err.Error())
	}

	zlog.Info("ws连接成功")
}

// ClientLogout 当接受到前端有登出消息时，会调用该函数
func ClientLogout(clientId string) (string, int) {
	client := ChatServer.Clients[clientId]
	if client != nil {
		if err := client.Conn.Close(); err != nil {
			zlog.Error(err.Error())
			return constants.SYSTEM_ERROR, -1
		}

	}
	return "退出成功", 0
}

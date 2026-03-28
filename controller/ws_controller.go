package controller

import (
	"mychat_server/internal/dto/request"
	"mychat_server/internal/service/chat"
	"mychat_server/pkg/constants"
	"mychat_server/pkg/utils/zlog"
	"net/http"

	"github.com/gin-gonic/gin"
)

// WsLogin wss登录 Get
func WsLogin(c *gin.Context) {
	clientId := c.Query("client_id")
	if clientId == "" {
		zlog.Error("clientId获取失败")
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": "clientId获取失败",
		})
		return
	}
	chat.NewClientInit(c, clientId)
}

// WsLogout wss登出
func WsLogout(c *gin.Context) {
	var req request.WsLogoutRequest
	if err := c.BindJSON(&req); err != nil {
		zlog.Error(err.Error())
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": constants.SYSTEM_ERROR,
		})
		return
	}
	message, ret := chat.ClientLogout(req.OwnerId)
	JsonBack(c, message, ret, nil)
}

package controller

import (
	"mychat_server/internal/dto/request"
	"mychat_server/internal/service/gorm"
	"mychat_server/pkg/constants"
	"net/http"

	"github.com/gin-gonic/gin"
)

var messageService gorm.MessageService

// GetMessageList 获取聊天记录
func GetMessageList(c *gin.Context) {
	var req request.GetMessageListRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": constants.SYSTEM_ERROR,
		})
		return
	}
	message, rsp, ret := messageService.GetMessageList(req.UserOneId, req.UserTwoId)
	JsonBack(c, message, ret, rsp)
}

// GetGroupMessageList 获取群聊消息记录
func GetGroupMessageList(c *gin.Context) {
	var req request.GetGroupMessageListRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": constants.SYSTEM_ERROR,
		})
		return
	}
	message, rsp, ret := messageService.GetGroupMessageList(req.GroupId)
	JsonBack(c, message, ret, rsp)
}

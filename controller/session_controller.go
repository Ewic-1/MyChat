package controller

import (
	"mychat_server/internal/dto/request"
	"mychat_server/internal/service/gorm"
	"mychat_server/pkg/constants"
	"mychat_server/pkg/utils/zlog"
	"net/http"

	"github.com/gin-gonic/gin"
)

var sessionService = new(gorm.SessionService)

// OpenSession 打开会话
func OpenSession(c *gin.Context) {
	var openSessionReq request.OpenSessionRequest
	if err := c.BindJSON(&openSessionReq); err != nil {
		zlog.Error(err.Error())
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": constants.SYSTEM_ERROR,
		})
		return
	}
	message, sessionId, ret := sessionService.OpenSession(openSessionReq)
	JsonBack(c, message, ret, sessionId)
}

// GetUserSessionList 获取用户会话列表
func GetUserSessionList(c *gin.Context) {
	var getUserSessionListReq request.OwnlistRequest
	if err := c.BindJSON(&getUserSessionListReq); err != nil {
		zlog.Error(err.Error())
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": constants.SYSTEM_ERROR,
		})
		return
	}
	message, sessionList, ret := sessionService.GetUserSessionList(getUserSessionListReq.OwnerId)
	JsonBack(c, message, ret, sessionList)
}

// GetGroupSessionList 获取用户群聊列表
func GetGroupSessionList(c *gin.Context) {
	var getGroupListReq request.OwnlistRequest
	if err := c.BindJSON(&getGroupListReq); err != nil {
		zlog.Error(err.Error())
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": constants.SYSTEM_ERROR,
		})
		return
	}
	message, groupList, ret := sessionService.GetGroupSessionList(getGroupListReq.OwnerId)
	JsonBack(c, message, ret, groupList)
}

// DeleteSession 删除会话
func DeleteSession(c *gin.Context) {
	var deleteSessionReq request.DeleteSessionRequest
	if err := c.BindJSON(&deleteSessionReq); err != nil {
		zlog.Error(err.Error())
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": constants.SYSTEM_ERROR,
		})
		return
	}
	message, ret := sessionService.DeleteSession(deleteSessionReq.OwnerId, deleteSessionReq.SessionId)
	JsonBack(c, message, ret, nil)
}

// CheckOpenSessionAllowed 检查是否可以打开会话
func CheckOpenSessionAllowed(c *gin.Context) {
	var req request.CreateSessionRequest
	if err := c.BindJSON(&req); err != nil {
		zlog.Error(err.Error())
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": constants.SYSTEM_ERROR,
		})
		return
	}
	message, res, ret := sessionService.CheckOpenSessionAllowed(req.SendId, req.ReceiveId)
	JsonBack(c, message, ret, res)
}

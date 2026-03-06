package controller

import (
	"mychat_server/internal/dto/request"
	"mychat_server/internal/service/gorm"
	"mychat_server/pkg/constants"
	"mychat_server/pkg/utils/zlog"
	"net/http"

	"github.com/gin-gonic/gin"
)

var groupInfoService = new(gorm.GroupInfoService)

func CreateGroup(c *gin.Context) {
	var req request.CreateGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		zlog.Error(err.Error())
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": constants.SYSTEM_ERROR,
		})
		return
	}
	msg, ret := groupInfoService.CreateGroup(req)
	JsonBack(c, msg, ret, nil)
}

// LoadMyGroup 获取我创建的群聊
func LoadMyGroup(c *gin.Context) {
	var loadMyGroupReq request.OwnlistRequest
	if err := c.BindJSON(&loadMyGroupReq); err != nil {
		zlog.Error(err.Error())
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": constants.SYSTEM_ERROR,
		})
		return
	}
	message, groupList, ret := groupInfoService.LoadMyGroup(loadMyGroupReq.OwnerId)
	JsonBack(c, message, ret, groupList)
}

// CheckGroupAddMode 检查群聊加群方式
func CheckGroupAddMode(c *gin.Context) {
	var req request.CheckGroupAddModeRequest
	if err := c.BindJSON(&req); err != nil {
		zlog.Error(err.Error())
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": constants.SYSTEM_ERROR,
		})
		return
	}
	message, addMode, ret := groupInfoService.CheckGroupAddMode(req.GroupId)
	JsonBack(c, message, ret, addMode)
}

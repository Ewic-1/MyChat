package controller

import (
	"mychat_server/internal/dto/request"
	"mychat_server/internal/service/gorm"
	"mychat_server/pkg/constants"
	"mychat_server/pkg/utils/zlog"
	"net/http"

	"github.com/gin-gonic/gin"
)

var userInfoService = new(gorm.UserInfoService)

// 登录
func Login(c *gin.Context) {
	var loginReq request.LoginRequest
	if err := c.ShouldBindJSON(&loginReq); err != nil {
		zlog.Error(err.Error())
		JsonBack(c, "参数错误", -2, nil)
		return
	}
	message, data, ret := userInfoService.Login(loginReq)
	JsonBack(c, message, ret, data)
}

// SendSmsCode 发送短信验证码
func SendSmsCode(c *gin.Context) {
	var req request.SendSmsCodeRequest
	if err := c.BindJSON(&req); err != nil {
		zlog.Error(err.Error())
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": constants.SYSTEM_ERROR,
		})
		return
	}
	message, ret := userInfoService.SendSmsCode(req.Telephone)
	JsonBack(c, message, ret, nil)
}

// Register  注册
func Register(c *gin.Context) {
	var req request.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		zlog.Error(err.Error())
		JsonBack(c, "参数错误", -2, nil)
		return
	}
	message, registerRespond, ret := userInfoService.Register(req)
	JsonBack(c, message, ret, registerRespond)
}

// SmsLogin 验证码登录
func SmsLogin(c *gin.Context) {
	var req request.SmsLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		zlog.Error(err.Error())
		JsonBack(c, "参数错误", -2, nil)
		return
	}
	msg, loginRespond, ret := userInfoService.SmsLogin(req)
	JsonBack(c, msg, ret, loginRespond)
}

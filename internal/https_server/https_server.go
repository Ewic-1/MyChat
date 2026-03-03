package https_server

import (
	"mychat_server/controller"
	"mychat_server/internal/config"
	"mychat_server/internal/middleware"
	"mychat_server/pkg/ssl"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

var GE *gin.Engine

func init() {
	GE = gin.Default()
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowOrigins = []string{"*"}
	corsConfig.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	corsConfig.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization"}
	GE.Use(cors.New(corsConfig))
	GE.Use(ssl.TlsHandler(config.GetConfig().MainConfig.Host, config.GetConfig().MainConfig.Port))
	GE.Static("/static/avatars", config.GetConfig().StaticAvatarPath)
	GE.Static("/static/files", config.GetConfig().StaticFilePath)

	// 公开路由：无需 JWT。
	GE.POST("/login", controller.Login)
	GE.POST("/register", controller.Register)
	GE.POST("/user/sendSmsCode", controller.SendSmsCode)
	GE.POST("/user/smsLogin", controller.SmsLogin)
	GE.POST("/auth/refresh", controller.RefreshToken)

	// 受保护路由：统一经过 JWT 鉴权中间件。
	protected := GE.Group("/")
	protected.Use(middleware.JWTAuthMiddleware())

	protected.POST("/user/updateUserInfo", controller.UpdateUserInfo)
	protected.POST("/user/getUserInfoList", controller.GetUserInfoList)
	protected.POST("/user/ableUsers", controller.AbleUsers)
	protected.POST("/user/getUserInfo", controller.GetUserInfo)
	protected.POST("/user/disableUsers", controller.DisableUsers)
	protected.POST("/user/deleteUsers", controller.DeleteUsers)
	protected.POST("/user/setAdmin", controller.SetAdmin)
	protected.POST("/user/wsLogout", controller.WsLogout)
	protected.POST("/group/createGroup", controller.CreateGroup)
	protected.POST("/group/loadMyGroup", controller.LoadMyGroup)
	protected.POST("/group/checkGroupAddMode", controller.CheckGroupAddMode)
	protected.POST("/group/enterGroupDirectly", controller.EnterGroupDirectly)
	protected.POST("/group/leaveGroup", controller.LeaveGroup)
	protected.POST("/group/dismissGroup", controller.DismissGroup)
	protected.POST("/group/getGroupInfo", controller.GetGroupInfo)
	protected.POST("/group/getGroupInfoList", controller.GetGroupInfoList)
	protected.POST("/group/deleteGroups", controller.DeleteGroups)
	protected.POST("/group/setGroupsStatus", controller.SetGroupsStatus)
	protected.POST("/group/updateGroupInfo", controller.UpdateGroupInfo)
	protected.POST("/group/getGroupMemberList", controller.GetGroupMemberList)
	protected.POST("/group/removeGroupMembers", controller.RemoveGroupMembers)
	protected.POST("/session/openSession", controller.OpenSession)
	protected.POST("/session/getUserSessionList", controller.GetUserSessionList)
	protected.POST("/session/getGroupSessionList", controller.GetGroupSessionList)
	protected.POST("/session/deleteSession", controller.DeleteSession)
	protected.POST("/session/checkOpenSessionAllowed", controller.CheckOpenSessionAllowed)
	protected.POST("/contact/getUserList", controller.GetUserList)
	protected.POST("/contact/loadMyJoinedGroup", controller.LoadMyJoinedGroup)
	protected.POST("/contact/getContactInfo", controller.GetContactInfo)
	protected.POST("/contact/deleteContact", controller.DeleteContact)
	protected.POST("/contact/applyContact", controller.ApplyContact)
	protected.POST("/contact/getNewContactList", controller.GetNewContactList)
	protected.POST("/contact/passContactApply", controller.PassContactApply)
	protected.POST("/contact/blackContact", controller.BlackContact)
	protected.POST("/contact/cancelBlackContact", controller.CancelBlackContact)
	protected.POST("/contact/getAddGroupList", controller.GetAddGroupList)
	protected.POST("/contact/refuseContactApply", controller.RefuseContactApply)
	protected.POST("/contact/blackApply", controller.BlackApply)
	protected.POST("/message/getMessageList", controller.GetMessageList)
	protected.POST("/message/getGroupMessageList", controller.GetGroupMessageList)
	protected.POST("/message/uploadAvatar", controller.UploadAvatar)
	protected.POST("/message/uploadFile", controller.UploadFile)
	protected.POST("/chatroom/getCurContactListInChatRoom", controller.GetCurContactListInChatRoom)
	protected.POST("/auth/logout", controller.Logout)

	// WebSocket 路由先保持原行为
	GE.GET("/wss", controller.WsLogin)

}

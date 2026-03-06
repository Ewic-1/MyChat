package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func notImplemented(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{
		"code":    501,
		"message": "not implemented yet",
	})
}

func GetUserInfoList(c *gin.Context) { notImplemented(c) }
func AbleUsers(c *gin.Context)       { notImplemented(c) }

func DisableUsers(c *gin.Context) { notImplemented(c) }
func DeleteUsers(c *gin.Context)  { notImplemented(c) }
func SetAdmin(c *gin.Context)     { notImplemented(c) }

func WsLogout(c *gin.Context) { notImplemented(c) }

func CheckGroupAddMode(c *gin.Context)           { notImplemented(c) }
func EnterGroupDirectly(c *gin.Context)          { notImplemented(c) }
func LeaveGroup(c *gin.Context)                  { notImplemented(c) }
func DismissGroup(c *gin.Context)                { notImplemented(c) }
func GetGroupInfo(c *gin.Context)                { notImplemented(c) }
func GetGroupInfoList(c *gin.Context)            { notImplemented(c) }
func DeleteGroups(c *gin.Context)                { notImplemented(c) }
func SetGroupsStatus(c *gin.Context)             { notImplemented(c) }
func UpdateGroupInfo(c *gin.Context)             { notImplemented(c) }
func GetGroupMemberList(c *gin.Context)          { notImplemented(c) }
func RemoveGroupMembers(c *gin.Context)          { notImplemented(c) }
func OpenSession(c *gin.Context)                 { notImplemented(c) }
func GetUserSessionList(c *gin.Context)          { notImplemented(c) }
func GetGroupSessionList(c *gin.Context)         { notImplemented(c) }
func DeleteSession(c *gin.Context)               { notImplemented(c) }
func CheckOpenSessionAllowed(c *gin.Context)     { notImplemented(c) }
func GetUserList(c *gin.Context)                 { notImplemented(c) }
func LoadMyJoinedGroup(c *gin.Context)           { notImplemented(c) }
func GetContactInfo(c *gin.Context)              { notImplemented(c) }
func DeleteContact(c *gin.Context)               { notImplemented(c) }
func ApplyContact(c *gin.Context)                { notImplemented(c) }
func GetNewContactList(c *gin.Context)           { notImplemented(c) }
func PassContactApply(c *gin.Context)            { notImplemented(c) }
func BlackContact(c *gin.Context)                { notImplemented(c) }
func CancelBlackContact(c *gin.Context)          { notImplemented(c) }
func GetAddGroupList(c *gin.Context)             { notImplemented(c) }
func RefuseContactApply(c *gin.Context)          { notImplemented(c) }
func BlackApply(c *gin.Context)                  { notImplemented(c) }
func GetMessageList(c *gin.Context)              { notImplemented(c) }
func GetGroupMessageList(c *gin.Context)         { notImplemented(c) }
func UploadAvatar(c *gin.Context)                { notImplemented(c) }
func UploadFile(c *gin.Context)                  { notImplemented(c) }
func GetCurContactListInChatRoom(c *gin.Context) { notImplemented(c) }
func WsLogin(c *gin.Context)                     { notImplemented(c) }

package dao

import (
	"errors"
	"mychat_server/internal/model"
	"mychat_server/pkg/constants"
	"mychat_server/pkg/utils/zlog"

	"gorm.io/gorm"
)

type UserInfoDao struct{}

func (u *UserInfoDao) GetUserInfoByTelephone(telephone string) (message string, user *model.UserInfo, ret int) {
	if DB == nil {
		return "database is not initialized", nil, -2
	}

	var userInfo model.UserInfo
	result := DB.Where("telephone = ?", telephone).First(&userInfo)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return "user not found", nil, -1
		}
		zlog.Error(result.Error.Error())
		return "query user failed", nil, -2
	}

	return "query user success", &userInfo, 0
}

func (u *UserInfoDao) ExistsByTelephone(telephone string) (string, bool, int) {
	var msg string
	if DB == nil {
		msg = "系统错误(DB)"
		zlog.Error(msg)
		return msg, false, -1
	}
	var user model.UserInfo
	result := DB.Model(&model.UserInfo{}).Where("telephone = ?", telephone).First(&user)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		msg = "手机号不存在"
		zlog.Info(msg)
		return msg, false, 0
	}
	msg = "手机号存在"
	zlog.Info(msg)
	return msg, true, 0
}

func (u *UserInfoDao) NewUser(newUser *model.UserInfo) (string, int) {
	res := DB.Create(newUser)
	if res.Error != nil {
		zlog.Error(res.Error.Error())
		return constants.SYSTEM_ERROR, -1
	}
	return "新用户注册成功", 0
}

func (u *UserInfoDao) GetUserInfoByUuid(uuid string) (string, *model.UserInfo, int) {
	if DB == nil {
		msg := "系统错误(DB)"
		zlog.Error(msg)
		return msg, nil, -1
	}

	var user model.UserInfo
	if res := DB.First(&user, "uuid = ?", uuid); res.Error != nil {
		zlog.Error(res.Error.Error())
		return constants.SYSTEM_ERROR, nil, -1
	}

	return "查询用户信息成功", &user, 0
}

func (u *UserInfoDao) SaveUserInfo(user *model.UserInfo) (string, int) {
	if DB == nil {
		msg := "系统错误(DB)"
		zlog.Error(msg)
		return msg, -1
	}

	res := DB.Save(user)
	if res.Error != nil {
		zlog.Error(res.Error.Error())
		return constants.SYSTEM_ERROR, -1
	}

	return "修改用户信息成功", 0
}

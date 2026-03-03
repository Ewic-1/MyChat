package dao

import (
	"errors"
	"mychat_server/internal/model"
	"mychat_server/pkg/utils/zlog"

	"gorm.io/gorm"
)

type UserInfoDao struct{}

var DB *gorm.DB

func SetDB(db *gorm.DB) {
	DB = db
}

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

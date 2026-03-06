package dao

import (
	"mychat_server/internal/model"
	"mychat_server/pkg/constants"
	"mychat_server/pkg/utils/zlog"
)

type ContactInfoDao struct{}

func (*ContactInfoDao) CreateContact(contact model.UserContact) (string, int) {
	if res := DB.Create(&contact); res.Error != nil {
		zlog.Error(res.Error.Error())
		return constants.SYSTEM_ERROR, -1
	}
	return "创建成功", 0
}

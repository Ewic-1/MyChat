package dao

import (
	"mychat_server/internal/model"
	"mychat_server/pkg/constants"
	"mychat_server/pkg/utils/zlog"
)

type ContactApplyDao struct{}

func (d *ContactApplyDao) GetContactApplyById(contactId string, userId string) (string, model.ContactApply, int) {
	var c model.ContactApply
	res := DB.Where("contact_id = ? and user_id = ?", contactId, userId).First(&c)
	if res.Error != nil {
		zlog.Error(res.Error.Error())
		return constants.SYSTEM_ERROR, c, -1
	}
	return "获取成功", c, 0
}

func (d *ContactApplyDao) SaveContactApply(contactApply model.ContactApply) (string, int) {
	res := DB.Save(&contactApply)
	if res.Error != nil {
		zlog.Error(res.Error.Error())
		return constants.SYSTEM_ERROR, -1
	}
	return "保存成功", 0
}

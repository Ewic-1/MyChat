package dao

import (
	"errors"
	"mychat_server/internal/model"
	"mychat_server/pkg/constants"
	"mychat_server/pkg/utils/zlog"

	"gorm.io/gorm"
)

type ContactApplyDao struct{}

func (d *ContactApplyDao) GetContactApplyById(contactId string, userId string) (string, model.ContactApply, int) {
	var c model.ContactApply
	res := DB.Where("contact_id = ? and user_id = ?", contactId, userId).First(&c)
	if errors.Is(res.Error, gorm.ErrRecordNotFound) {
		return "申请不存在", c, -2
	}
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

func (d *ContactApplyDao) GetContactApplyByContactId(id string) (string, []model.ContactApply, int) {
	var c []model.ContactApply
	res := DB.Where("contact_id = ?", id).Find(&c)
	if res.Error != nil {
		zlog.Error(res.Error.Error())
		return "暂时没有申请", c, -1
	}
	return "获取成功", c, 0
}

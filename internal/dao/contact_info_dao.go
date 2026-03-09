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

func (d *ContactInfoDao) GetContactById(userId string, contactId string) (string, model.UserContact, int) {
	var c model.UserContact
	res := DB.Where("user_id = ? AND contact_id = ?", userId, contactId).First(&c)
	if res.Error != nil {
		zlog.Error(res.Error.Error())
		return constants.SYSTEM_ERROR, c, -1
	}
	return "获取成功", c, 0
}

func (d *ContactInfoDao) SaveContact(contact model.UserContact) (string, int) {

	res := DB.Save(&contact)
	if res.Error != nil {
		zlog.Error(res.Error.Error())
		return constants.SYSTEM_ERROR, -1
	}
	return "保存成功", 0
}

func (d *ContactInfoDao) GetContactByContactId(id string) (string, []model.UserContact, int) {
	var contacts []model.UserContact
	res := DB.Where("contact_id = ?", id).Find(&contacts)
	if res.Error != nil {
		zlog.Error(res.Error.Error())
		return constants.SYSTEM_ERROR, nil, -1
	}
	return "获取成功", contacts, 0
}

func (d *ContactInfoDao) GetContactByUserId(userId string) (string, []model.UserContact, int) {
	var contacts []model.UserContact
	res := DB.Where("user_id = ?", userId).Find(&contacts)
	if res.Error != nil {
		zlog.Error(res.Error.Error())
		return constants.SYSTEM_ERROR, nil, -1
	}
	return "获取成功", contacts, 0
}

package dao

import (
	"mychat_server/internal/model"
	"mychat_server/pkg/constants"
	"mychat_server/pkg/utils/zlog"
)

type MessageDao struct{}

func (d *MessageDao) GetMessageList(id1 string, id2 string) (string, []model.Message, int) {
	var messageList []model.Message
	tx := DB.Where("(send_id = ? and receive_id = ?) or (send_id = ? and receive_id = ?)", id1, id2, id2, id1).
		Order("created_at ASC").
		Find(&messageList)
	if tx.Error != nil {
		zlog.Error(tx.Error.Error())
		return constants.SYSTEM_ERROR, nil, -1
	}
	return "获取成功", messageList, 0
}

func (d *MessageDao) GetGroupMessageList(groupId string) (string, []model.Message, int) {
	var messageList []model.Message
	tx := DB.Where("receive_id = ?", groupId).
		Order("created_at ASC").
		Find(&messageList)
	if tx.Error != nil {
		zlog.Error(tx.Error.Error())
		return constants.SYSTEM_ERROR, nil, -1
	}
	return "获取成功", messageList, 0
}

func (d *MessageDao) SaveMessage(message model.Message) (string, int) {
	tx := DB.Create(&message)
	if tx.Error != nil {
		zlog.Error(tx.Error.Error())
		return constants.SYSTEM_ERROR, -1
	}
	return "保存成功", 0
}

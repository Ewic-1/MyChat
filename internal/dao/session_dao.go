package dao

import (
	"mychat_server/internal/model"
	"mychat_server/pkg/constants"
	"mychat_server/pkg/utils/zlog"
)

type SessionDao struct{}

func (d *SessionDao) GetSessionById(sendId string, receiveId string) (string, model.Session, int) {
	var session model.Session
	res := DB.Where("send_id=? and receive_id=?", sendId, receiveId).First(&session)
	if res.Error != nil {
		zlog.Error(res.Error.Error())
		return constants.SYSTEM_ERROR, session, -1
	}
	return "获取成功", session, 0
}

func (d *SessionDao) SaveSession(session model.Session) (string, int) {
	res := DB.Save(&session)
	if res.Error != nil {
		zlog.Error(res.Error.Error())
		return constants.SYSTEM_ERROR, -1
	}
	return "保存成功", 0
}

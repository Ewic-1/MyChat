package dao

import (
	"errors"
	"mychat_server/internal/model"
	"mychat_server/pkg/constants"
	"mychat_server/pkg/utils/zlog"

	"gorm.io/gorm"
)

type SessionDao struct{}

func (d *SessionDao) GetSessionById(sendId string, receiveId string) (string, model.Session, int) {
	var session model.Session
	res := DB.Where("send_id=? and receive_id=?", sendId, receiveId).First(&session)
	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			zlog.Info("没有找到该会话")
			return "", session, -2
		}
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

func (d *SessionDao) GetSessionByReceiveId(id string) (string, []model.Session, int) {
	var sessions []model.Session
	res := DB.Where("receive_id=?", id).Find(&sessions)
	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			zlog.Info("未创建会话")
			return "", sessions, -2
		}
		zlog.Error(res.Error.Error())
		return constants.SYSTEM_ERROR, nil, -1
	}
	return "获取成功", sessions, 0
}

func (d *SessionDao) GetSessionBySendId(sendId string) (string, []model.Session, int) {
	var sessions []model.Session
	res := DB.Where("send_id=?", sendId).Find(&sessions)
	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			zlog.Info("未创建会话")
			return "", sessions, -2
		}
		zlog.Error(res.Error.Error())
		return constants.SYSTEM_ERROR, sessions, -1
	}
	return "获取成功", sessions, 0
}

func (d *SessionDao) GetSessionBySessionId(id string) (string, model.Session, int) {
	var session model.Session
	res := DB.Where("uuid=?", id).First(&session)
	if res.Error != nil {
		zlog.Error(res.Error.Error())
		return constants.SYSTEM_ERROR, session, -1
	}
	return "获取成功", session, 0
}

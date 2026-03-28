package gorm

import (
	"mychat_server/internal/dao"
	"mychat_server/internal/dto/respond"
	"mychat_server/pkg/utils/zlog"
)

type MessageService struct{}

var messageDao dao.MessageDao

func (s *MessageService) GetMessageList(id1 string, id2 string) (string, []respond.GetMessageListRespond, int) {
	msg, messageList, ret := messageDao.GetMessageList(id1, id2)
	if ret != 0 {
		zlog.Error(msg)
		return msg, nil, ret
	}
	var rspList []respond.GetMessageListRespond
	for _, message := range messageList {
		rspList = append(rspList, respond.GetMessageListRespond{
			SendId:     message.SendId,
			SendName:   message.SendName,
			SendAvatar: message.SendAvatar,
			ReceiveId:  message.ReceiveId,
			Content:    message.Content,
			Url:        message.URL,
			Type:       message.Type,
			FileType:   message.FileType,
			FileName:   message.FileName,
			FileSize:   message.FileSize,
			CreatedAt:  message.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}
	return "获取聊天记录成功", rspList, 0
}

func (s *MessageService) GetGroupMessageList(groupId string) (string, []respond.GetGroupMessageListRespond, int) {
	msg, messageList, ret := messageDao.GetGroupMessageList(groupId)
	if ret != 0 {
		zlog.Error(msg)
		return msg, nil, ret
	}
	var rspList []respond.GetGroupMessageListRespond
	for _, message := range messageList {
		rsp := respond.GetGroupMessageListRespond{
			SendId:     message.SendId,
			SendName:   message.SendName,
			SendAvatar: message.SendAvatar,
			ReceiveId:  message.ReceiveId,
			Content:    message.Content,
			Url:        message.URL,
			Type:       message.Type,
			FileType:   message.FileType,
			FileName:   message.FileName,
			FileSize:   message.FileSize,
			CreatedAt:  message.CreatedAt.Format("2006-01-02 15:04:05"),
		}
		rspList = append(rspList, rsp)
	}
	return "获取聊天记录成功", rspList, 0
}

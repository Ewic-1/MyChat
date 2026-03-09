package gorm

import (
	"mychat_server/internal/dto/respond"
	"mychat_server/pkg/enum/contact/contact_type_enum"
	"mychat_server/pkg/utils/zlog"
)

type ContactService struct{}

func (s *ContactService) GetUserList(id string) (string, []respond.MyUserListRespond, int) {
	msg, contactList, ret := contactInfoDao.GetContactByUserId(id)
	if ret != 0 {
		if contactList == nil {
			msg = "目前没有联系人"
			zlog.Info(msg)
			return msg, nil, 0
		}
		zlog.Error(msg)
		return "", nil, ret
	}

	rep := make([]respond.MyUserListRespond, 0, len(contactList))

	for _, v := range contactList {
		if v.Status == 4 {
			msg = "联系人已删除" + v.ContactId
			zlog.Info(msg)
			continue
		}
		if v.ContactType != contact_type_enum.USER {
			continue
		}
		msg, user, ret := userInfoDao.GetUserInfoByUuid(v.ContactId)
		if ret != 0 {
			zlog.Error(msg)
			return "", nil, ret
		}
		var r respond.MyUserListRespond
		r.UserId = user.Uuid
		r.Avatar = user.Avatar
		r.UserName = user.Nickname
		rep = append(rep, r)

	}
	return "获取用户列表成功", rep, 0
}

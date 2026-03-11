package gorm

import (
	"encoding/json"
	"mychat_server/internal/dto/respond"
	"mychat_server/pkg/constants"
	"mychat_server/pkg/enum/contact/contact_type_enum"
	"mychat_server/pkg/enum/group_info/group_status_enum"
	"mychat_server/pkg/enum/user_info/user_status_enum"
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

func (s *ContactService) LoadMyJoinedGroup(id string) (string, []respond.LoadMyJoinedGroupRespond, int) {
	msg, contactList, ret := contactInfoDao.GetContactByUserId(id)
	if ret != 0 {
		if contactList == nil {
			return "用户没有加入群聊", nil, 0
		}
		zlog.Error(msg)
		return msg, nil, ret
	}
	rep := make([]respond.LoadMyJoinedGroupRespond, 0, len(contactList))
	for _, v := range contactList {
		if v.Status == 6 || v.Status == 7 {
			t, err := json.Marshal(v)
			if err != nil {
				zlog.Error(err.Error())
				return constants.SYSTEM_ERROR, nil, -1
			}
			msg = string(t) + "已退出或被提出的群"
			zlog.Info(msg)
			continue
		}
		if v.ContactType != contact_type_enum.GROUP {
			continue
		}
		msg, group, ret := groupInfoDao.GetGroupInfoById(v.ContactId)
		if ret != 0 {
			zlog.Error(msg)
			return msg, nil, ret
		}
		rep = append(rep, respond.LoadMyJoinedGroupRespond{
			GroupId:   group.Uuid,
			GroupName: group.Name,
			Avatar:    group.Avatar,
		})
	}
	return "获取加入群列表成功", rep, 0
}

func (s *ContactService) GetContactInfo(contactId string) (string, respond.GetContactInfoRespond, int) {
	if contactId[0] == 'G' {
		msg, group, ret := groupInfoDao.GetGroupInfoById(contactId)
		if ret != 0 {
			zlog.Error(msg)
			return msg, respond.GetContactInfoRespond{}, -1
		}
		if group.Status == group_status_enum.DISABLE {
			zlog.Info("群聊已被禁用")
			return "群聊已被禁用", respond.GetContactInfoRespond{}, -2
		}
		return "获取联系人信息成功", respond.GetContactInfoRespond{
			ContactId:        group.Uuid,
			ContactName:      group.Name,
			ContactAvatar:    group.Avatar,
			ContactNotice:    group.Notice,
			ContactAddMode:   group.AddMode,
			ContactMembers:   group.Members,
			ContactMemberCnt: group.MemberCnt,
			ContactOwnerId:   group.OwnerId,
		}, 0
	} else {
		msg, user, ret := userInfoDao.GetUserInfoByUuid(contactId)
		if ret != 0 {
			zlog.Error(msg)
			return msg, respond.GetContactInfoRespond{}, -1
		}
		if user.Status == user_status_enum.DISABLE {
			zlog.Info("用户已被禁用")
			return "用户已被禁用", respond.GetContactInfoRespond{}, -2
		}
		return "获取联系人信息成功", respond.GetContactInfoRespond{
			ContactId:        user.Uuid,
			ContactName:      user.Nickname,
			ContactAvatar:    user.Avatar,
			ContactBirthday:  user.Birthday,
			ContactEmail:     user.Email,
			ContactPhone:     user.Telephone,
			ContactGender:    user.Gender,
			ContactSignature: user.Signature,
		}, 0
	}
}

package gorm

import (
	"encoding/json"
	"fmt"
	"mychat_server/internal/dto/request"
	"mychat_server/internal/dto/respond"
	"mychat_server/internal/model"
	"mychat_server/pkg/constants"
	"mychat_server/pkg/enum/contact/contact_status_enum"
	"mychat_server/pkg/enum/contact/contact_type_enum"
	"mychat_server/pkg/enum/contact_apply/contact_apply_status_enum"
	"mychat_server/pkg/enum/group_info/group_status_enum"
	"mychat_server/pkg/enum/user_info/user_status_enum"
	"mychat_server/pkg/utils/random"
	"mychat_server/pkg/utils/zlog"
	"time"

	"gorm.io/gorm"
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

func (s *ContactService) DeleteContact(id1 string, id2 string) (string, int) {
	// deletedAt
	deletedAt := gorm.DeletedAt{
		Time:  time.Now(),
		Valid: true,
	}
	// contact
	msg, contact, ret := contactInfoDao.GetContactById(id1, id2)
	if ret != 0 {
		zlog.Error(msg)
		return msg, ret
	}
	contact.Status = contact_status_enum.DELETE
	contact.DeletedAt = deletedAt
	contactInfoDao.SaveContact(contact)

	msg, contact, ret = contactInfoDao.GetContactById(id2, id1)
	if ret != 0 {
		zlog.Error(msg)
		return msg, ret
	}
	contact.Status = contact_status_enum.DELETE
	contact.DeletedAt = deletedAt
	contactInfoDao.SaveContact(contact)
	// session
	msg, session, ret := sessionDao.GetSessionById(id1, id2)
	if ret != 0 {
		zlog.Error(msg)
		return msg, ret
	}
	session.DeletedAt = deletedAt
	sessionDao.SaveSession(session)

	msg, session, ret = sessionDao.GetSessionById(id2, id1)
	if ret != 0 {
		zlog.Error(msg)
		return msg, ret
	}
	session.DeletedAt = deletedAt
	sessionDao.SaveSession(session)
	// apply
	msg, apply, ret := contactApplyDao.GetContactApplyById(id1, id2)
	if ret != 0 {
		zlog.Error(msg)
		return msg, ret
	}
	apply.DeletedAt = deletedAt
	contactApplyDao.SaveContactApply(apply)

	msg, apply, ret = contactApplyDao.GetContactApplyById(id2, id1)
	if ret != 0 {
		zlog.Error(msg)
		return msg, ret
	}
	apply.DeletedAt = deletedAt
	contactApplyDao.SaveContactApply(apply)

	return "删除联系人成功", 0
}

func (s *ContactService) ApplyContact(req request.ApplyContactRequest) (string, int) {
	// 申请用户
	if req.ContactId[0] == 'U' {
		msg, user, ret := userInfoDao.GetUserInfoByUuid(req.ContactId)
		if ret != 0 {
			zlog.Error(msg)
			return msg, ret
		}
		if user.Status == user_status_enum.DISABLE {
			msg = "用户已被禁用"
			zlog.Error(msg)
			return msg, -2
		}

		msg, apply, ret := contactApplyDao.GetContactApplyById(req.OwnerId, req.ContactId)
		if ret == -1 {
			zlog.Error(msg)
			return msg, ret
		} else if ret == -2 { // 申请不存在
			apply = model.ContactApply{
				Uuid:        fmt.Sprintf("A%s", random.GetNowAndLenRandomString(11)),
				UserId:      req.OwnerId,
				ContactId:   req.ContactId,
				ContactType: contact_type_enum.USER,
				Status:      contact_apply_status_enum.PENDING,
				Message:     req.Message,
				LastApplyAt: time.Now(),
			}
			msg, ret = contactApplyDao.SaveContactApply(apply)
			if ret != 0 {
				zlog.Error(msg)
				return msg, ret
			}
		}
		if apply.Status == contact_apply_status_enum.BLACK {
			msg = "对方已将你拉黑"
			zlog.Error(msg)
			return msg, -2
		}
	} else if req.ContactId[0] == 'G' {
		// 申请群聊
		msg, group, ret := groupInfoDao.GetGroupInfoById(req.ContactId)
		if ret != 0 {
			zlog.Error(msg)
			return msg, ret
		}
		if group.Status == group_status_enum.DISABLE {
			msg = "群聊已被禁用"
			zlog.Error(msg)
			return msg, -2
		}
		msg, apply, ret := contactApplyDao.GetContactApplyById(req.OwnerId, req.ContactId)
		if ret == -1 {
			zlog.Error(msg)
			return msg, ret
		} else if ret == -2 { // 申请不存在
			apply = model.ContactApply{
				Uuid:        fmt.Sprintf("A%s", random.GetNowAndLenRandomString(11)),
				UserId:      req.OwnerId,
				ContactId:   req.ContactId,
				ContactType: contact_type_enum.GROUP,
				Status:      contact_apply_status_enum.PENDING,
				Message:     req.Message,
				LastApplyAt: time.Now(),
			}
			msg, ret = contactApplyDao.SaveContactApply(apply)
			if ret != 0 {
				zlog.Error(msg)
				return msg, ret
			}
		}
	}
	return "申请发送成功", 0
}

func (s *ContactService) GetNewContactList(ownerId string) (string, []respond.NewContactListRespond, int) {
	msg, contactApplyList, ret := contactApplyDao.GetContactApplyByContactId(ownerId)
	if ret != 0 {
		zlog.Error(msg)
		return msg, nil, 0
	}
	var rep []respond.NewContactListRespond
	for _, contactApply := range contactApplyList {
		if contactApply.Status == contact_apply_status_enum.PENDING {
			var applyMessage string
			if contactApply.Message == "" {
				applyMessage = "申请理由：无"
			} else {
				applyMessage = "申请理由：" + contactApply.Message
			}
			msg, user, ret := userInfoDao.GetUserInfoByUuid(contactApply.UserId)
			if ret != 0 {
				zlog.Error(msg)
				return msg, nil, -1
			}
			var r respond.NewContactListRespond = respond.NewContactListRespond{
				ContactId:     contactApply.Uuid,
				Message:       applyMessage,
				ContactName:   user.Nickname,
				ContactAvatar: user.Avatar,
			}
			rep = append(rep, r)
		}
	}
	return "获取成功", rep, 0
}

func (s *ContactService) PassContactApply(ownerId string, contactId string) (string, int) {
	msg, contactApply, ret := contactApplyDao.GetContactApplyById(ownerId, contactId)
	if ret != 0 {
		zlog.Error(msg)
		return msg, -2
	}
	if ownerId[0] == 'U' {
		msg, user, ret := userInfoDao.GetUserInfoByUuid(contactId)
		if ret != 0 {
			zlog.Error(msg)
			return msg, -1
		}
		if user.Status == user_status_enum.DISABLE {
			zlog.Info("用户被禁用")
			return "用户被禁用", 0
		}
		contactApply.Status = contact_apply_status_enum.AGREE
		contactApplyDao.SaveContactApply(contactApply)

		contact1 := model.UserContact{
			ContactId:   contactId,
			UserId:      ownerId,
			ContactType: contact_type_enum.USER,
			Status:      contact_status_enum.NORMAL,
			UpdateAt:    time.Now(),
			CreatedAt:   time.Now(),
		}
		contact2 := model.UserContact{
			ContactId:   ownerId,
			UserId:      contactId,
			ContactType: contact_type_enum.USER,
			Status:      contact_status_enum.NORMAL,
			UpdateAt:    time.Now(),
			CreatedAt:   time.Now(),
		}
		contactInfoDao.SaveContact(contact1)
		contactInfoDao.SaveContact(contact2)
		return "已添加该联系人", 0
	} else {
		msg, group, ret := groupInfoDao.GetGroupInfoById(ownerId)
		if ret != 0 {
			zlog.Error(msg)
			return msg, -1
		}
		if group.Status == group_status_enum.DISABLE {
			zlog.Info("群聊已被禁用")
			return "群聊已被禁用", 0
		}
		contact := model.UserContact{
			ContactId:   contactId,
			UserId:      ownerId,
			ContactType: contact_type_enum.GROUP,
			Status:      contact_status_enum.NORMAL,
			UpdateAt:    time.Now(),
			CreatedAt:   time.Now(),
		}
		contactInfoDao.SaveContact(contact)
		members := []string{}
		err := json.Unmarshal(group.Members, &members)
		if err != nil {
			zlog.Error(err.Error())
			return msg, -1
		}
		members = append(members, contactId)
		group.Members, _ = json.Marshal(members)
		group.MemberCnt++
		groupInfoDao.SaveGroup(group)
		return "已通过加群申请", 0
	}
}

func (s *ContactService) BlackContact(ownerId string, contactId string) (string, int) {
	// 拉黑
	msg, contact1, ret := contactInfoDao.GetContactById(ownerId, contactId)
	if ret != 0 {
		zlog.Error(msg)
		return msg, -1
	}
	contact1.Status = contact_status_enum.BLACK
	contact1.UpdateAt = time.Now()
	contactInfoDao.SaveContact(contact1)
	// 被拉黑
	msg, contact2, ret := contactInfoDao.GetContactById(contactId, ownerId)
	if ret != 0 {
		zlog.Error(msg)
		return msg, -1
	}
	contact2.Status = contact_status_enum.BE_BLACK
	contact2.UpdateAt = time.Now()
	contactInfoDao.SaveContact(contact2)
	// 删除会话
	delatedAt := gorm.DeletedAt{
		Time:  time.Now(),
		Valid: true,
	}
	msg, session, ret := sessionDao.GetSessionById(ownerId, contactId)
	if ret != 0 {
		zlog.Error(msg)
		return msg, -1
	}
	session.DeletedAt = delatedAt
	sessionDao.SaveSession(session)
	return "已拉黑联系人", 0
}

func (s *ContactService) CancelBlackContact(ownerId string, contactId string) (string, int) {
	// 因为前端的设定，这里需要判断一下ownerId和contactId是不是有拉黑和被拉黑的状态
	msg, blackContact, ret := contactInfoDao.GetContactById(ownerId, contactId)
	if ret != 0 {
		zlog.Error(msg)
		return msg, -1
	}
	if blackContact.Status != contact_status_enum.BLACK {
		return "未拉黑该联系人，无需解除拉黑", -2
	}
	msg, beBlackContact, ret := contactInfoDao.GetContactById(contactId, ownerId)
	if ret != 0 {
		zlog.Error(msg)
		return msg, -1
	}
	if beBlackContact.Status != contact_status_enum.BE_BLACK {
		return "该联系人未被拉黑，无需解除拉黑", -2
	}

	// 取消拉黑
	blackContact.Status = contact_status_enum.NORMAL
	beBlackContact.Status = contact_status_enum.NORMAL
	contactInfoDao.SaveContact(blackContact)
	contactInfoDao.SaveContact(beBlackContact)

	return "已解除拉黑该联系人", 0
}

func (s *ContactService) GetAddGroupList(groupId string) (string, []respond.AddGroupListRespond, int) {
	msg, contactApplyList, ret := contactApplyDao.GetContactApplyByContactId(groupId)
	if ret != 0 {
		zlog.Error(msg)
		return msg, nil, ret
	}

	var rsp []respond.AddGroupListRespond
	for _, contactApply := range contactApplyList {
		if contactApply.Status != contact_apply_status_enum.PENDING {
			continue
		}
		var message string
		if contactApply.Message == "" {
			message = "申请理由：无"
		} else {
			message = "申请理由：" + contactApply.Message
		}
		newContact := respond.AddGroupListRespond{
			ContactId: contactApply.Uuid,
			Message:   message,
		}
		msg, user, ret := userInfoDao.GetUserInfoByUuid(contactApply.UserId)
		if ret != 0 {
			zlog.Error(msg)
			return msg, nil, ret
		}
		newContact.ContactId = user.Uuid
		newContact.ContactName = user.Nickname
		newContact.ContactAvatar = user.Avatar
		rsp = append(rsp, newContact)
	}
	return "获取成功", rsp, 0
}

func (s *ContactService) RefuseContactApply(ownerId string, contactId string) (string, int) {
	// ownerId 如果是用户的话就是登录用户，如果是群聊的话就是群聊id
	msg, contactApply, ret := contactApplyDao.GetContactApplyById(ownerId, contactId)
	if ret != 0 {
		zlog.Error(msg)
		return msg, -1
	}
	contactApply.Status = contact_apply_status_enum.REFUSE
	contactApplyDao.SaveContactApply(contactApply)
	if ownerId[0] == 'U' {
		return "已拒绝该联系人申请", 0
	} else {
		return "已拒绝该加群申请", 0
	}
}

func (s *ContactService) BlackApply(ownerId string, contactId string) (string, int) {
	msg, contactApply, ret := contactApplyDao.GetContactApplyById(ownerId, contactId)
	if ret != 0 {
		zlog.Error(msg)
		return msg, -1
	}

	contactApply.Status = contact_apply_status_enum.BLACK
	contactApplyDao.SaveContactApply(contactApply)
	return "已拉黑该申请", 0
}

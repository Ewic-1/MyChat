package gorm

import (
	"encoding/json"
	"fmt"
	"mychat_server/internal/dao"
	"mychat_server/internal/dto/request"
	"mychat_server/internal/dto/respond"
	"mychat_server/internal/model"
	"mychat_server/pkg/constants"
	"mychat_server/pkg/enum/contact/contact_status_enum"
	"mychat_server/pkg/enum/contact/contact_type_enum"
	"mychat_server/pkg/enum/group_info/group_status_enum"
	"mychat_server/pkg/utils/random"
	"mychat_server/pkg/utils/zlog"
	"time"

	"gorm.io/gorm"
)

type GroupInfoService struct{}

var groupInfoDao = new(dao.GroupInfoDao)
var contactInfoDao = new(dao.ContactInfoDao)
var sessionDao = new(dao.SessionDao)
var contactApplyDao = new(dao.ContactApplyDao)

func (s GroupInfoService) CreateGroup(req request.CreateGroupRequest) (msg string, ret int) {
	group := model.GroupInfo{
		Uuid:      fmt.Sprintf("G%s", random.GetNowAndLenRandomString(11)),
		Name:      req.Name,
		Notice:    req.Notice,
		OwnerId:   req.OwnerId,
		MemberCnt: 1,
		AddMode:   req.AddMode,
		Avatar:    req.Avatar,
		Status:    group_status_enum.NORMAL,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	// 群成员（添加群主一人）
	var members []string
	members = append(members, group.OwnerId)
	var err error
	group.Members, err = json.Marshal(members)
	if err != nil {
		zlog.Error(err.Error())
		return constants.SYSTEM_ERROR, -1
	}

	// 存
	msg, ret = groupInfoDao.CreateGroup(group)
	if ret != 0 {
		zlog.Error(msg)
		return msg, ret
	}

	// 创建联系人
	contact := model.UserContact{
		UserId:      req.OwnerId,
		ContactId:   group.Uuid,
		ContactType: contact_type_enum.GROUP,
		Status:      contact_status_enum.NORMAL,
		CreatedAt:   time.Now(),
		UpdateAt:    time.Now(),
	}
	// 存
	msg, ret = contactInfoDao.CreateContact(contact)
	if ret != 0 {
		zlog.Error(msg)
		return msg, ret
	}
	msg = "创建成功"
	ret = 0
	return
}

func (s *GroupInfoService) LoadMyGroup(ownerId string) (msg string, groupList []respond.LoadMyGroupRespond, ret int) {
	var res []model.GroupInfo
	msg, res, ret = groupInfoDao.GetGroupInfoByOwnerId(ownerId)
	if ret != 0 {
		zlog.Error(msg)
		return msg, nil, ret
	}
	for _, v := range res {
		groupList = append(groupList, respond.LoadMyGroupRespond{
			GroupId:   v.Uuid,
			GroupName: v.Name,
			Avatar:    v.Avatar,
		})
	}
	return msg, groupList, ret
}

func (s *GroupInfoService) CheckGroupAddMode(id string) (msg string, addMode int8, ret int) {
	msg, addMode, ret = groupInfoDao.CheckGroupAddMode(id)
	if ret != 0 {
		zlog.Error(msg)
		return msg, -1, ret
	}
	return
}

func (s *GroupInfoService) EnterGroupDirectly(req request.EnterGroupDirectlyRequest) (string, int) {
	// 根据id获取群
	uuid := req.OwnerId
	userId := req.ContactId
	var group model.GroupInfo
	msg, group, ret := groupInfoDao.GetGroupInfoById(uuid)
	if ret != 0 {
		zlog.Error(msg)
		return msg, ret
	}
	// 添加用户到群成员列表
	var members []string
	if err := json.Unmarshal(group.Members, &members); err != nil {
		zlog.Error(err.Error())
		return constants.SYSTEM_ERROR, -1
	}
	members = append(members, userId)
	data, err := json.Marshal(members)
	if err != nil {
		zlog.Error(err.Error())
		return constants.SYSTEM_ERROR, -1
	}
	group.Members = data

	// 群人数+1
	group.MemberCnt++

	// 保存群信息
	msg, ret = groupInfoDao.SaveGroup(group)
	if ret != 0 {
		zlog.Error(msg)
		return constants.SYSTEM_ERROR, ret
	}

	// 联系人列表中添加对应信息
	newContact := model.UserContact{}
	newContact.UserId = userId
	newContact.ContactId = uuid
	newContact.ContactType = contact_type_enum.GROUP
	newContact.Status = contact_status_enum.NORMAL
	newContact.CreatedAt = time.Now()
	newContact.UpdateAt = time.Now()
	msg, ret = contactInfoDao.CreateContact(newContact)
	if ret != 0 {
		zlog.Error(msg)
		return msg, ret
	}

	// 返回
	return "加入成功", 0
}

func (s *GroupInfoService) LeaveGroup(userId string, groupId string) (string, int) {
	// 从群聊删除用户
	msg, group, ret := groupInfoDao.GetGroupInfoById(groupId)
	if ret != 0 {
		zlog.Error(msg)
		return msg, -1
	}
	var members []string
	err := json.Unmarshal(group.Members, &members)
	if err != nil {
		zlog.Error(err.Error())
		return constants.SYSTEM_ERROR, -1
	}
	for i, member := range members {
		if member == userId {
			members = append(members[:i], members[i+1:]...)
			break
		}
	}
	group.Members, err = json.Marshal(members)
	if err != nil {
		zlog.Error(err.Error())
		return constants.SYSTEM_ERROR, -1
	}
	group.MemberCnt--

	// delateAt变量方便反复使用
	delatedAt := gorm.DeletedAt{
		Time:  time.Now(),
		Valid: true,
	}

	// 删除联系人
	msg, ctt, ret := contactInfoDao.GetContactById(userId, groupId)
	if ret != 0 {
		zlog.Error(msg)
		return msg, ret
	}
	ctt.DeletedAt = delatedAt
	ctt.Status = contact_status_enum.QUIT_GROUP
	msg, ret = contactInfoDao.SaveContact(ctt)
	if ret != 0 {
		zlog.Error(msg)
		return msg, ret
	}

	// 删除会话
	msg, session, ret := sessionDao.GetSessionById(userId, groupId)
	if ret != 0 {
		zlog.Error(msg)
		return msg, ret
	}
	session.DeletedAt = delatedAt
	msg, ret = sessionDao.SaveSession(session)
	if ret != 0 {
		zlog.Error(msg)
		return msg, ret
	}

	// 删除申请记录
	msg, contactApply, ret := contactApplyDao.GetContactApplyById(groupId, userId)
	if ret != 0 {
		zlog.Error(msg)
		return msg, ret
	}
	contactApply.DeletedAt = delatedAt
	msg, ret = contactApplyDao.SaveContactApply(contactApply)
	if ret != 0 {
		zlog.Error(msg)
		return msg, ret
	}
	// 返回
	return "退出群聊成功", 0
}

func (s *GroupInfoService) DismissGroup(groupId string) (string, int) {
	// 从groupinfo表中删除
	delatedAt := gorm.DeletedAt{
		Time:  time.Now(),
		Valid: true,
	}
	msg, group, ret := groupInfoDao.GetGroupInfoById(groupId)
	if ret != 0 {
		zlog.Error(msg)
		return msg, -1
	}
	group.DeletedAt = delatedAt
	msg, ret = groupInfoDao.SaveGroup(group)
	if ret != 0 {
		zlog.Error(msg)
		return msg, -1
	}
	// 从session列表中删除
	msg, sessions, ret := sessionDao.GetSessionByReceiveId(groupId)
	if ret != 0 {
		zlog.Error(msg)
		return msg, ret
	}
	for _, session := range sessions {
		session.DeletedAt = delatedAt
		msg, ret = sessionDao.SaveSession(session)
		if ret != 0 {
			zlog.Error(msg)
			return msg, ret
		}
	}
	// 从联系人列表中删除
	msg, contacts, ret := contactInfoDao.GetContactByContactId(groupId)
	if ret != 0 {
		zlog.Error(msg)
		return msg, ret
	}
	for _, contact := range contacts {
		contact.DeletedAt = delatedAt
		msg, ret = contactInfoDao.SaveContact(contact)
		if ret != 0 {
			zlog.Error(msg)
			return msg, ret
		}
	}
	// 删除所有申请该群聊的申请记录
	msg, contactApplies, ret := contactApplyDao.GetContactApplyByContactId(groupId)
	for _, contactApply := range contactApplies {
		contactApply.DeletedAt = delatedAt
		msg, ret = contactApplyDao.SaveContactApply(contactApply)
		if ret != 0 {
			zlog.Error(msg)
			return msg, ret
		}
	}
	// 返回
	return "解散群聊成功", 0
}

// GetGroupInfo 获取群聊详情
func (g *GroupInfoService) GetGroupInfo(groupId string) (string, *respond.GetGroupInfoRespond, int) {
	msg, group, ret := groupInfoDao.GetGroupInfoById(groupId)
	if ret != 0 {
		zlog.Error(msg)
		return msg, nil, ret
	}
	rsp := &respond.GetGroupInfoRespond{
		Uuid:      group.Uuid,
		Name:      group.Name,
		Notice:    group.Notice,
		Avatar:    group.Avatar,
		MemberCnt: group.MemberCnt,
		OwnerId:   group.OwnerId,
		AddMode:   group.AddMode,
		Status:    group.Status,
	}
	if group.DeletedAt.Valid {
		rsp.IsDeleted = true
	} else {
		rsp.IsDeleted = false
	}
	return "获取成功", rsp, 0
}

func (s *GroupInfoService) UpdateGroupInfo(req request.UpdateGroupInfoRequest) (string, int) {
	groupId := req.Uuid
	msg, group, ret := groupInfoDao.GetGroupInfoById(groupId)
	if ret != 0 {
		zlog.Error(msg)
		return msg, -1
	}
	if req.Name != "" {
		group.Name = req.Name
	}
	if req.Notice != "" {
		group.Notice = req.Notice
	}
	if req.Avatar != "" {
		group.Avatar = req.Avatar
	}
	if req.OwnerId != "" {
		group.OwnerId = req.OwnerId
	}
	if req.AddMode != -1 {
		group.AddMode = req.AddMode
	}
	groupInfoDao.SaveGroup(group)
	// 更新会话列表
	msg, sessionList, ret := sessionDao.GetSessionByReceiveId(groupId)
	if ret != 0 {
		zlog.Error(msg)
		return msg, -1
	}
	for _, session := range sessionList {
		session.Avatar = group.Avatar
		session.ReceiveName = group.Name
		msg, ret = sessionDao.SaveSession(session)
		if ret != 0 {
			zlog.Error(msg)
			return msg, -1
		}
	}
	return "更新成功", 0
}

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
)

type GroupInfoService struct{}

var groupInfoDao = new(dao.GroupInfoDao)
var contactInfoDao = new(dao.ContactInfoDao)

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
	members := []string{}
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

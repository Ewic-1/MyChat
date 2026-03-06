package dao

import (
	"mychat_server/internal/model"
	"mychat_server/pkg/constants"
	"mychat_server/pkg/utils/zlog"
)

type GroupInfoDao struct{}

func (*GroupInfoDao) CreateGroup(groupInfo model.GroupInfo) (message string, ret int) {
	if res := DB.Create(&groupInfo); res.Error != nil {
		zlog.Error(res.Error.Error())
		return constants.SYSTEM_ERROR, -1
	}
	return "创建成功", 0
}

func (d *GroupInfoDao) GetGroupInfoByOwnerId(id string) (msg string, groupList []model.GroupInfo, ret int) {
	res := DB.Order("created_at DESC").Where("owner_id = ?", id).Find(&groupList)
	if res.Error != nil {
		zlog.Error(res.Error.Error())
		return constants.SYSTEM_ERROR, nil, -1
	}
	msg = "获取成功"
	ret = 0
	return
}

func (d *GroupInfoDao) CheckGroupAddMode(id string) (string, int8, int) {
	group := model.GroupInfo{}
	res := DB.First(group, "uuid = ?", id)
	if res.Error != nil {
		zlog.Error(res.Error.Error())
		return constants.SYSTEM_ERROR, -1, -1
	}
	return "加群方式获取成功", group.AddMode, 0
}

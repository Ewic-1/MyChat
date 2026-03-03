package model

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

type GroupInfo struct {
	Id        int64           `gorm:"column:id;primaryKey;autoIncrement"`
	Uuid      string          `gorm:"column:uuid;uniqueIndex;type:char(20);not null"`
	Name      string          `gorm:"column:name;type:varchar(20);not null"`
	Notice    string          `gorm:"column:notice;type:varchar(500)"`
	Members   json.RawMessage `gorm:"column:members;type:json"`
	MemberCnt int             `gorm:"column:member_cnt;default:1"`
	OwnerId   string          `gorm:"column:owner_id;type:char(20);not null"`
	AddMode   int8            `gorm:"column:add_mode;default:0"`
	Avatar    string          `gorm:"column:avatar;type:char(255);not null"`
	Status    int8            `gorm:"column:status;default:0"`
	CreatedAt time.Time       `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt time.Time       `gorm:"column:updated_at;autoUpdateTime"`
	DeletedAt gorm.DeletedAt  `gorm:"column:deleted_at;index"`
}

func (GroupInfo) TableName() string {
	return "group_info"
}

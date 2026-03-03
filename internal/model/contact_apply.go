package model

import (
	"time"

	"gorm.io/gorm"
)

type ContactApply struct {
	Id          int64          `gorm:"column:id;primaryKey;autoIncrement"`
	Uuid        string         `gorm:"column:uuid;uniqueIndex;type:char(20)"`
	UserId      string         `gorm:"column:user_id;index;type:char(20);not null"`
	ContactId   string         `gorm:"column:contact_id;index;type:char(20);not null"`
	ContactType int8           `gorm:"column:contact_type;not null"`
	Status      int8           `gorm:"column:status;not null;default:0"`
	Message     string         `gorm:"column:message;type:varchar(100)"`
	LastApplyAt time.Time      `gorm:"column:last_apply_at;autoCreateTime"`
	DeletedAt   gorm.DeletedAt `gorm:"column:deleted_at;index"`
}

func (ContactApply) TableName() string {
	return "contact_apply"
}

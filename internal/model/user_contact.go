package model

import (
	"time"

	"gorm.io/gorm"
)

type UserContact struct {
	Id          int64          `gorm:"column:id;primaryKey;autoIncrement"`
	UserId      string         `gorm:"column:user_id;index;type:char(20);not null"`
	ContactId   string         `gorm:"column:contact_id;index;type:char(20);not null"`
	ContactType int8           `gorm:"column:contact_type;not null"`
	Status      int8           `gorm:"column:status;not null;default:0"`
	CreatedAt   time.Time      `gorm:"column:created_at;autoCreateTime"`
	UpdateAt    time.Time      `gorm:"column:update_at;autoUpdateTime"`
	DeletedAt   gorm.DeletedAt `gorm:"column:deleted_at;index"`
}

func (UserContact) TableName() string {
	return "user_contact"
}

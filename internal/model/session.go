package model

import (
	"database/sql"
	"time"

	"gorm.io/gorm"
)

type Session struct {
	Id            int64          `gorm:"column:id;primaryKey;autoIncrement"`
	Uuid          string         `gorm:"column:uuid;uniqueIndex;type:char(20)"`
	SendId        string         `gorm:"column:send_id;index;type:char(20);not null"`
	ReceiveId     string         `gorm:"column:receive_id;index;type:char(20);not null"`
	ReceiveName   string         `gorm:"column:receive_name;type:varchar(20);not null"`
	Avatar        string         `gorm:"column:avatar;type:char(255);not null"`
	LastMessage   string         `gorm:"column:last_message;type:text"`
	LastMessageAt sql.NullTime   `gorm:"column:last_message_at"`
	CreatedAt     time.Time      `gorm:"column:created_at;autoCreateTime"`
	DeletedAt     gorm.DeletedAt `gorm:"column:deleted_at;index"`
}

func (Session) TableName() string {
	return "session"
}

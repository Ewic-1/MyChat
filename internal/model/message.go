package model

import (
	"database/sql"
	"time"
)

type Message struct {
	Id         int64        `gorm:"column:id;primaryKey;autoIncrement"`
	Uuid       string       `gorm:"column:uuid;uniqueIndex;type:char(20);not null"`
	SessionId  string       `gorm:"column:session_id;index;type:char(20);not null"`
	Type       int8         `gorm:"column:type;not null"`
	Content    string       `gorm:"column:content;type:text"`
	URL        string       `gorm:"column:url;type:char(255)"`
	SendId     string       `gorm:"column:send_id;index;type:char(20);not null"`
	SendName   string       `gorm:"column:send_name;type:varchar(20);not null"`
	SendAvatar string       `gorm:"column:send_avatar;type:varchar(255);not null"`
	ReceiveId  string       `gorm:"column:receive_id;index;type:char(20);not null"`
	FileType   string       `gorm:"column:file_type;type:char(10)"`
	FileName   string       `gorm:"column:file_name;type:varchar(50)"`
	FileSize   string       `gorm:"column:file_size;type:char(20)"`
	Status     int8         `gorm:"column:status;not null;default:0"`
	CreatedAt  time.Time    `gorm:"column:created_at;autoCreateTime"`
	SendAt     sql.NullTime `gorm:"column:send_at"`
	AVData     string       `gorm:"column:av_data;type:text"`
}

func (Message) TableName() string {
	return "message"
}

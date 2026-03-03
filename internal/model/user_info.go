package model

import (
	"database/sql"
	"time"

	"gorm.io/gorm"
)

type UserInfo struct {
	Id            int64          `gorm:"column:id;primaryKey;autoIncrement"`
	Uuid          string         `gorm:"column:uuid;uniqueIndex;type:char(20)"`
	Nickname      string         `gorm:"column:nickname;type:varchar(20);not null"`
	Telephone     string         `gorm:"column:telephone;index;type:char(11);not null"`
	Email         string         `gorm:"column:email;type:char(30)"`
	Avatar        string         `gorm:"column:avatar;type:char(255);not null"`
	Gender        int8           `gorm:"column:gender"`
	Signature     string         `gorm:"column:signature;type:varchar(100)"`
	Password      string         `gorm:"column:password;type:char(72);not null"`
	Birthday      string         `gorm:"column:birthday;type:char(8)"`
	CreatedAt     time.Time      `gorm:"column:created_at;autoCreateTime"`
	DeletedAt     gorm.DeletedAt `gorm:"column:deleted_at;index"`
	LastOnlineAt  sql.NullTime   `gorm:"column:last_online_at"`
	LastOfflineAt sql.NullTime   `gorm:"column:last_offline_at"`
	IsAdmin       int8           `gorm:"column:is_admin;default:0"`
	Status        int8           `gorm:"column:status;index;default:0"`
}

func (UserInfo) TableName() string {
	return "user_info"
}

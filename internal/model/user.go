package model

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID        uint   `gorm:"primarykey"`
	OpenID    string `gorm:"size:64;index"`
	UnionID   string `gorm:"size:64;index"` // 微信unionid
	Username  string `gorm:"size:64"`
	Password  string `gorm:"size:64"`
	Nickname  string `gorm:"size:64"`
	Avatar    string `gorm:"size:255"`
	Sex       int    `gorm:"default:0"` // 0: 未知, 1: 男, 2: 女
	Country   string `gorm:"size:64"`
	Province  string `gorm:"size:64"`
	City      string `gorm:"size:64"`
	IsAdmin   bool   `gorm:"default:false"` // 是否是管理员
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

package model

import (
	"time"

	"gorm.io/gorm"
)

// AdminLoginLog 管理员登录日志
type AdminLoginLog struct {
	ID         uint           `gorm:"primarykey"`
	Username   string         `gorm:"size:64;index"` // 登录账号
	IP         string         `gorm:"size:64"`       // IP地址
	UserAgent  string         `gorm:"size:255"`      // 设备UA
	IsSuccess  bool           `gorm:"default:false"` // 登录是否成功
	FailReason string         `gorm:"size:255"`      // 失败原因，成功为空
	LoginTime  time.Time      // 登录时间
	CreatedAt  time.Time      // 创建时间
	UpdatedAt  time.Time      // 更新时间
	DeletedAt  gorm.DeletedAt `gorm:"index"` // 软删除
}

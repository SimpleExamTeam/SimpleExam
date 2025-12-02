package model

import (
	"time"

	"gorm.io/gorm"
)

type QRCode struct {
	ID        uint   `gorm:"primarykey"`
	SceneStr  string `gorm:"size:64;index"` // 场景值，用于标识二维码
	Status    string `gorm:"size:20"`       // pending, scanned, confirmed, expired
	OpenID    string `gorm:"size:64"`       // 扫码用户的OpenID
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

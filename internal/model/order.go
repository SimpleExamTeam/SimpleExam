package model

import (
	"time"

	"gorm.io/gorm"
)

type Order struct {
	ID          uint   `gorm:"primarykey"`
	OrderNo     string `gorm:"size:64;index"`
	UserID      uint   `gorm:"index"`
	User        User   `gorm:"foreignKey:UserID"`
	CourseID    uint   `gorm:"index"`
	Amount      float64
	Status      string     `gorm:"size:20"` // pending, paid, canceled
	PaymentType string     `gorm:"size:20"` // 支付方式
	PayTime     *time.Time // 使用指针类型，可以为 NULL
	ExpireTime  *time.Time // 订单过期时间，可以为 NULL
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}

// 检查订单是否过期（15分钟未支付）
func (o *Order) IsExpired() bool {
	return o.Status != "paid" && o.Status != "cancelled" && time.Since(o.CreatedAt) > 15*time.Minute
}

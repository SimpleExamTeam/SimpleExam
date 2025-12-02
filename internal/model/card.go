package model

import (
	"time"

	"gorm.io/gorm"
)

// Card 卡券模型
type Card struct {
	ID         uint           `json:"id" gorm:"primarykey"`
	CardNo     string         `json:"card_no" gorm:"size:18;uniqueIndex;comment:卡券唯一编号"` // 卡券的唯一ID，数字+字母混合18位
	CourseID   *uint          `json:"course_id" gorm:"index;comment:绑定的课程ID"`            // 绑定的课程，如果为空可以兑换所有课程
	Amount     float64        `json:"amount" gorm:"comment:卡券金额"`                        // 使用卡券创建订单时的金额，如果为空或者为0，订单金额为0
	Total      int            `json:"total" gorm:"comment:可兑换总数"`                        // 发放的可兑换总数
	ExpireDays int            `json:"expire_days" gorm:"comment:有效期天数"`                  // 卡券兑换的有效期，超过创建时间+expire_days无法兑换
	Used       int            `json:"used" gorm:"default:0;comment:已兑换数量"`               // 已经兑换的数量
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `json:"-" gorm:"index"`
}

// CardRecord 卡券兑换记录
type CardRecord struct {
	ID        uint           `json:"id" gorm:"primarykey"`
	CardID    uint           `json:"card_id" gorm:"index;comment:卡券ID"`
	CardNo    string         `json:"card_no" gorm:"size:18;index;comment:卡券编号"`
	UserID    uint           `json:"user_id" gorm:"index;comment:用户ID"`
	OrderID   uint           `json:"order_id" gorm:"index;comment:订单ID"`
	OrderNo   string         `json:"order_no" gorm:"size:32;comment:订单编号"`
	CourseID  uint           `json:"course_id" gorm:"index;comment:课程ID"`
	Amount    float64        `json:"amount" gorm:"comment:订单金额"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

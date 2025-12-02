package model

import (
	"time"

	"gorm.io/gorm"
)

// UserFeedback 用户反馈模型
type UserFeedback struct {
	ID              uint           `json:"id" gorm:"primarykey"`
	UserID          uint           `json:"user_id" gorm:"index"` // 关联用户ID
	User            User           `json:"-" gorm:"foreignKey:UserID"`
	FeedbackContent string         `json:"feedback_content" gorm:"type:text"` // 反馈内容
	Status          int            `json:"status" gorm:"default:0"`           // 状态：0-未确认，1-已确认
	ReplyContent    string         `json:"reply_content" gorm:"type:text"`    // 回复内容
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `json:"-" gorm:"index"`
}

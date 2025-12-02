package service

import (
	"errors"
	"exam-system/internal/config"
	"exam-system/internal/model"
	"exam-system/internal/pkg/database"
	"time"
)

var User = new(UserService)

type UserService struct{}

func (s *UserService) GetProfile(userId uint) (*model.User, error) {
	var user model.User
	if err := database.DB.First(&user, userId).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *UserService) UpdateProfile(userId uint, nickname, avatar string) error {
	updates := make(map[string]interface{})
	if nickname != "" {
		updates["nickname"] = nickname
	}
	if avatar != "" {
		updates["avatar"] = avatar
	}

	return database.DB.Model(&model.User{}).Where("id = ?", userId).Updates(updates).Error
}

// GetTokenExpireTime 获取用户token的过期时间
func (s *UserService) GetTokenExpireTime(userId uint) (int64, error) {
	// 获取JWT配置
	if config.GlobalConfig == nil {
		return 0, errors.New("配置未初始化")
	}
	if config.GlobalConfig == nil {
		return 0, errors.New("配置未初始化")
	}
	cfg := config.GlobalConfig

	// 计算过期时间
	expireSeconds := cfg.JWT.ExpireTime
	expireTime := time.Now().Add(time.Duration(expireSeconds) * time.Second)

	return expireTime.Unix(), nil
}

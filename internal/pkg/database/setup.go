package database

import (
	"fmt"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"exam-system/internal/config"
	"exam-system/internal/model"
)

// DB 全局数据库连接
var DB *gorm.DB

// Setup 初始化数据库连接和迁移
func Setup() error {
	var err error

	// 获取配置
	cfg := config.GlobalConfig.Database

	// 构建DSN
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.Username,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.DBName,
	)

	// 连接数据库
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("failed to connect database: %v", err)
	}

	// 自动迁移
	if err := DB.AutoMigrate(
		&model.User{},
		&model.Course{},
		&model.Question{},
		&model.Order{},
		&model.QRCode{},
		&model.AdminLoginLog{},
		&model.UserFeedback{},
		&model.ExamRecord{},
		&model.Card{},
		&model.CardRecord{},
	); err != nil {
		return fmt.Errorf("failed to migrate database: %v", err)
	}

	return nil
}

// GetDB 获取数据库连接
func GetDB() *gorm.DB {
	return DB
}

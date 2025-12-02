package model

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

// ExamConfigItem 考试配置项
type ExamConfigItem struct {
	Type  string `json:"type"`
	Count int    `json:"count"`
	Score int    `json:"score"`
}

// MockExamConfig 模拟考试配置
type MockExamConfig struct {
	Min   int `json:"min"`
	Count int `json:"count"`
	Score int `json:"score"`
}

// 课程分类
type CourseCategory struct {
	ID        uint   `gorm:"primarykey"`
	ParentID  uint   `gorm:"index"` // 父分类ID，0表示一级分类
	Name      string `gorm:"size:64"`
	Sort      int    // 排序
	Level     int    // 层级：1=一级分类，2=二级分类，3=三级分类
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

// 课程
type Course struct {
	ID             uint   `gorm:"primarykey"`
	Sort           int    `gorm:"default:0"` // 课程自身排序
	CategorySort1  int    `gorm:"default:0"` // 一级分类排序
	CategorySort2  int    `gorm:"default:0"` // 二级分类排序
	Name           string `gorm:"size:64"`
	Cover          string `gorm:"size:255"`
	CategoryLevel1 string `gorm:"size:50"` // 一级分类
	CategoryLevel2 string `gorm:"size:50"` // 二级分类
	Price          float64
	Description    string `gorm:"type:text"`
	ExpireDays     int    `gorm:"default:0"` // 课程有效期（天）
	ExamConfig     string `gorm:"type:json"` // 考试配置，JSON字符串
	MockExamConfig string `gorm:"type:json"` // 模拟考试配置，JSON字符串
	CreatedAt      time.Time
	UpdatedAt      time.Time
	DeletedAt      gorm.DeletedAt `gorm:"index"`
}

// GetExamConfig 获取考试配置
func (c *Course) GetExamConfig() ([]ExamConfigItem, error) {
	var config []ExamConfigItem
	if c.ExamConfig == "" {
		return config, nil
	}
	err := json.Unmarshal([]byte(c.ExamConfig), &config)
	return config, err
}

// SetExamConfig 设置考试配置
func (c *Course) SetExamConfig(config []ExamConfigItem) error {
	data, err := json.Marshal(config)
	if err != nil {
		return err
	}
	c.ExamConfig = string(data)
	return nil
}

// GetMockExamConfig 获取模拟考试配置
func (c *Course) GetMockExamConfig() (MockExamConfig, error) {
	var config MockExamConfig
	if c.MockExamConfig == "" {
		return config, nil
	}
	err := json.Unmarshal([]byte(c.MockExamConfig), &config)
	return config, err
}

// SetMockExamConfig 设置模拟考试配置
func (c *Course) SetMockExamConfig(config MockExamConfig) error {
	data, err := json.Marshal(config)
	if err != nil {
		return err
	}
	c.MockExamConfig = string(data)
	return nil
}

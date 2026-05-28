package model

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"
)

// 由于和question.go中存在冲突，把类型定义都合并到这里
type StringArray []string

// 实现 Scanner 接口
func (a *StringArray) Scan(value interface{}) error {
	if value == nil {
		*a = nil
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return errors.New("failed to unmarshal JSON value")
	}

	return json.Unmarshal(bytes, a)
}

// 实现 Valuer 接口
func (a StringArray) Value() (driver.Value, error) {
	if a == nil {
		return nil, nil
	}
	return json.Marshal(a)
}

// 问题选项
type QuestionOption struct {
	Label string `json:"label"`
	Text  string `json:"text"`
}

// QuestionOptions 类型用于存储选项数组
type QuestionOptions []QuestionOption

// 实现 Scanner 接口
func (o *QuestionOptions) Scan(value interface{}) error {
	if value == nil {
		*o = nil
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return errors.New("failed to unmarshal JSON value")
	}

	var opts []QuestionOption
	if err := json.Unmarshal(bytes, &opts); err == nil {
		*o = opts
		return nil
	}

	var strOpts []string
	if err := json.Unmarshal(bytes, &strOpts); err == nil {
		for i, text := range strOpts {
			if strings.TrimSpace(text) == "" {
				continue
			}
			parts := strings.SplitN(text, ".", 2)
			if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
				*o = append(*o, QuestionOption{Label: parts[0], Text: strings.TrimSpace(parts[1])})
			} else {
				label := string(rune('A' + i))
				*o = append(*o, QuestionOption{Label: label, Text: strings.TrimSpace(text)})
			}
		}
		return nil
	}

	return errors.New("failed to unmarshal JSON value for options")
}

// 实现 Valuer 接口
func (o QuestionOptions) Value() (driver.Value, error) {
	if o == nil {
		return nil, nil
	}
	return json.Marshal(o)
}

type Question struct {
	ID          uint            `json:"id" gorm:"primarykey"`
	Type        string          `json:"type" gorm:"size:20"` // single或multiple
	Question    string          `json:"question" gorm:"type:text"`
	Options     QuestionOptions `json:"options" gorm:"type:json"` // JSON格式存储选项
	Answer      string          `json:"answer" gorm:"size:255"`   // 改回字符串类型
	Explanation string          `json:"explanation" gorm:"type:text"`
	CourseID    uint            `json:"course_id" gorm:"index"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
	DeletedAt   gorm.DeletedAt  `json:"-" gorm:"index"`
}

// type Exam struct {
// 	ID        uint           `json:"id" gorm:"primarykey"`
// 	CourseID  uint           `json:"course_id" gorm:"index"`
// 	Name      string         `json:"name" gorm:"size:64"`
// 	Duration  int            `json:"duration"` // 考试时长(分钟)
// 	Questions []Question     `json:"questions" gorm:"many2many:exam_questions;"`
// 	CreatedAt time.Time      `json:"created_at"`
// 	UpdatedAt time.Time      `json:"updated_at"`
// 	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
// }

type ExamRecord struct {
	ID           uint           `json:"id" gorm:"primarykey"`
	UserID       uint           `json:"user_id" gorm:"index"`
	CourseID     uint           `json:"course_id"`
	Score        float64        `json:"score"`
	WrongAnswers string         `json:"wrong_answers" gorm:"type:json"` // JSON格式存储错题ID数组
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`
}

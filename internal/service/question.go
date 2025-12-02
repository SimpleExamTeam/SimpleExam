package service

import (
	"encoding/json"
	"errors"
	"exam-system/internal/model"
	"exam-system/internal/pkg/database"
	"strings"
	"time"

	"gorm.io/gorm"
)

var Question = new(QuestionService)

type QuestionService struct{}

// 问题返回结构
type QuestionResponse struct {
	ID          uint                   `json:"id"`
	Type        string                 `json:"type"`
	Question    string                 `json:"question"`
	Options     []model.QuestionOption `json:"options"`
	Answer      string                 `json:"answer"`
	Explanation string                 `json:"explanation"`
	CourseID    uint                   `json:"course_id"`
}

// 获取课程题目
func (s *QuestionService) GetQuestionsByCourse(userId, courseId uint, questionType string) ([]QuestionResponse, error) {
	// 1. 检查用户是否购买了该课程且未过期
	var order model.Order
	err := database.DB.Where("user_id = ? AND course_id = ? AND status = ?",
		userId, courseId, "paid").
		Where("expire_time IS NULL OR expire_time > ?", time.Now()).
		Order("expire_time DESC").
		First(&order).Error

	if err != nil {
		return nil, errors.New("您尚未购买该课程或课程已过期")
	}

	// 2. 查询指定类型的题目
	query := database.DB.Table("questions").Where("course_id = ?", courseId)

	if questionType != "" && questionType != "all" {
		query = query.Where("type = ?", questionType)
	}

	// 使用临时结构体接收数据
	type RawQuestion struct {
		ID          uint   `json:"id"`
		Type        string `json:"type"`
		Question    string `json:"question"`
		Options     string `json:"options"`
		Answer      string `json:"answer"`
		Explanation string `json:"explanation"`
		CourseID    uint   `json:"course_id"`
		CreatedAt   time.Time
		UpdatedAt   time.Time
		DeletedAt   gorm.DeletedAt
	}

	var rawQuestions []RawQuestion
	err = query.Find(&rawQuestions).Error
	if err != nil {
		return nil, err
	}

	// 3. 转换为标准响应格式
	var response []QuestionResponse
	for _, q := range rawQuestions {
		// 手动解析options字段
		var options []model.QuestionOption
		if q.Options != "" {
			// 首先尝试解析为字符串数组，这是数据库中实际存储的格式
			var textOptions []string
			if err := json.Unmarshal([]byte(q.Options), &textOptions); err == nil {
				// 成功解析为字符串数组
				for _, optText := range textOptions {
					// 处理选项文本中的标签前缀，如 "A.选项内容"
					parts := strings.SplitN(optText, ".", 2)
					if len(parts) == 2 {
						// 正确分割出标签和文本
						options = append(options, model.QuestionOption{
							Label: parts[0],
							Text:  parts[1],
						})
					} else {
						// 如果没有分隔符，使用整个字符串作为文本
						options = append(options, model.QuestionOption{
							Label: "", // 空标签
							Text:  optText,
						})
					}
				}
			} else {
				// 如果解析失败，尝试其他格式
				// 尝试解析为 QuestionOption 数组
				if err := json.Unmarshal([]byte(q.Options), &options); err != nil {
					// 如果仍然失败，使用默认选项
					for i := 0; i < 4; i++ {
						label := string(rune('A' + i))
						options = append(options, model.QuestionOption{
							Label: label,
							Text:  "选项" + label,
						})
					}
				}
			}
		} else {
			// 如果选项为空，根据题目类型生成默认选项
			if q.Type == "judge" {
				// 判断题默认有两个选项
				options = append(options, model.QuestionOption{Label: "A", Text: "正确"})
				options = append(options, model.QuestionOption{Label: "B", Text: "错误"})
			} else {
				// 其他类型默认有4个选项
				for i := 0; i < 4; i++ {
					label := string(rune('A' + i))
					options = append(options, model.QuestionOption{
						Label: label,
						Text:  "选项" + label,
					})
				}
			}
		}

		// 创建响应结构
		response = append(response, QuestionResponse{
			ID:          q.ID,
			Type:        q.Type,
			Question:    q.Question,
			Options:     options,
			Answer:      q.Answer,
			Explanation: q.Explanation,
			CourseID:    q.CourseID,
		})
	}

	return response, nil
}

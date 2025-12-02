package service

import (
	"errors"
	"exam-system/internal/pkg/database"
	"time"
)

var Exam = new(ExamService)

type ExamService struct{}

// ExamResultItem 表示考试结果项
type ExamResultItem struct {
	ID         uint      `json:"id"`
	Score      float64   `json:"score"`
	CourseID   uint      `json:"course_id"`
	CourseName string    `json:"course_name"`
	CreatedAt  time.Time `json:"created_at"`
	Passed     bool      `json:"passed"`
}

// GetAllResults 获取用户的所有考试结果
func (s *ExamService) GetAllResults(userId uint) ([]ExamResultItem, error) {
	// 定义一个临时结构体用于查询结果
	type QueryResult struct {
		ID             uint      `json:"id"`
		Score          float64   `json:"score"`
		CourseID       uint      `json:"course_id"`
		CategoryLevel2 string    `json:"category_level2"`
		Name           string    `json:"name"`
		CreatedAt      time.Time `json:"created_at"`
	}

	var queryResults []QueryResult

	// 查询用户的考试记录，联表查询课程信息
	err := database.DB.Table("exam_records").
		Select("exam_records.id, exam_records.score, exam_records.course_id, courses.category_level2 AS category_level2, courses.name, exam_records.created_at").
		Joins("LEFT JOIN courses ON exam_records.course_id = courses.id").
		Where("exam_records.user_id = ? AND exam_records.deleted_at IS NULL", userId).
		Order("exam_records.created_at DESC").
		Find(&queryResults).Error

	if err != nil {
		return nil, errors.New("获取考试记录失败: " + err.Error())
	}

	// 组装返回数据
	var results []ExamResultItem
	for _, result := range queryResults {
		// 拼接课程名称：category_level2-name
		var courseName string
		if result.CategoryLevel2 != "" {
			courseName = result.CategoryLevel2 + "-" + result.Name
		} else {
			courseName = result.Name // 如果二级分类为空，则只使用课程名称
		}

		results = append(results, ExamResultItem{
			ID:         result.ID,
			Score:      result.Score,
			CourseID:   result.CourseID,
			CourseName: courseName,
			CreatedAt:  result.CreatedAt,
			Passed:     result.Score >= 60, // 默认60分及格
		})
	}

	return results, nil
}

package types

// 提交练习答案的请求结构
type SubmitPracticeRequest struct {
	QuestionID uint     `json:"question_id" binding:"required"`
	Answer     []string `json:"answer" binding:"required"`
}

package types

type SubmitAnswerRequest struct {
	QuestionID uint     `json:"questionId"`
	Answer     []string `json:"answer"`
}

type ExamResult struct {
	Score        float64 `json:"score"`
	Passed       bool    `json:"passed"`
	CorrectCount int     `json:"correctCount"`
	WrongCount   int     `json:"wrongCount"`
}

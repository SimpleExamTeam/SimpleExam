package service

import (
	"bytes"
	"encoding/json"
	"errors"
	"exam-system/internal/config"
	"exam-system/internal/model"
	"fmt"
	"io"
	"net/http"
	"time"
)

var AI = new(AIService)

type AIService struct{}

type deepseekRequest struct {
	Model    string            `json:"model"`
	Messages []deepseekMessage `json:"messages"`
}

type deepseekMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type deepseekResponse struct {
	Choices []struct {
		Message deepseekMessage `json:"message"`
	} `json:"choices"`
}

func (s *AIService) GenerateExplanation(questionText, questionType string, options []model.QuestionOption, answer string) (string, error) {
	cfg := config.GlobalConfig.AI.Explanation
	if cfg.APIKey == "" {
		return "", errors.New("未配置AI API密钥")
	}

	optionsText := ""
	for _, opt := range options {
		optionsText += fmt.Sprintf("%s. %s\n", opt.Label, opt.Text)
	}

	typeName := typeLabel(questionType)

	prompt := fmt.Sprintf(`你是一个考试题目解析助手。请为以下题目生成简短解析。

题型：%s
题目：%s
选项：
%s
正确答案：%s

要求：
- 使用纯文本，禁止使用Markdown格式（禁用**加粗**、编号列表、标题等任何标记语法）
- 解析尽量简短，控制在150字以内
- 直接输出解析内容，不要添加任何额外说明`, typeName, questionText, optionsText, answer)

	reqBody := deepseekRequest{
		Model: "deepseek-chat",
		Messages: []deepseekMessage{
			{Role: "system", Content: "你是一个考试题目解析助手，擅长给出简短、准确的题目解析。请始终使用纯文本，不要使用任何Markdown格式。"},
			{Role: "user", Content: prompt},
		},
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", cfg.APIURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+cfg.APIKey)

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("调用AI接口失败: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("AI接口返回错误: %s, %s", resp.Status, string(body))
	}

	var aiResp deepseekResponse
	if err := json.Unmarshal(body, &aiResp); err != nil {
		return "", err
	}

	if len(aiResp.Choices) == 0 {
		return "", errors.New("AI未返回有效解析")
	}

	return aiResp.Choices[0].Message.Content, nil
}

func typeLabel(questionType string) string {
	switch questionType {
	case "single":
		return "单选题"
	case "multiple":
		return "多选题"
	case "judge":
		return "判断题"
	default:
		return questionType
	}
}

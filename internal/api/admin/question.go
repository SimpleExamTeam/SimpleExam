package admin

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"exam-system/internal/model"
	"exam-system/internal/pkg/database"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// 题库查询参数
type QuestionQuery struct {
	Page     int    `form:"page,default=1"`
	Size     int    `form:"size,default=10"`
	Type     string `form:"type"`
	Question string `form:"question"`
	CourseID uint   `form:"course_id"`
}

// GetQuestions 获取题库列表
func GetQuestions(c *gin.Context) {
	var query QuestionQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "参数错误",
		})
		return
	}

	// 设置默认值
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.Size <= 0 {
		query.Size = 10
	}

	db := database.DB.Model(&model.Question{})

	// 构建查询条件
	if query.Type != "" {
		db = db.Where("type = ?", query.Type)
	}
	if query.Question != "" {
		db = db.Where("question LIKE ?", "%"+query.Question+"%")
	}
	if query.CourseID > 0 {
		db = db.Where("course_id = ?", query.CourseID)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "获取题目总数失败",
		})
		return
	}

	// 使用临时结构体接收原始数据
	type RawQuestion struct {
		ID          uint           `json:"id" gorm:"column:id"`
		Type        string         `json:"type" gorm:"column:type"`
		Question    string         `json:"question" gorm:"column:question"`
		Options     string         `json:"options" gorm:"column:options"` // 接收原始字符串
		Answer      string         `json:"answer" gorm:"column:answer"`
		Explanation string         `json:"explanation" gorm:"column:explanation"`
		CourseID    uint           `json:"course_id" gorm:"column:course_id"`
		CreatedAt   time.Time      `json:"created_at" gorm:"column:created_at"`
		UpdatedAt   time.Time      `json:"updated_at" gorm:"column:updated_at"`
		DeletedAt   gorm.DeletedAt `json:"-" gorm:"column:deleted_at"`
	}

	var rawQuestions []RawQuestion
	if err := db.Order("id DESC").
		Offset((query.Page - 1) * query.Size).
		Limit(query.Size).
		Find(&rawQuestions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "获取题目列表失败",
		})
		return
	}

	// 处理返回数据
	questionList := make([]gin.H, 0)
	for _, q := range rawQuestions {
		// 查询课程信息
		var course model.Course
		courseName := "未知课程"
		if err := database.DB.First(&course, q.CourseID).Error; err == nil {
			courseName = course.Name
		}

		// 获取题目类型的中文描述
		typeDesc := getQuestionTypeDesc(q.Type)

		// 解析选项
		var questionOptions model.QuestionOptions
		if q.Type == "judge" {
			// 判断题使用固定格式
			questionOptions = model.QuestionOptions{
				{Label: "A", Text: "正确"},
				{Label: "B", Text: "错误"},
			}
		} else {
			// 尝试解析选项
			var options []string
			if err := json.Unmarshal([]byte(q.Options), &options); err != nil {
				// 如果解析失败，使用空数组
				options = []string{}
			}

			// 转换为 QuestionOption 格式
			questionOptions = make(model.QuestionOptions, len(options))
			for i, text := range options {
				// 如果文本包含标签前缀（如 "A.选项内容"），则提取实际内容
				parts := strings.SplitN(text, ".", 2)
				if len(parts) == 2 {
					questionOptions[i] = model.QuestionOption{
						Label: parts[0],
						Text:  strings.TrimSpace(parts[1]),
					}
				} else {
					questionOptions[i] = model.QuestionOption{
						Label: string(rune('A' + i)),
						Text:  text,
					}
				}
			}
		}

		questionList = append(questionList, gin.H{
			"id":          q.ID,
			"type":        q.Type,
			"type_desc":   typeDesc,
			"question":    q.Question,
			"options":     questionOptions,
			"answer":      q.Answer,
			"explanation": q.Explanation,
			"course_id":   q.CourseID,
			"course_name": courseName,
			"created_at":  q.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": gin.H{
			"total": total,
			"items": questionList,
		},
	})
}

// GetQuestion 获取单个题目
func GetQuestion(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "参数错误",
		})
		return
	}

	// 使用临时结构体接收原始数据
	type RawQuestion struct {
		ID          uint           `json:"id" gorm:"column:id"`
		Type        string         `json:"type" gorm:"column:type"`
		Question    string         `json:"question" gorm:"column:question"`
		Options     string         `json:"options" gorm:"column:options"` // 接收原始字符串
		Answer      string         `json:"answer" gorm:"column:answer"`
		Explanation string         `json:"explanation" gorm:"column:explanation"`
		CourseID    uint           `json:"course_id" gorm:"column:course_id"`
		CreatedAt   time.Time      `json:"created_at" gorm:"column:created_at"`
		UpdatedAt   time.Time      `json:"updated_at" gorm:"column:updated_at"`
		DeletedAt   gorm.DeletedAt `json:"-" gorm:"column:deleted_at"`
	}

	var question RawQuestion
	if err := database.DB.Table("questions").First(&question, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code": 404,
			"msg":  "题目不存在",
		})
		return
	}

	// 查询课程信息
	var course model.Course
	courseName := "未知课程"
	if err = database.DB.First(&course, question.CourseID).Error; err == nil {
		courseName = course.Name
	}

	// 获取题目类型的中文描述
	typeDesc := getQuestionTypeDesc(question.Type)

	// 解析选项
	var questionOptions model.QuestionOptions
	if question.Type == "judge" {
		// 判断题使用固定格式
		questionOptions = model.QuestionOptions{
			{Label: "A", Text: "正确"},
			{Label: "B", Text: "错误"},
		}
	} else {
		// 尝试解析选项
		var options []string
		if err := json.Unmarshal([]byte(question.Options), &options); err != nil {
			// 如果解析失败，使用空数组
			options = []string{}
		}

		// 转换为 QuestionOption 格式
		questionOptions = make(model.QuestionOptions, len(options))
		for i, text := range options {
			// 如果文本包含标签前缀（如 "A.选项内容"），则提取实际内容
			parts := strings.SplitN(text, ".", 2)
			if len(parts) == 2 {
				questionOptions[i] = model.QuestionOption{
					Label: parts[0],
					Text:  strings.TrimSpace(parts[1]),
				}
			} else {
				questionOptions[i] = model.QuestionOption{
					Label: string(rune('A' + i)),
					Text:  text,
				}
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": gin.H{
			"id":          question.ID,
			"type":        question.Type,
			"type_desc":   typeDesc,
			"question":    question.Question,
			"options":     questionOptions,
			"answer":      question.Answer,
			"explanation": question.Explanation,
			"course_id":   question.CourseID,
			"course_name": courseName,
			"created_at":  question.CreatedAt,
		},
	})
}

// getQuestionTypeDesc 获取题目类型的中文描述
func getQuestionTypeDesc(qType string) string {
	switch qType {
	case "single":
		return "单选题"
	case "multiple":
		return "多选题"
	case "judge":
		return "判断题"
	default:
		return "未知类型"
	}
}

// QuestionRequest 创建/更新题目请求
type QuestionRequest struct {
	Type        string                `json:"type" binding:"required"`
	Question    string                `json:"question" binding:"required"`
	Options     model.QuestionOptions `json:"options" binding:"required"`
	Answer      string                `json:"answer" binding:"required"`
	Explanation string                `json:"explanation"`
	CourseID    uint                  `json:"course_id" binding:"required"`
}

// validateQuestionOptions 验证题目选项和答案格式
func validateQuestionOptions(qType string, options model.QuestionOptions, answer string) error {
	switch qType {
	case "judge":
		// 判断题必须只有两个选项：A.正确、B.错误
		if len(options) != 2 {
			return errors.New("判断题必须有且只有两个选项")
		}
		if options[0].Text != "正确" || options[1].Text != "错误" {
			return errors.New("判断题选项必须为：正确、错误")
		}
		if answer != "A" && answer != "B" {
			return errors.New("判断题答案必须为A或B")
		}
	case "single":
		// 单选题选项必须是A-D，答案必须是其中之一
		if len(options) < 2 || len(options) > 26 {
			return errors.New("单选题选项数量必须在2-26之间")
		}
		if answer < "A" || answer > string(rune('A'+len(options)-1)) {
			return errors.New("单选题答案必须是选项中的一个字母")
		}
	case "multiple":
		// 多选题选项必须是A-Z，答案必须是其中的一个或多个
		if len(options) < 2 || len(options) > 26 {
			return errors.New("多选题选项数量必须在2-26之间")
		}
		// 验证答案格式（必须是大写字母组合，如"ABC"）
		answerMap := make(map[rune]bool)
		for _, a := range answer {
			if a < 'A' || a > rune('A'+len(options)-1) {
				return errors.New("多选题答案必须是选项标签的组合")
			}
			answerMap[a] = true
		}
		if len(answer) == 0 {
			return errors.New("多选题答案不能为空")
		}
	default:
		return errors.New("不支持的题目类型")
	}
	return nil
}

// CreateQuestion 创建题目
func CreateQuestion(c *gin.Context) {
	var req QuestionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "参数错误",
		})
		return
	}

	// 验证题目类型
	if req.Type != "single" && req.Type != "multiple" && req.Type != "judge" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "题目类型错误，只支持single(单选题)、multiple(多选题)或judge(判断题)",
		})
		return
	}

	// 验证选项和答案格式
	if err := validateQuestionOptions(req.Type, req.Options, req.Answer); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  err.Error(),
		})
		return
	}

	// 验证课程是否存在
	var course model.Course
	if err := database.DB.First(&course, req.CourseID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "课程不存在",
		})
		return
	}

	// 创建题目 - 将选项格式转换为字符串数组格式
	var optionsJSON []byte
	var err error

	if req.Type == "judge" {
		// 判断题使用固定格式
		optionsJSON = []byte(`["A.正确","B.错误"]`)
	} else {
		// 构建格式化的选项数组
		options := make([]string, len(req.Options))
		for i, opt := range req.Options {
			// 只使用文本内容，不包含标签
			options[i] = string(rune('A'+i)) + "." + opt.Text
		}
		optionsJSON, err = json.Marshal(options)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code": 500,
				"msg":  "处理选项数据失败",
			})
			return
		}
	}

	// 创建题目记录
	question := model.Question{
		Type:        req.Type,
		Question:    req.Question,
		Answer:      req.Answer,
		Explanation: req.Explanation,
		CourseID:    req.CourseID,
	}

	// 使用原生SQL语句来插入JSON格式的选项
	tx := database.DB.Begin()
	if err := tx.Create(&question).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "创建题目失败",
		})
		return
	}

	// 使用原生SQL更新options字段为正确的JSON格式
	if err := tx.Exec("UPDATE questions SET options = ? WHERE id = ?", string(optionsJSON), question.ID).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "更新选项数据失败",
		})
		return
	}

	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "提交事务失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": gin.H{
			"id": question.ID,
		},
	})
}

// UpdateQuestion 更新题目
func UpdateQuestion(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "参数错误",
		})
		return
	}

	var req QuestionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "参数错误",
		})
		return
	}

	// 验证题目类型
	if req.Type != "single" && req.Type != "multiple" && req.Type != "judge" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "题目类型错误，只支持single(单选题)、multiple(多选题)或judge(判断题)",
		})
		return
	}

	// 验证选项和答案格式
	if err := validateQuestionOptions(req.Type, req.Options, req.Answer); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  err.Error(),
		})
		return
	}

	// 验证课程是否存在
	var course model.Course
	if err := database.DB.First(&course, req.CourseID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "课程不存在",
		})
		return
	}

	// 检查题目是否存在
	var count int64
	if err := database.DB.Model(&model.Question{}).Where("id = ?", id).Count(&count).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "查询题目失败",
		})
		return
	}

	if count == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"code": 404,
			"msg":  "题目不存在",
		})
		return
	}

	// 创建题目 - 将选项格式转换为字符串数组格式
	var optionsJSON []byte

	if req.Type == "judge" {
		// 判断题使用固定格式
		optionsJSON = []byte(`["A.正确","B.错误"]`)
	} else {
		// 构建格式化的选项数组
		options := make([]string, len(req.Options))
		for i, opt := range req.Options {
			// 只使用文本内容，不包含标签
			options[i] = string(rune('A'+i)) + "." + opt.Text
		}
		optionsJSON, err = json.Marshal(options)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code": 500,
				"msg":  "处理选项数据失败",
			})
			return
		}
	}

	// 开始事务
	tx := database.DB.Begin()

	// 更新基本字段
	if err := tx.Model(&model.Question{}).Where("id = ?", id).Updates(map[string]interface{}{
		"type":        req.Type,
		"question":    req.Question,
		"answer":      req.Answer,
		"explanation": req.Explanation,
		"course_id":   req.CourseID,
	}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "更新题目失败",
		})
		return
	}

	// 使用原生SQL更新options字段为正确的JSON格式
	if err := tx.Exec("UPDATE questions SET options = ? WHERE id = ?", string(optionsJSON), id).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "更新选项数据失败",
		})
		return
	}

	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "提交事务失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "更新成功",
	})
}

// DeleteQuestion 删除题目
func DeleteQuestion(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "参数错误",
		})
		return
	}

	result := database.DB.Delete(&model.Question{}, id)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "删除题目失败",
		})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"code": 404,
			"msg":  "题目不存在",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "删除成功",
	})
}

// ExportQuestions 导出题库
func ExportQuestions(c *gin.Context) {
	courseID, _ := strconv.ParseUint(c.Query("course_id"), 10, 32)

	// 构建查询条件
	db := database.DB.Model(&model.Question{})
	if courseID > 0 {
		db = db.Where("course_id = ?", courseID)
	}

	// 使用临时结构体接收原始数据
	type RawQuestion struct {
		ID          uint           `json:"id" gorm:"column:id"`
		Type        string         `json:"type" gorm:"column:type"`
		Question    string         `json:"question" gorm:"column:question"`
		Options     string         `json:"options" gorm:"column:options"` // 接收原始字符串
		Answer      string         `json:"answer" gorm:"column:answer"`
		Explanation string         `json:"explanation" gorm:"column:explanation"`
		CourseID    uint           `json:"course_id" gorm:"column:course_id"`
		CreatedAt   time.Time      `json:"created_at" gorm:"column:created_at"`
		UpdatedAt   time.Time      `json:"updated_at" gorm:"column:updated_at"`
		DeletedAt   gorm.DeletedAt `json:"-" gorm:"column:deleted_at"`
	}

	var questions []RawQuestion
	if err := db.Find(&questions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "获取题目失败",
		})
		return
	}

	// 返回CSV格式数据
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", "attachment; filename=questions.csv")

	// 添加BOM头，解决Excel打开中文乱码问题
	c.Writer.Write([]byte{0xEF, 0xBB, 0xBF})

	// 创建CSV写入器
	writer := csv.NewWriter(c.Writer)
	defer writer.Flush()

	// 写入CSV头
	header := []string{"ID", "题目类型", "题目内容", "选项", "答案", "解析", "课程ID", "题目类型说明"}
	writer.Write(header)

	// 写入数据
	for _, q := range questions {
		// 获取题目类型的中文描述
		typeDesc := getQuestionTypeDesc(q.Type)

		// 处理选项格式
		var optionsStr string
		if q.Type == "judge" {
			// 判断题使用固定格式，不需要从数据库读取
			optionsStr = `["A.正确","B.错误"]`
		} else {
			// 直接使用数据库中的选项字符串
			optionsStr = q.Options
		}

		record := []string{
			strconv.FormatUint(uint64(q.ID), 10),
			q.Type,
			q.Question,
			optionsStr,
			q.Answer,
			q.Explanation,
			strconv.FormatUint(uint64(q.CourseID), 10),
			typeDesc,
		}

		writer.Write(record)
	}
}

// ImportQuestions 导入题库
func ImportQuestions(c *gin.Context) {
	// 获取上传的文件
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "请选择要上传的CSV文件",
		})
		return
	}

	// 打开文件
	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "无法打开文件",
		})
		return
	}
	defer src.Close()

	// 检测并跳过BOM头
	bomBuffer := make([]byte, 3)
	if _, err := src.Read(bomBuffer); err != nil || bomBuffer[0] != 0xEF || bomBuffer[1] != 0xBB || bomBuffer[2] != 0xBF {
		// 如果不是BOM头，回到文件开始处
		src.Seek(0, 0)
	}

	// 读取CSV文件内容
	reader := csv.NewReader(src)
	reader.FieldsPerRecord = -1 // 允许每行不同的字段数

	// 读取并跳过第一行（表头）
	if _, err := reader.Read(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "CSV文件格式不正确",
		})
		return
	}

	// 开始事务
	tx := database.DB.Begin()

	// 导入数据
	importCount := 0
	errorCount := 0
	errorMessages := make([]string, 0)

	for lineNum := 2; ; lineNum++ { // 从第2行开始（表头为第1行）
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			errorCount++
			errorMessages = append(errorMessages, fmt.Sprintf("第%d行: 读取错误", lineNum))
			continue
		}

		if len(record) < 7 {
			errorCount++
			errorMessages = append(errorMessages, fmt.Sprintf("第%d行: 字段数量不足", lineNum))
			continue // 跳过格式不正确的行
		}

		// 解析数据
		courseID, err := strconv.ParseUint(record[6], 10, 32)
		if err != nil {
			errorCount++
			errorMessages = append(errorMessages, fmt.Sprintf("第%d行: 课程ID格式错误", lineNum))
			continue
		}

		// 验证课程是否存在
		var courseCount int64
		if err := tx.Model(&model.Course{}).Where("id = ?", courseID).Count(&courseCount).Error; err != nil || courseCount == 0 {
			errorCount++
			errorMessages = append(errorMessages, fmt.Sprintf("第%d行: 课程ID %d 不存在", lineNum, courseID))
			continue
		}

		questionType := strings.TrimSpace(record[1])
		// 验证题目类型
		if questionType != "single" && questionType != "multiple" && questionType != "judge" {
			errorCount++
			errorMessages = append(errorMessages, fmt.Sprintf("第%d行: 题目类型错误，只支持single(单选题)、multiple(多选题)或judge(判断题)", lineNum))
			continue
		}

		// 处理选项
		var optionsJSON string
		var questionOptions model.QuestionOptions

		if questionType == "judge" {
			// 判断题固定选项格式
			optionsJSON = `["A.正确","B.错误"]`
			questionOptions = model.QuestionOptions{
				{Label: "A", Text: "正确"},
				{Label: "B", Text: "错误"},
			}
		} else {
			// 尝试解析选项JSON
			optionsField := strings.TrimSpace(record[3])
			if optionsField == "" {
				errorCount++
				errorMessages = append(errorMessages, fmt.Sprintf("第%d行: 选项为空", lineNum))
				continue
			}

			var options []string
			if err = json.Unmarshal([]byte(optionsField), &options); err != nil {
				errorCount++
				errorMessages = append(errorMessages, fmt.Sprintf("第%d行: 选项JSON格式错误", lineNum))
				continue
			}

			// 转换为 QuestionOption 格式
			questionOptions = make(model.QuestionOptions, len(options))
			formattedOptions := make([]string, len(options))

			for i, text := range options {
				// 清理选项文本，确保只保留实际内容
				optionText := strings.TrimSpace(text)

				// 如果文本包含标签前缀（如 "A.选项内容"），则提取实际内容
				parts := strings.SplitN(optionText, ".", 2)
				if len(parts) == 2 {
					optionText = strings.TrimSpace(parts[1])
				}

				// 创建选项对象
				questionOptions[i] = model.QuestionOption{
					Label: string(rune('A' + i)),
					Text:  optionText,
				}

				// 创建格式化的选项文本
				formattedOptions[i] = string(rune('A'+i)) + "." + optionText
			}

			// 将格式化后的选项转为JSON字符串
			optionsBytes, err := json.Marshal(formattedOptions)
			if err != nil {
				errorCount++
				errorMessages = append(errorMessages, fmt.Sprintf("第%d行: 选项格式化失败", lineNum))
				continue
			}
			optionsJSON = string(optionsBytes)
		}

		// 验证答案
		answer := strings.TrimSpace(record[4])
		if answer == "" {
			errorCount++
			errorMessages = append(errorMessages, fmt.Sprintf("第%d行: 答案为空", lineNum))
			continue
		}

		// 清理答案，确保只包含选项序号（ABCDE等）
		answer = cleanAnswer(answer)

		// 根据题目类型验证答案格式
		switch questionType {
		case "judge":
			if answer != "A" && answer != "B" {
				errorCount++
				errorMessages = append(errorMessages, fmt.Sprintf("第%d行: 判断题答案必须为A(正确)或B(错误)", lineNum))
				continue
			}
		case "single":
			// 单选题答案必须是A-Z中的一个字母
			if len(answer) != 1 || answer[0] < 'A' || answer[0] > 'Z' {
				errorCount++
				errorMessages = append(errorMessages, fmt.Sprintf("第%d行: 单选题答案必须是A-Z中的一个字母", lineNum))
				continue
			}
			// 检查答案是否在选项范围内
			if int(answer[0]-'A') >= len(questionOptions) {
				errorCount++
				errorMessages = append(errorMessages, fmt.Sprintf("第%d行: 答案%s超出了选项范围", lineNum, answer))
				continue
			}
		case "multiple":
			// 多选题答案必须是A-Z的组合
			for _, ch := range answer {
				if ch < 'A' || ch > 'Z' {
					errorCount++
					errorMessages = append(errorMessages, fmt.Sprintf("第%d行: 多选题答案必须是A-Z的组合", lineNum))
					continue
				}
				// 检查答案是否在选项范围内
				if int(ch-'A') >= len(questionOptions) {
					errorCount++
					errorMessages = append(errorMessages, fmt.Sprintf("第%d行: 答案%s中的%c超出了选项范围", lineNum, answer, ch))
					continue
				}
			}
		}

		// 创建题目
		question := model.Question{
			Type:        questionType,
			Question:    record[2],
			Answer:      answer,
			Explanation: record[5],
			CourseID:    uint(courseID),
		}

		// 跳过ID字段，让数据库自动生成
		if err := tx.Create(&question).Error; err != nil {
			errorCount++
			errorMessages = append(errorMessages, fmt.Sprintf("第%d行: 创建题目失败: %s", lineNum, err.Error()))
			continue
		}

		// 使用原生SQL更新options字段为正确的JSON格式
		if err := tx.Exec("UPDATE questions SET options = ? WHERE id = ?", optionsJSON, question.ID).Error; err != nil {
			errorCount++
			errorMessages = append(errorMessages, fmt.Sprintf("第%d行: 更新选项数据失败: %s", lineNum, err.Error()))
			continue
		}

		importCount++
	}

	// 如果有错误但也有成功导入的记录，仍然提交事务
	if importCount > 0 {
		if err := tx.Commit().Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{
				"code": 500,
				"msg":  "导入失败: " + err.Error(),
			})
			return
		}
	} else if errorCount > 0 {
		// 如果全部导入失败，回滚事务
		tx.Rollback()
	} else {
		// 如果没有记录，也提交事务
		tx.Commit()
	}

	response := gin.H{
		"code": 200,
		"data": gin.H{
			"import_count": importCount,
			"error_count":  errorCount,
		},
		"msg": "导入完成",
	}

	// 如果有错误，添加错误信息
	if errorCount > 0 {
		if len(errorMessages) > 10 {
			errorMessages = append(errorMessages[:10], "...")
		}
		response["errors"] = errorMessages
	}

	c.JSON(http.StatusOK, response)
}

// cleanAnswer 清理答案，确保只包含选项序号（ABCDE等）
func cleanAnswer(answer string) string {
	// 将答案转为大写
	answer = strings.ToUpper(answer)

	// 过滤出所有A-Z的字符
	var result strings.Builder
	for _, ch := range answer {
		if ch >= 'A' && ch <= 'Z' {
			result.WriteRune(ch)
		}
	}

	return result.String()
}

// BatchDeleteQuestions 批量删除题目
func BatchDeleteQuestions(c *gin.Context) {
	var req struct {
		IDs []uint `json:"ids" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "参数错误",
		})
		return
	}

	if len(req.IDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "请选择要删除的题目",
		})
		return
	}

	// 执行批量删除
	result := database.DB.Delete(&model.Question{}, req.IDs)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "删除题目失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "删除成功",
		"data": gin.H{
			"deleted_count": result.RowsAffected,
		},
	})
}

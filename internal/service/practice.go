package service

import (
	"encoding/json"
	"exam-system/internal/model"
	"exam-system/internal/pkg/database"
	"strings"
	"time"
)

var Practice = new(PracticeService)

type PracticeService struct{}

// 错题详情
type WrongQuestionDetail struct {
	ID          uint                   `json:"id"`
	Type        string                 `json:"type"` // 题目类型
	Question    string                 `json:"question"`
	Options     []model.QuestionOption `json:"options"`
	Answer      string                 `json:"answer"`
	Explanation string                 `json:"explanation"`
	UpdatedAt   time.Time              `json:"updated_at"`
	CourseID    uint                   `json:"course_id"`
	CourseName  string                 `json:"course_name"` // 新增字段：课程名称
}

// 错题统计信息（课程维度）
type WrongQuestionCourse struct {
	CourseID   uint   `json:"course_id"`
	CourseName string `json:"course_name"`
	Single     int    `json:"single"`
	Multiple   int    `json:"multiple"`
	Judge      int    `json:"judge"`
	Total      int    `json:"total"`
}

// 获取错题列表 - 重构版本
func (s *PracticeService) GetWrongQuestions(userId uint, courseId int, page, pageSize int) ([]WrongQuestionDetail, int64, error) {
	// 1. 查询用户的所有考试记录，并关联订单表检查过期时间
	var records []model.ExamRecord
	query := database.DB.Table("exam_records").
		Joins("LEFT JOIN orders ON exam_records.course_id = orders.course_id AND exam_records.user_id = orders.user_id").
		Where("exam_records.user_id = ?", userId).
		Where("orders.status = ?", "paid").
		Where("(orders.expire_time IS NULL OR orders.expire_time > ?)", time.Now())

	// 如果指定了课程ID，则过滤特定课程的记录
	if courseId > 0 {
		query = query.Where("exam_records.course_id = ?", courseId)
	}

	err := query.Find(&records).Error
	if err != nil {
		return nil, 0, err
	}

	// 2. 从所有记录中提取错题ID并去重
	wrongQuestionIdsMap := make(map[uint]bool)
	for _, record := range records {
		// 解析JSON格式的wrong_answers字段
		var wrongIds []uint
		if record.WrongAnswers != "" {
			if err = json.Unmarshal([]byte(record.WrongAnswers), &wrongIds); err != nil {
				// 记录错误但继续处理其他记录
				continue
			}

			// 将ID添加到Map中进行去重
			for _, id := range wrongIds {
				wrongQuestionIdsMap[id] = true
			}
		}
	}

	// 3. 将去重后的ID转换为切片
	var wrongQuestionIds []uint
	for id := range wrongQuestionIdsMap {
		wrongQuestionIds = append(wrongQuestionIds, id)
	}

	// 4. 计算总数
	total := int64(len(wrongQuestionIds))

	// 如果没有错题，直接返回
	if total == 0 {
		return []WrongQuestionDetail{}, 0, nil
	}

	// 5. 分页处理
	// 确保页码和页面大小合法
	if page < 1 {
		page = 1
	}

	start := (page - 1) * pageSize
	end := start + pageSize
	if end > len(wrongQuestionIds) {
		end = len(wrongQuestionIds)
	}

	// 处理起始位置超出范围的情况
	if start >= len(wrongQuestionIds) {
		return []WrongQuestionDetail{}, total, nil
	}

	// 获取当前页的ID列表
	pageIds := wrongQuestionIds[start:end]

	// 6. 使用自定义查询获取题目信息
	type RawQuestion struct {
		ID          uint      `json:"id"`
		Type        string    `json:"type"`
		Question    string    `json:"question"`
		Options     string    `json:"options"` // 先保留为原始字符串
		Answer      string    `json:"answer"`
		Explanation string    `json:"explanation"`
		CourseID    uint      `json:"course_id"`
		UpdatedAt   time.Time `json:"updated_at"`
	}

	var rawQuestions []RawQuestion
	err = database.DB.Table("questions").
		Select("id, type, question, options, answer, explanation, course_id, updated_at").
		Where("id IN ?", pageIds).
		Where("deleted_at IS NULL").
		Find(&rawQuestions).Error

	if err != nil {
		return nil, 0, err
	}

	// 获取课程信息用于构建课程名称
	type CourseInfo struct {
		ID             uint   `json:"id"`
		CategoryLevel2 string `json:"category_level2"`
		Name           string `json:"name"`
	}

	// 收集所有需要的课程ID
	courseIdsMap := make(map[uint]bool)
	for _, q := range rawQuestions {
		courseIdsMap[q.CourseID] = true
	}

	// 将map的key转为切片
	var courseIds []uint
	for id := range courseIdsMap {
		courseIds = append(courseIds, id)
	}

	// 查询课程信息
	var courses []CourseInfo
	database.DB.Table("courses").
		Select("id, category_level2, name").
		Where("id IN ?", courseIds).
		Find(&courses)

	// 构建课程ID到课程名称的映射
	courseNameMap := make(map[uint]string)
	for _, course := range courses {
		// 拼接课程名称：CategoryLevel2-Name
		courseName := course.Name
		if course.CategoryLevel2 != "" {
			courseName = course.CategoryLevel2 + "-" + course.Name
		}
		courseNameMap[course.ID] = courseName
	}

	// 7. 转换为响应格式，手动处理options字段并过滤空选项
	var result []WrongQuestionDetail
	for _, q := range rawQuestions {
		// 处理options字段
		var options []model.QuestionOption

		// 尝试不同的方式解析options
		if q.Options != "" {
			// 1. 先尝试直接解析为QuestionOption数组
			err := json.Unmarshal([]byte(q.Options), &options)

			// 2. 如果失败，尝试解析为字符串数组
			if err != nil {
				var optStrings []string
				if err = json.Unmarshal([]byte(q.Options), &optStrings); err == nil {
					// 清空options以确保不会添加空选项
					options = []model.QuestionOption{}

					for i, optText := range optStrings {
						// 跳过空字符串
						if strings.TrimSpace(optText) == "" {
							continue
						}

						// 处理格式为 "A.选项文本" 的情况
						parts := strings.SplitN(optText, ".", 2)
						if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
							options = append(options, model.QuestionOption{
								Label: parts[0],
								Text:  strings.TrimSpace(parts[1]),
							})
						} else {
							// 如果没有分隔符，使用索引创建标签
							label := string(rune('A' + i))
							options = append(options, model.QuestionOption{
								Label: label,
								Text:  strings.TrimSpace(optText),
							})
						}
					}
				} else {
					// 3. 如果仍然失败，提供默认选项
					options = []model.QuestionOption{} // 重置选项数组

					if q.Type == "judge" {
						// 判断题默认有两个选项
						options = append(options, model.QuestionOption{Label: "A", Text: "正确"})
						options = append(options, model.QuestionOption{Label: "B", Text: "错误"})
					} else {
						// 其他题型提供4个默认选项
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
				// 解析成功，现在过滤掉空选项
				filteredOptions := []model.QuestionOption{}
				for _, opt := range options {
					if opt.Label != "" && strings.TrimSpace(opt.Text) != "" {
						filteredOptions = append(filteredOptions, model.QuestionOption{
							Label: opt.Label,
							Text:  strings.TrimSpace(opt.Text),
						})
					}
				}
				options = filteredOptions
			}
		} else {
			// 如果选项为空，提供默认选项
			options = []model.QuestionOption{} // 重置确保干净

			if q.Type == "judge" {
				options = append(options, model.QuestionOption{Label: "A", Text: "正确"})
				options = append(options, model.QuestionOption{Label: "B", Text: "错误"})
			} else {
				for i := 0; i < 4; i++ {
					label := string(rune('A' + i))
					options = append(options, model.QuestionOption{
						Label: label,
						Text:  "选项" + label,
					})
				}
			}
		}

		// 如果经过处理后没有有效选项，则添加默认选项
		if len(options) == 0 {
			if q.Type == "judge" {
				options = append(options, model.QuestionOption{Label: "A", Text: "正确"})
				options = append(options, model.QuestionOption{Label: "B", Text: "错误"})
			} else {
				for i := 0; i < 4; i++ {
					label := string(rune('A' + i))
					options = append(options, model.QuestionOption{
						Label: label,
						Text:  "选项" + label,
					})
				}
			}
		}

		// 获取课程名称，如果找不到则使用默认值
		courseName := "未知课程"
		if name, exists := courseNameMap[q.CourseID]; exists {
			courseName = name
		}

		result = append(result, WrongQuestionDetail{
			ID:          q.ID,
			Type:        q.Type,
			Question:    q.Question,
			Options:     options,
			Answer:      q.Answer,
			Explanation: q.Explanation,
			UpdatedAt:   q.UpdatedAt,
			CourseID:    q.CourseID,
			CourseName:  courseName, // 添加课程名称
		})
	}

	return result, total, nil
}

// 获取错题统计信息
func (s *PracticeService) GetWrongQuestionsStats(userId uint) ([]WrongQuestionCourse, int64, error) {
	// 1. 查询用户的所有考试记录，并关联订单表检查过期时间
	var records []model.ExamRecord
	err := database.DB.Table("exam_records").
		Joins("LEFT JOIN orders ON exam_records.course_id = orders.course_id AND exam_records.user_id = orders.user_id").
		Where("exam_records.user_id = ?", userId).
		Where("orders.status = ?", "paid").
		Where("(orders.expire_time IS NULL OR orders.expire_time > ?)", time.Now()).
		Find(&records).Error
	if err != nil {
		return nil, 0, err
	}

	// 2. 从所有记录中提取错题ID
	questionIdsByCoursemap := make(map[uint]map[uint]bool)
	for _, record := range records {
		courseId := record.CourseID

		// 确保每个课程ID都有对应的map
		if _, exists := questionIdsByCoursemap[courseId]; !exists {
			questionIdsByCoursemap[courseId] = make(map[uint]bool)
		}

		// 解析JSON格式的wrong_answers字段
		var wrongIds []uint
		if record.WrongAnswers != "" {
			if err := json.Unmarshal([]byte(record.WrongAnswers), &wrongIds); err != nil {
				continue
			}

			// 将ID添加到对应课程的Map中
			for _, id := range wrongIds {
				questionIdsByCoursemap[courseId][id] = true
			}
		}
	}

	// 如果没有错题，直接返回
	if len(questionIdsByCoursemap) == 0 {
		return []WrongQuestionCourse{}, 0, nil
	}

	// 3. 收集所有课程ID
	var courseIds []uint
	for courseId := range questionIdsByCoursemap {
		courseIds = append(courseIds, courseId)
	}

	// 4. 查询课程信息
	type CourseInfo struct {
		ID             uint   `json:"id"`
		CategoryLevel2 string `json:"category_level2"`
		Name           string `json:"name"`
	}

	var courses []CourseInfo
	database.DB.Table("courses").
		Select("id, category_level2, name").
		Where("id IN ?", courseIds).
		Find(&courses)

	// 构建课程ID到课程名称的映射
	courseNameMap := make(map[uint]string)
	for _, course := range courses {
		courseName := course.Name
		if course.CategoryLevel2 != "" {
			courseName = course.CategoryLevel2 + "-" + course.Name
		}
		courseNameMap[course.ID] = courseName
	}

	// 5. 为每个课程查询错题类型统计
	var result []WrongQuestionCourse

	for courseId, questionIdsMap := range questionIdsByCoursemap {
		// 提取当前课程的所有错题ID
		var questionIds []uint
		for id := range questionIdsMap {
			questionIds = append(questionIds, id)
		}

		// 如果没有错题，跳过
		if len(questionIds) == 0 {
			continue
		}

		// 查询这些题目的类型
		type QuestionType struct {
			ID   uint   `json:"id"`
			Type string `json:"type"`
		}

		var questionTypes []QuestionType
		database.DB.Table("questions").
			Select("id, type").
			Where("id IN ?", questionIds).
			Where("deleted_at IS NULL").
			Find(&questionTypes)

		// 统计各类型题目数量
		single := 0
		multiple := 0
		judge := 0

		for _, q := range questionTypes {
			switch q.Type {
			case "single":
				single++
			case "multiple":
				multiple++
			case "judge":
				judge++
			}
		}

		// 获取课程名称
		courseName := "未知课程"
		if name, exists := courseNameMap[courseId]; exists {
			courseName = name
		}

		// 计算总数
		totalCount := len(questionTypes)

		// 添加到结果
		result = append(result, WrongQuestionCourse{
			CourseID:   courseId,
			CourseName: courseName,
			Single:     single,
			Multiple:   multiple,
			Judge:      judge,
			Total:      totalCount,
		})
	}

	return result, int64(len(result)), nil
}

// 提交练习答案
func (s *PracticeService) Submit(userId uint, questionId uint, answer []string) (bool, error) {
	var question model.Question
	err := database.DB.First(&question, questionId).Error
	if err != nil {
		return false, err
	}

	// 检查答案是否正确
	if !compareAnswers(question.Answer, answer) {
		// 答案错误，记录到当前用户的最近一次考试记录中
		var latestRecord model.ExamRecord
		err := database.DB.Where("user_id = ?", userId).
			Where("course_id = ?", question.CourseID).
			Order("created_at DESC").
			First(&latestRecord).Error

		if err == nil {
			// 找到记录，更新wrong_answers
			var wrongIds []uint
			if latestRecord.WrongAnswers != "" {
				json.Unmarshal([]byte(latestRecord.WrongAnswers), &wrongIds)
			}

			// 检查是否已经包含此题目ID
			found := false
			for _, id := range wrongIds {
				if id == questionId {
					found = true
					break
				}
			}

			// 如果没有包含，则添加
			if !found {
				wrongIds = append(wrongIds, questionId)
				wrongIdsJson, _ := json.Marshal(wrongIds)
				database.DB.Model(&latestRecord).Update("wrong_answers", string(wrongIdsJson))
			}
		} else {
			// 没有找到记录，创建新记录
			wrongIds := []uint{questionId}
			wrongIdsJson, _ := json.Marshal(wrongIds)
			newRecord := model.ExamRecord{
				UserID:       userId,
				CourseID:     question.CourseID,
				Score:        0,
				WrongAnswers: string(wrongIdsJson),
			}
			database.DB.Create(&newRecord)
		}

		return false, nil
	}

	return true, nil
}

// 获取特定课程的所有错题（不分页）
func (s *PracticeService) GetAllWrongQuestionsByCourse(userId uint, courseId int) ([]WrongQuestionDetail, int64, error) {
	// 1. 查询用户的指定课程的所有考试记录，并关联订单表检查过期时间
	var records []model.ExamRecord
	query := database.DB.Table("exam_records").
		Joins("LEFT JOIN orders ON exam_records.course_id = orders.course_id AND exam_records.user_id = orders.user_id").
		Where("exam_records.user_id = ? AND exam_records.course_id = ?", userId, courseId).
		Where("orders.status = ?", "paid").
		Where("(orders.expire_time IS NULL OR orders.expire_time > ?)", time.Now())

	err := query.Find(&records).Error
	if err != nil {
		return nil, 0, err
	}

	// 2. 从所有记录中提取错题ID并去重
	wrongQuestionIdsMap := make(map[uint]bool)
	for _, record := range records {
		// 解析JSON格式的wrong_answers字段
		var wrongIds []uint
		if record.WrongAnswers != "" {
			if err = json.Unmarshal([]byte(record.WrongAnswers), &wrongIds); err != nil {
				// 记录错误但继续处理其他记录
				continue
			}

			// 将ID添加到Map中进行去重
			for _, id := range wrongIds {
				wrongQuestionIdsMap[id] = true
			}
		}
	}

	// 3. 将去重后的ID转换为切片
	var wrongQuestionIds []uint
	for id := range wrongQuestionIdsMap {
		wrongQuestionIds = append(wrongQuestionIds, id)
	}

	// 4. 计算总数
	total := int64(len(wrongQuestionIds))

	// 如果没有错题，直接返回
	if total == 0 {
		return []WrongQuestionDetail{}, 0, nil
	}

	// 5. 使用自定义查询获取题目信息
	type RawQuestion struct {
		ID          uint      `json:"id"`
		Type        string    `json:"type"`
		Question    string    `json:"question"`
		Options     string    `json:"options"`
		Answer      string    `json:"answer"`
		Explanation string    `json:"explanation"`
		CourseID    uint      `json:"course_id"`
		UpdatedAt   time.Time `json:"updated_at"`
	}

	var rawQuestions []RawQuestion
	err = database.DB.Table("questions").
		Select("id, type, question, options, answer, explanation, course_id, updated_at").
		Where("id IN ?", wrongQuestionIds).
		Where("deleted_at IS NULL").
		Find(&rawQuestions).Error

	if err != nil {
		return nil, 0, err
	}

	// 6. 获取课程信息
	var courseIds []uint
	courseIdMap := make(map[uint]bool)
	for _, q := range rawQuestions {
		if !courseIdMap[q.CourseID] {
			courseIds = append(courseIds, q.CourseID)
			courseIdMap[q.CourseID] = true
		}
	}

	type CourseInfo struct {
		ID             uint   `json:"id"`
		CategoryLevel2 string `json:"category_level2"`
		Name           string `json:"name"`
	}

	var courses []CourseInfo
	database.DB.Table("courses").
		Select("id, category_level2, name").
		Where("id IN ?", courseIds).
		Find(&courses)

	// 构建课程ID到课程名称的映射
	courseNameMap := make(map[uint]string)
	for _, course := range courses {
		courseName := course.Name
		if course.CategoryLevel2 != "" {
			courseName = course.CategoryLevel2 + "-" + course.Name
		}
		courseNameMap[course.ID] = courseName
	}

	// 7. 转换为响应格式，手动处理options字段并过滤空选项
	var result []WrongQuestionDetail
	for _, q := range rawQuestions {
		// 解析options字段
		var options []model.QuestionOption
		if q.Options != "" {
			// 尝试解析为QuestionOption数组
			err := json.Unmarshal([]byte(q.Options), &options)
			if err != nil {
				// 如果解析失败，尝试解析为字符串数组
				var optStrings []string
				if err := json.Unmarshal([]byte(q.Options), &optStrings); err == nil {
					for i, optStr := range optStrings {
						// 尝试从字符串中提取标签和文本
						parts := strings.SplitN(optStr, ".", 2)
						if len(parts) == 2 {
							options = append(options, model.QuestionOption{
								Label: parts[0],
								Text:  strings.TrimSpace(parts[1]),
							})
						} else {
							// 如果没有标签格式，使用默认标签
							label := string(rune('A' + i))
							options = append(options, model.QuestionOption{
								Label: label,
								Text:  strings.TrimSpace(optStr),
							})
						}
					}
				}
			}
		}

		// 过滤空选项
		var filteredOptions []model.QuestionOption
		for _, opt := range options {
			if opt.Label != "" || opt.Text != "" {
				// 确保选项的文本前后没有空格
				opt.Text = strings.TrimSpace(opt.Text)
				filteredOptions = append(filteredOptions, opt)
			}
		}

		// 确保至少有一个选项
		if len(filteredOptions) == 0 && (q.Type == "single" || q.Type == "multiple" || q.Type == "judge") {
			// 根据题目类型添加默认选项
			if q.Type == "judge" {
				filteredOptions = append(filteredOptions, model.QuestionOption{Label: "A", Text: "正确"})
				filteredOptions = append(filteredOptions, model.QuestionOption{Label: "B", Text: "错误"})
			} else {
				for i := 0; i < 4; i++ {
					label := string(rune('A' + i))
					filteredOptions = append(filteredOptions, model.QuestionOption{
						Label: label,
						Text:  "选项" + label,
					})
				}
			}
		}

		// 获取课程名称
		courseName := "未知课程"
		if name, exists := courseNameMap[q.CourseID]; exists {
			courseName = name
		}

		result = append(result, WrongQuestionDetail{
			ID:          q.ID,
			Type:        q.Type,
			Question:    q.Question,
			Options:     filteredOptions,
			Answer:      q.Answer,
			Explanation: q.Explanation,
			UpdatedAt:   q.UpdatedAt,
			CourseID:    q.CourseID,
			CourseName:  courseName,
		})
	}

	return result, total, nil
}

package service

import (
	"encoding/json"
	"errors"
	"exam-system/internal/model"
	"exam-system/internal/pkg/database"
	"fmt"
	"sort"
	"strings"
	"time"

	"gorm.io/gorm"
)

var Course = new(CourseService)

type CourseService struct{}

// 分类树节点
type CategoryNode struct {
	ID       uint           `json:"id"`
	Name     string         `json:"name"`
	Level    int            `json:"level"`
	Sort     int            `json:"sort"`
	Price    float64        `json:"price,omitempty"` // 只在三级分类中显示价格
	Children []CategoryNode `json:"children,omitempty"`
	Courses  []CourseInfo   `json:"courses,omitempty"` // 只在三级分类中包含课程
}

// 课程信息
type CourseInfo struct {
	ID          uint    `json:"id"`
	Name        string  `json:"name"`
	Cover       string  `json:"cover"`
	Price       float64 `json:"price"`
	Description string  `json:"description"`
	Purchased   bool    `json:"purchased"`   // 是否已购买
	ExpireDays  int     `json:"expire_days"` // 剩余有效期（天）
}

// 课程列表项
type CourseListItem struct {
	ID          uint    `json:"id"`
	Name        string  `json:"name"`
	Cover       string  `json:"cover"`
	Price       float64 `json:"price"`
	Description string  `json:"description"`
}

// 考试配置项
type ExamConfigItem struct {
	Type  string  `json:"type"`
	Count int     `json:"count"`
	Score float64 `json:"score"`
}

// 模拟考试全局配置
type MockExamConfig struct {
	Min   int `json:"min"`   // 考试时间（分钟）
	Count int `json:"count"` // 考试题目数量
	Score int `json:"score"` // 及格分数
}

// 课程详情
type CourseDetail struct {
	ID             uint             `json:"id"`
	Name           string           `json:"name"` // 修改为 Name + "-" + CategoryLevel2
	Cover          string           `json:"cover"`
	Price          float64          `json:"price"`
	Description    string           `json:"description"`
	QuestionStats  map[string]int   `json:"question_stats"`   // 各题型的数量
	ExamConfig     []ExamConfigItem `json:"exam_config"`      // 模拟考试配置
	MockExamConfig *MockExamConfig  `json:"mock_exam_config"` // 模拟考试全局配置
}

// 章节详情
type ChapterDetail struct {
	ID       uint            `json:"id"`
	Name     string          `json:"name"`
	Sections []SectionDetail `json:"sections"`
}

// 小节详情
type SectionDetail struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

// 模拟考试题目响应结构体（包含答案、解释和分数）
type ExamQuestionResponse struct {
	ID          uint                   `json:"id"`
	Type        string                 `json:"type"`
	Question    string                 `json:"question"`
	Options     []model.QuestionOption `json:"options"`
	Answer      string                 `json:"answer"`
	Explanation string                 `json:"explanation"`
	Score       float64                `json:"score"` // 添加分数字段
	CourseID    uint                   `json:"course_id"`
}

// 获取课程分类树
func (s *CourseService) GetCategoryTree(userId uint) ([]CategoryNode, error) {
	// 从课程表中获取所有课程，按sort字段排序
	var courses []model.Course
	if err := database.DB.Order("sort").Find(&courses).Error; err != nil {
		return nil, err
	}

	// 预处理分类排序值
	level1Sorts := make(map[string]int)
	level2Sorts := make(map[string]map[string]int)

	for _, course := range courses {
		// 处理一级分类排序
		if _, exists := level1Sorts[course.CategoryLevel1]; !exists {
			level1Sorts[course.CategoryLevel1] = course.CategorySort1
		}

		// 处理二级分类排序
		if _, exists := level2Sorts[course.CategoryLevel1]; !exists {
			level2Sorts[course.CategoryLevel1] = make(map[string]int)
		}
		if _, exists := level2Sorts[course.CategoryLevel1][course.CategoryLevel2]; !exists {
			level2Sorts[course.CategoryLevel1][course.CategoryLevel2] = course.CategorySort2
		}
	}

	// 构建分类映射
	level1Map := make(map[string]*CategoryNode)            // 一级分类映射
	level2Map := make(map[string]map[string]*CategoryNode) // 二级分类映射

	// 遍历所有课程，构建分类树
	for _, course := range courses {
		// 处理一级分类
		if _, exists := level1Map[course.CategoryLevel1]; !exists {
			level1Map[course.CategoryLevel1] = &CategoryNode{
				ID:       0, // 这里没有ID，可以用索引代替
				Name:     course.CategoryLevel1,
				Level:    1,
				Sort:     level1Sorts[course.CategoryLevel1], // 使用预处理的排序值
				Children: make([]CategoryNode, 0),
			}
		}

		// 处理二级分类
		if _, exists := level2Map[course.CategoryLevel1]; !exists {
			level2Map[course.CategoryLevel1] = make(map[string]*CategoryNode)
		}

		if _, exists := level2Map[course.CategoryLevel1][course.CategoryLevel2]; !exists {
			level2Node := &CategoryNode{
				ID:       0, // 这里没有ID，可以用索引代替
				Name:     course.CategoryLevel2,
				Level:    2,
				Sort:     level2Sorts[course.CategoryLevel1][course.CategoryLevel2], // 使用预处理的排序值
				Children: make([]CategoryNode, 0),
			}
			level2Map[course.CategoryLevel1][course.CategoryLevel2] = level2Node
		}
	}

	// 将二级分类添加到一级分类的子节点中
	for level1Name, level1Node := range level1Map {
		for _, level2Node := range level2Map[level1Name] {
			level1Node.Children = append(level1Node.Children, *level2Node)
		}
	}

	// 将课程添加到二级分类的子节点中
	// 课程ID收集，用于后续批量查询订单
	courseIDs := make([]uint, 0, len(courses))
	courseMap := make(map[uint]model.Course)
	for _, c := range courses {
		courseIDs = append(courseIDs, c.ID)
		courseMap[c.ID] = c
	}

	// 批量查询所有课程的订单状态
	purchaseStatusMap := make(map[uint]struct {
		Purchased  bool
		ExpireDays int
	})

	// 默认所有课程未购买
	for _, cid := range courseIDs {
		purchaseStatusMap[cid] = struct {
			Purchased  bool
			ExpireDays int
		}{
			Purchased:  false,
			ExpireDays: 0,
		}
	}

	// 只有登录用户才查询订单
	if userId > 0 {
		var orders []model.Order
		// 一次性查询该用户所有已支付的课程订单
		if err := database.DB.Where("user_id = ? AND course_id IN (?) AND status = ?",
			userId, courseIDs, "paid").
			Order("course_id, expire_time DESC").
			Find(&orders).Error; err != nil {
			// 查询出错，记录日志但继续处理
			fmt.Printf("查询用户订单失败: %v\n", err)
		} else {
			// 处理订单数据，为每个课程找到最新的有效订单
			courseOrderMap := make(map[uint]model.Order)
			for _, order := range orders {
				// 如果该课程还没有记录订单，或者当前订单比已记录的更新
				if _, exists := courseOrderMap[order.CourseID]; !exists {
					courseOrderMap[order.CourseID] = order
				}
			}

			// 计算每个课程的购买状态和剩余天数
			for courseID, order := range courseOrderMap {
				status := struct {
					Purchased  bool
					ExpireDays int
				}{
					Purchased:  false,
					ExpireDays: 0,
				}

				// 如果订单未过期，标记为已购买
				if order.ExpireTime != nil && order.ExpireTime.After(time.Now()) {
					status.Purchased = true
					// 计算剩余天数（向上取整）
					duration := order.ExpireTime.Sub(time.Now())
					durationDays := duration.Hours() / 24

					// 向上取整：如果有小数部分，就加1天
					status.ExpireDays = int(durationDays)
					if durationDays > float64(status.ExpireDays) {
						status.ExpireDays++
					}

					if status.ExpireDays < 0 {
						status.ExpireDays = 0
					}
				} else if order.ExpireTime == nil {
					// 如果订单没有过期时间，视为永久有效
					status.Purchased = true
					status.ExpireDays = 999999 // 表示永久有效
				}

				purchaseStatusMap[courseID] = status
			}
		}
	}

	// 使用已查询的状态信息
	for _, course := range courses {
		status := purchaseStatusMap[course.ID]

		// 创建课程节点
		courseNode := CategoryNode{
			ID:    course.ID,
			Name:  course.Name,
			Level: 3,
			Sort:  course.Sort,
			Price: course.Price,
			Courses: []CourseInfo{
				{
					ID:          course.ID,
					Name:        course.Name,
					Cover:       course.Cover,
					Price:       course.Price,
					Description: course.Description,
					Purchased:   status.Purchased,
					ExpireDays:  status.ExpireDays,
				},
			},
		}

		// 找到对应的二级分类节点
		for i := range level1Map[course.CategoryLevel1].Children {
			if level1Map[course.CategoryLevel1].Children[i].Name == course.CategoryLevel2 {
				// 添加课程到二级分类的子节点
				level1Map[course.CategoryLevel1].Children[i].Children = append(
					level1Map[course.CategoryLevel1].Children[i].Children,
					courseNode,
				)
				break
			}
		}
	}

	// 构建结果数组
	var result []CategoryNode
	for _, node := range level1Map {
		result = append(result, *node)
	}

	// 按排序字段排序一级分类
	sort.Slice(result, func(i, j int) bool {
		if result[i].Sort != result[j].Sort {
			return result[i].Sort < result[j].Sort
		}
		return result[i].Name < result[j].Name // 排序值相同时按名称排序
	})

	// 对每个一级分类的子节点排序
	for i := range result {
		// 对二级分类按排序字段排序
		sort.Slice(result[i].Children, func(j, k int) bool {
			if result[i].Children[j].Sort != result[i].Children[k].Sort {
				return result[i].Children[j].Sort < result[i].Children[k].Sort
			}
			return result[i].Children[j].Name < result[i].Children[k].Name // 排序值相同时按名称排序
		})

		// 对每个二级分类的子节点(课程)按sort字段排序
		for j := range result[i].Children {
			sort.Slice(result[i].Children[j].Children, func(k, l int) bool {
				// 优先按sort排序，sort相同时按名称排序
				if result[i].Children[j].Children[k].Sort != result[i].Children[j].Children[l].Sort {
					return result[i].Children[j].Children[k].Sort < result[i].Children[j].Children[l].Sort
				}
				return result[i].Children[j].Children[k].Name < result[i].Children[j].Children[l].Name
			})
		}
	}

	return result, nil
}

// 获取分类详情
func (s *CourseService) GetCategoryDetail(userId, categoryId uint) (*CategoryNode, error) {
	// 由于没有分类表，我们直接查询课程
	var course model.Course
	if err := database.DB.First(&course, categoryId).Error; err != nil {
		return nil, errors.New("课程不存在")
	}

	// 检查课程是否已购买
	purchased := false
	expireDays := 0

	// 只有登录用户才查询订单
	if userId > 0 {
		var order model.Order
		// 查询用户最近的已支付订单
		if err := database.DB.Where("user_id = ? AND course_id = ? AND status = ?",
			userId, course.ID, "paid").
			Order("expire_time DESC").
			First(&order).Error; err == nil {

			// 如果订单未过期，标记为已购买
			if order.ExpireTime != nil && order.ExpireTime.After(time.Now()) {
				purchased = true
				// 计算剩余天数（向上取整）
				duration := order.ExpireTime.Sub(time.Now())
				durationDays := duration.Hours() / 24

				// 向上取整：如果有小数部分，就加1天
				expireDays = int(durationDays)
				if durationDays > float64(expireDays) {
					expireDays++
				}

				if expireDays < 0 {
					expireDays = 0
				}
			} else if order.ExpireTime == nil {
				// 如果订单没有过期时间，视为永久有效
				purchased = true
				expireDays = 999999 // 表示永久有效
			}
		}
	}

	// 构建分类节点
	node := &CategoryNode{
		ID:    course.ID,
		Name:  course.Name,
		Level: 3, // 课程作为三级分类
		Price: course.Price,
		Courses: []CourseInfo{
			{
				ID:          course.ID,
				Name:        course.Name,
				Cover:       course.Cover,
				Price:       course.Price,
				Description: course.Description,
				Purchased:   purchased,  // 设置是否已购买
				ExpireDays:  expireDays, // 设置剩余有效期
			},
		},
	}

	return node, nil
}

func (s *CourseService) GetList(page, size int, courseType string) ([]CourseListItem, int64, error) {
	var courses []model.Course
	var total int64

	query := database.DB.Model(&model.Course{})
	if courseType != "" {
		query = query.Where("type = ?", courseType)
	}

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = query.Offset((page - 1) * size).Limit(size).Find(&courses).Error
	if err != nil {
		return nil, 0, err
	}

	var items []CourseListItem
	for _, course := range courses {
		items = append(items, CourseListItem{
			ID:          course.ID,
			Name:        course.Name,
			Cover:       course.Cover,
			Price:       course.Price,
			Description: course.Description,
		})
	}

	return items, total, nil
}

func (s *CourseService) GetDetail(courseId uint) (*CourseDetail, error) {
	var course model.Course
	err := database.DB.First(&course, courseId).Error
	if err != nil {
		return nil, errors.New("课程不存在")
	}

	// 1. 拼接课程名称: Name + "-" + CategoryLevel2
	courseName := course.Name
	if course.CategoryLevel2 != "" {
		courseName = course.CategoryLevel2 + "-" + course.Name
	}

	// 2. 统计各题型的数量
	type QuestionCount struct {
		Type  string
		Count int
	}

	var questionCounts []QuestionCount
	err = database.DB.Model(&model.Question{}).
		Select("type, count(*) as count").
		Where("course_id = ?", courseId).
		Group("type").
		Find(&questionCounts).Error

	if err != nil {
		return nil, err
	}

	// 将查询结果转换为map
	questionStats := make(map[string]int)
	for _, qc := range questionCounts {
		questionStats[qc.Type] = qc.Count
	}

	// 3. 解析考试配置
	var examConfig []ExamConfigItem
	courseExamConfig, err := course.GetExamConfig()
	if err != nil || len(courseExamConfig) == 0 {
		// 如果没有配置或解析失败，使用默认配置
		examConfig = []ExamConfigItem{
			{Type: "single", Count: 20, Score: 2},   // 单选题
			{Type: "multiple", Count: 10, Score: 3}, // 多选题
			{Type: "judge", Count: 10, Score: 1},    // 判断题
		}
	} else {
		// 将 model.ExamConfigItem 转换为 service.ExamConfigItem
		for _, item := range courseExamConfig {
			examConfig = append(examConfig, ExamConfigItem{
				Type:  item.Type,
				Count: item.Count,
				Score: float64(item.Score),
			})
		}
	}

	// 4. 解析模拟考试全局配置
	var mockExamConfig *MockExamConfig
	courseMockConfig, err := course.GetMockExamConfig()
	if err != nil || (courseMockConfig == model.MockExamConfig{}) {
		// 如果没有配置或解析失败，使用默认值
		mockExamConfig = &MockExamConfig{
			Min:   120,
			Count: 50,
			Score: 60,
		}
	} else {
		mockExamConfig = &MockExamConfig{
			Min:   courseMockConfig.Min,
			Count: courseMockConfig.Count,
			Score: courseMockConfig.Score,
		}
	}

	detail := &CourseDetail{
		ID:             course.ID,
		Name:           courseName,
		Cover:          course.Cover,
		Price:          course.Price,
		Description:    course.Description,
		QuestionStats:  questionStats,
		ExamConfig:     examConfig,
		MockExamConfig: mockExamConfig,
	}

	return detail, nil
}

// 随机生成模拟考试题目
func (s *CourseService) GenerateExamQuestions(courseId, userId uint) ([]ExamQuestionResponse, int, float64, float64, error) {
	// 1. 检查用户是否购买了该课程且未过期
	var order model.Order
	err := database.DB.Where("user_id = ? AND course_id = ? AND status = ?",
		userId, courseId, "paid").
		Where("expire_time IS NULL OR expire_time > ?", time.Now()).
		Order("expire_time DESC").
		First(&order).Error

	if err != nil {
		return nil, 0, 0, 0, errors.New("您尚未购买该课程或课程已过期")
	}

	// 2. 获取课程信息，特别是考试配置
	var course model.Course
	err = database.DB.First(&course, courseId).Error
	if err != nil {
		return nil, 0, 0, 0, errors.New("课程不存在")
	}

	// 3. 解析考试配置
	var examConfig []ExamConfigItem
	courseExamConfig, err := course.GetExamConfig()
	if err != nil || len(courseExamConfig) == 0 {
		// 如果没有配置或解析失败，使用默认配置
		examConfig = []ExamConfigItem{
			{Type: "single", Count: 20, Score: 2},   // 单选题
			{Type: "multiple", Count: 10, Score: 3}, // 多选题
			{Type: "judge", Count: 10, Score: 1},    // 判断题
		}
	} else {
		// 将 model.ExamConfigItem 转换为 service.ExamConfigItem
		for _, item := range courseExamConfig {
			examConfig = append(examConfig, ExamConfigItem{
				Type:  item.Type,
				Count: item.Count,
				Score: float64(item.Score),
			})
		}
	}

	// 4. 获取模拟考试时长和及格分数配置
	duration := 120  // 默认120分钟
	passScore := 0.0 // 默认及格分数，将在计算总分后设置为60%

	// 获取模拟考试配置
	courseMockConfig, err := course.GetMockExamConfig()
	if err == nil && courseMockConfig != (model.MockExamConfig{}) {
		if courseMockConfig.Min > 0 {
			duration = courseMockConfig.Min
		}
		// 记录配置中的及格分数
		if courseMockConfig.Score > 0 {
			passScore = float64(courseMockConfig.Score)
		}
	}

	// 5. 根据配置，从每种题型中随机抽取题目
	var allQuestions []ExamQuestionResponse
	var totalScore float64

	for _, config := range examConfig {
		// 使用临时结构体接收数据，以便手动处理 options 字段
		type RawQuestion struct {
			ID          uint   `json:"id"`
			Type        string `json:"type"`
			Question    string `json:"question"`
			Options     string `json:"options"` // options 作为字符串
			Answer      string `json:"answer"`
			Explanation string `json:"explanation"`
			CourseID    uint   `json:"course_id"`
			CreatedAt   time.Time
			UpdatedAt   time.Time
			DeletedAt   gorm.DeletedAt
		}

		var rawQuestions []RawQuestion
		// 查询该类型的所有题目
		err := database.DB.Table("questions").
			Where("course_id = ? AND type = ?", courseId, config.Type).
			Where("deleted_at IS NULL").
			Order("RAND()"). // MySQL特有的随机排序函数
			Limit(config.Count).
			Find(&rawQuestions).Error

		if err != nil {
			continue // 如果查询失败，跳过该题型
		}

		// 处理每个题目
		for _, q := range rawQuestions {
			// 手动解析 options 字段
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

			// 处理每个题目时，需要包含答案、解释和分数
			allQuestions = append(allQuestions, ExamQuestionResponse{
				ID:          q.ID,
				Type:        q.Type,
				Question:    q.Question,
				Options:     options,
				Answer:      q.Answer,
				Explanation: q.Explanation,
				Score:       config.Score, // 设置题目分数
				CourseID:    q.CourseID,
			})

			// 累加分数
			totalScore += config.Score
		}
	}

	// 如果没有从配置中读取到及格分数，则设置为总分的60%
	if passScore <= 0 {
		passScore = totalScore * 0.6
	}

	return allQuestions, duration, totalScore, passScore, nil
}

// 记录模拟考试结果
func (s *CourseService) SubmitExamAnswers(userId, courseId uint, score float64, wrongAnswers []uint) (*model.ExamRecord, error) {
	// 检查用户是否购买了该课程且未过期
	var order model.Order
	err := database.DB.Where("user_id = ? AND course_id = ? AND status = ?",
		userId, courseId, "paid").
		Where("expire_time IS NULL OR expire_time > ?", time.Now()).
		Order("expire_time DESC").
		First(&order).Error

	if err != nil {
		return nil, errors.New("您尚未购买该课程或课程已过期")
	}

	// 将错题ID数组转换为JSON
	wrongAnswersJSON, err := json.Marshal(wrongAnswers)
	if err != nil {
		return nil, errors.New("处理错题数据失败")
	}

	// 创建考试记录
	record := &model.ExamRecord{
		UserID:       userId,
		CourseID:     courseId,
		Score:        score,
		WrongAnswers: string(wrongAnswersJSON),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// 保存到数据库
	if err := database.DB.Create(record).Error; err != nil {
		return nil, errors.New("保存考试记录失败")
	}

	return record, nil
}

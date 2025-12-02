package admin

import (
	"encoding/json"
	"exam-system/internal/model"
	"exam-system/internal/pkg/database"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// GetCourses 获取课程列表
func GetCourses(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))
	keyword := c.Query("keyword")

	var courses []model.Course
	var total int64
	query := database.DB.Model(&model.Course{})

	// 关键字搜索
	if keyword != "" {
		query = query.Where("name LIKE ? OR category_level1 LIKE ? OR category_level2 LIKE ?",
			"%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}

	// 统计总数
	if err := query.Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "获取课程总数失败",
		})
		return
	}

	// 分页查询
	err := query.Order("sort DESC").Offset((page - 1) * size).Limit(size).Find(&courses).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "获取课程列表失败",
		})
		return
	}

	// 处理返回数据
	var courseList []gin.H
	for _, course := range courses {
		courseList = append(courseList, gin.H{
			"id":              course.ID,
			"name":            course.Name,
			"cover":           course.Cover,
			"category_level1": course.CategoryLevel1,
			"category_level2": course.CategoryLevel2,
			"price":           course.Price,
			"description":     course.Description,
			"expire_days":     course.ExpireDays,
			"sort":            course.Sort,
			"category_sort1":  course.CategorySort1,
			"category_sort2":  course.CategorySort2,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": gin.H{
			"total": total,
			"items": courseList,
		},
	})
}

// GetCourse 获取单个课程
func GetCourse(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "参数错误",
		})
		return
	}

	var course model.Course
	if err := database.DB.First(&course, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code": 404,
			"msg":  "课程不存在",
		})
		return
	}

	// 获取考试配置
	examConfig, _ := course.GetExamConfig()

	// 获取模拟考试配置
	mockExamConfig, _ := course.GetMockExamConfig()

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": gin.H{
			"id":               course.ID,
			"name":             course.Name,
			"cover":            course.Cover,
			"category_level1":  course.CategoryLevel1,
			"category_level2":  course.CategoryLevel2,
			"price":            course.Price,
			"description":      course.Description,
			"expire_days":      course.ExpireDays,
			"sort":             course.Sort,
			"category_sort1":   course.CategorySort1,
			"category_sort2":   course.CategorySort2,
			"exam_config":      examConfig,
			"mock_exam_config": mockExamConfig,
		},
	})
}

// CreateCourseRequest 创建课程请求
type CreateCourseRequest struct {
	Name           string                 `json:"name" binding:"required"`
	Cover          string                 `json:"cover" binding:"required"`
	CategoryLevel1 string                 `json:"category_level1" binding:"required"`
	CategoryLevel2 string                 `json:"category_level2" binding:"required"`
	Price          float64                `json:"price" binding:"required"`
	Description    string                 `json:"description"`
	ExpireDays     int                    `json:"expire_days"`
	Sort           int                    `json:"sort"`
	CategorySort1  int                    `json:"category_sort1"`
	CategorySort2  int                    `json:"category_sort2"`
	ExamConfig     []model.ExamConfigItem `json:"exam_config"`
	MockExamConfig model.MockExamConfig   `json:"mock_exam_config"`
}

// CreateCourse 创建课程
func CreateCourse(c *gin.Context) {
	var req CreateCourseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "参数错误",
		})
		return
	}

	course := model.Course{
		Name:           req.Name,
		Cover:          req.Cover,
		CategoryLevel1: req.CategoryLevel1,
		CategoryLevel2: req.CategoryLevel2,
		Price:          req.Price,
		Description:    req.Description,
		ExpireDays:     req.ExpireDays,
		Sort:           req.Sort,
		CategorySort1:  req.CategorySort1,
		CategorySort2:  req.CategorySort2,
	}

	// 设置考试配置
	if req.ExamConfig != nil {
		if err := course.SetExamConfig(req.ExamConfig); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"code": 400,
				"msg":  "考试配置格式错误",
			})
			return
		}
	}

	// 设置模拟考试配置
	if req.MockExamConfig != (model.MockExamConfig{}) {
		if err := course.SetMockExamConfig(req.MockExamConfig); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"code": 400,
				"msg":  "模拟考试配置格式错误",
			})
			return
		}
	}

	if err := database.DB.Create(&course).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "创建课程失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": gin.H{
			"id": course.ID,
		},
	})
}

// UpdateCourseRequest 更新课程请求
type UpdateCourseRequest struct {
	Name           string                 `json:"name"`
	Cover          string                 `json:"cover"`
	CategoryLevel1 string                 `json:"category_level1"`
	CategoryLevel2 string                 `json:"category_level2"`
	Price          float64                `json:"price"`
	Description    string                 `json:"description"`
	ExpireDays     int                    `json:"expire_days"`
	Sort           int                    `json:"sort"`
	CategorySort1  int                    `json:"category_sort1"`
	CategorySort2  int                    `json:"category_sort2"`
	ExamConfig     []model.ExamConfigItem `json:"exam_config"`
	MockExamConfig model.MockExamConfig   `json:"mock_exam_config"`
}

// UpdateCourse 更新课程
func UpdateCourse(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "参数错误",
		})
		return
	}

	var req UpdateCourseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "参数错误",
		})
		return
	}

	// 先获取课程
	var course model.Course
	if err := database.DB.First(&course, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code": 404,
			"msg":  "课程不存在",
		})
		return
	}

	// 更新基本字段
	updates := make(map[string]interface{})

	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Cover != "" {
		updates["cover"] = req.Cover
	}
	if req.CategoryLevel1 != "" {
		updates["category_level1"] = req.CategoryLevel1
	}
	if req.CategoryLevel2 != "" {
		updates["category_level2"] = req.CategoryLevel2
	}
	if req.Price >= 0 {
		updates["price"] = req.Price
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}
	if req.ExpireDays != 0 {
		updates["expire_days"] = req.ExpireDays
	}
	if req.Sort != 0 {
		updates["sort"] = req.Sort
	}
	if req.CategorySort1 != 0 {
		updates["category_sort1"] = req.CategorySort1
	}
	if req.CategorySort2 != 0 {
		updates["category_sort2"] = req.CategorySort2
	}

	// 更新考试配置
	if req.ExamConfig != nil {
		examConfigJson, err := json.Marshal(req.ExamConfig)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"code": 400,
				"msg":  "考试配置格式错误",
			})
			return
		}
		updates["exam_config"] = string(examConfigJson)
	}

	// 更新模拟考试配置
	if req.MockExamConfig != (model.MockExamConfig{}) {
		mockExamConfigJson, err := json.Marshal(req.MockExamConfig)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"code": 400,
				"msg":  "模拟考试配置格式错误",
			})
			return
		}
		updates["mock_exam_config"] = string(mockExamConfigJson)
	}

	result := database.DB.Model(&model.Course{}).Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "更新课程失败",
		})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"code": 404,
			"msg":  "课程不存在",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "更新成功",
	})
}

// DeleteCourse 删除课程
func DeleteCourse(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "参数错误",
		})
		return
	}

	// 检查是否有关联的订单
	var orderCount int64
	if err := database.DB.Model(&model.Order{}).Where("course_id = ?", id).Count(&orderCount).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "检查关联订单失败",
		})
		return
	}

	if orderCount > 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "该课程已有订单关联，无法删除",
		})
		return
	}

	// 删除课程
	result := database.DB.Delete(&model.Course{}, id)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "删除课程失败",
		})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"code": 404,
			"msg":  "课程不存在",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "删除成功",
	})
}

package api

import (
	"exam-system/internal/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func GetCourseList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	courseType := c.Query("type")

	courses, total, err := service.Course.GetList(page, size, courseType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": gin.H{
			"total": total,
			"list":  courses,
		},
	})
}

func GetCourseDetail(c *gin.Context) {
	courseId, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "参数错误",
		})
		return
	}

	course, err := service.Course.GetDetail(uint(courseId))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": course,
	})
}

// 获取课程模拟考试题目
func GetCourseExam(c *gin.Context) {
	courseId, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "参数错误",
		})
		return
	}

	// 获取用户ID
	userId := c.GetUint("userId")
	if userId == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code": 401,
			"msg":  "请先登录",
		})
		return
	}

	// 随机生成模拟考试题目
	examQuestions, duration, totalScore, passScore, err := service.Course.GenerateExamQuestions(uint(courseId), userId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": gin.H{
			"questions":   examQuestions,
			"duration":    duration,   // 考试时长（分钟）
			"total_score": totalScore, // 总分
			"pass_score":  passScore,  // 及格分数，从配置中获取
		},
	})
}

func GetCourseCategories(c *gin.Context) {
	// 获取用户ID，未登录用户为0
	userId := c.GetUint("userId")

	categories, err := service.Course.GetCategoryTree(userId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": categories,
	})
}

func GetCategoryDetail(c *gin.Context) {
	categoryId, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "参数错误",
		})
		return
	}

	// 获取用户ID，未登录用户为0
	userId := c.GetUint("userId")

	category, err := service.Course.GetCategoryDetail(userId, uint(categoryId))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": category,
	})
}

// 提交模拟考试答案
func SubmitCourseExam(c *gin.Context) {
	courseId, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "参数错误",
		})
		return
	}

	// 解析请求体
	var req struct {
		UserID       uint    `json:"user_id"`
		CourseID     uint    `json:"course_id"`
		Score        float64 `json:"score"`
		WrongAnswers []uint  `json:"wrong_answers"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "请求参数错误: " + err.Error(),
		})
		return
	}

	// 获取用户ID
	userId := c.GetUint("userId")
	if userId == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code": 401,
			"msg":  "请先登录",
		})
		return
	}

	// 确保请求中的courseId与URL中的一致
	if req.CourseID != uint(courseId) {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "请求参数错误: course_id不匹配",
		})
		return
	}

	// 记录考试结果
	record, err := service.Course.SubmitExamAnswers(userId, uint(courseId), req.Score, req.WrongAnswers)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  err.Error(),
		})
		return
	}

	// 判断是否通过考试
	passScore := 60.0 // 默认及格分数为60分

	// 获取课程信息以确定实际及格分数
	course, err := service.Course.GetDetail(uint(courseId))
	if err == nil && course.MockExamConfig != nil && course.MockExamConfig.Score > 0 {
		passScore = float64(course.MockExamConfig.Score)
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": gin.H{
			"id":     record.ID,
			"score":  record.Score,
			"passed": record.Score >= passScore,
		},
	})
}

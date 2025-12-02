package api

import (
	"exam-system/internal/service"
	"exam-system/internal/types"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// 获取错题统计信息
func GetWrongQuestionsStats(c *gin.Context) {
	userId := c.GetUint("userId")
	stats, total, err := service.Practice.GetWrongQuestionsStats(userId)
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
			"courses": stats,
			"total":   total,
		},
	})
}

// 获取所有错题列表（详细信息）
func GetWrongQuestions(c *gin.Context) {
	userId := c.GetUint("userId")
	// 从查询参数获取课程ID（可选）
	courseId, _ := strconv.Atoi(c.Query("course_id"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	questions, total, err := service.Practice.GetWrongQuestions(userId, courseId, page, pageSize)
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
			"questions": questions,
			"total":     total,
		},
	})
}

// 根据课程ID获取错题列表（不分页，返回全部）
func GetWrongQuestionsByCourse(c *gin.Context) {
	userId := c.GetUint("userId")
	// 从路径参数获取课程ID
	courseId, err := strconv.Atoi(c.Param("course_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "无效的课程ID",
		})
		return
	}

	// 直接获取全部错题，不使用分页
	questions, total, err := service.Practice.GetAllWrongQuestionsByCourse(userId, courseId)
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
			"questions": questions,
			"total":     total,
		},
	})
}

// 提交练习答案
func SubmitPractice(c *gin.Context) {
	userId := c.GetUint("userId")
	var req types.SubmitPracticeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "参数错误: " + err.Error(),
		})
		return
	}

	correct, err := service.Practice.Submit(userId, req.QuestionID, req.Answer)
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
			"correct": correct,
		},
	})
}

package api

import (
	"exam-system/internal/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// 获取课程题目
func GetCourseQuestions(c *gin.Context) {
	// 获取参数
	courseIdStr := c.Param("course_id")
	courseId, err := strconv.ParseUint(courseIdStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "课程ID格式错误",
		})
		return
	}

	// 获取题目类型（single、multiple、judge）
	questionType := c.DefaultQuery("type", "")

	// 获取用户ID
	userId := c.GetUint("userId")
	if userId == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code": 401,
			"msg":  "请先登录",
		})
		return
	}

	// 调用服务获取题目
	questions, err := service.Question.GetQuestionsByCourse(userId, uint(courseId), questionType)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{
			"code": 403,
			"msg":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": questions,
	})
}

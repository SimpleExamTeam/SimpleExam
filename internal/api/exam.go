package api

import (
	"exam-system/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

// 获取用户的所有考试结果
func GetExamResults(c *gin.Context) {
	// 获取用户ID
	userId := c.GetUint("userId")

	// 调用服务层获取考试结果
	results, err := service.Exam.GetAllResults(userId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": results,
	})
}

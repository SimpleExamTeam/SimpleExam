package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// SimpleHealthCheck 简单健康检查
// 用于 Docker 健康检查和负载均衡器
func SimpleHealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
	})
}

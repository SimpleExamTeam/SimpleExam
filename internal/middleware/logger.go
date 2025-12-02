package middleware

import (
	"exam-system/internal/pkg/logger"
	"time"

	"github.com/gin-gonic/gin"
)

// Logger 自定义日志中间件
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 开始时间
		startTime := time.Now()

		// 处理请求
		c.Next()

		// 结束时间
		endTime := time.Now()

		// 执行时间
		latencyTime := endTime.Sub(startTime)

		// 请求方式
		reqMethod := c.Request.Method

		// 请求路由
		reqUri := c.Request.RequestURI

		// 状态码
		statusCode := c.Writer.Status()

		// 请求IP
		clientIP := c.ClientIP()

		// 用户代理
		userAgent := c.Request.UserAgent()

		// 记录请求日志
		if statusCode >= 500 {
			logger.Errorf("[%s] %s %s %d %v \"%s\" - Internal Server Error",
				clientIP,
				reqMethod,
				reqUri,
				statusCode,
				latencyTime,
				userAgent)
		} else if statusCode >= 400 {
			logger.Warnf("[%s] %s %s %d %v \"%s\" - Client Error",
				clientIP,
				reqMethod,
				reqUri,
				statusCode,
				latencyTime,
				userAgent)
		} else {
			logger.Infof("[%s] %s %s %d %v \"%s\"",
				clientIP,
				reqMethod,
				reqUri,
				statusCode,
				latencyTime,
				userAgent)
		}
	}
}

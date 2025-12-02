package middleware

import (
	"exam-system/internal/model"
	"exam-system/internal/pkg/database"
	"net/http"

	"github.com/gin-gonic/gin"
)

// AdminAuth 管理员认证中间件
func AdminAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从上下文中获取用户ID（JWT中间件已经验证过token并设置了user_id）
		userId, exists := c.Get("userId")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code": 401,
				"msg":  "未登录",
			})
			c.Abort()
			return
		}

		// 查询用户
		var user model.User
		if err := database.DB.Unscoped().First(&user, userId).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code": 401,
				"msg":  "用户不存在或已被删除",
			})
			c.Abort()
			return
		}

		// 检查用户是否已被删除
		if !user.DeletedAt.Time.IsZero() {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code": 401,
				"msg":  "用户已被删除",
			})
			c.Abort()
			return
		}

		// 验证是否是管理员
		if !user.IsAdmin {
			c.JSON(http.StatusForbidden, gin.H{
				"code": 403,
				"msg":  "无管理员权限",
			})
			c.Abort()
			return
		}

		// 将用户信息保存到上下文
		c.Set("admin_user", user)
		c.Next()
	}
}

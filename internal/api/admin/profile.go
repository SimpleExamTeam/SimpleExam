package admin

import (
	"exam-system/internal/model"
	"exam-system/internal/pkg/database"
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// UpdateProfileRequest 更新个人信息请求
type UpdateProfileRequest struct {
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
	Sex      *int   `json:"sex"`
	Password string `json:"password"`
}

// GetAdminProfile 获取当前管理员个人信息
func GetAdminProfile(c *gin.Context) {
	// 从上下文中获取管理员ID
	adminID, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code": 401,
			"msg":  "未登录或登录已过期",
		})
		return
	}

	// 查询管理员信息
	var user model.User
	if err := database.DB.First(&user, adminID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "获取管理员信息失败",
		})
		return
	}

	// 返回管理员信息，不包含敏感字段
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": gin.H{
			"id":         user.ID,
			"username":   user.Username,
			"nickname":   user.Nickname,
			"avatar":     user.Avatar,
			"sex":        user.Sex,
			"is_admin":   user.IsAdmin,
			"created_at": user.CreatedAt,
		},
	})
}

// UpdateAdminProfile 更新当前管理员个人信息
func UpdateAdminProfile(c *gin.Context) {
	// 从上下文中获取管理员ID
	adminID, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code": 401,
			"msg":  "未登录或登录已过期",
		})
		return
	}

	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "参数错误",
		})
		return
	}

	// 查询管理员信息
	var user model.User
	if err := database.DB.First(&user, adminID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "获取管理员信息失败",
		})
		return
	}

	// 构建更新内容
	updates := make(map[string]interface{})
	if req.Nickname != "" {
		updates["nickname"] = req.Nickname
	}
	if req.Avatar != "" {
		updates["avatar"] = req.Avatar
	}
	if req.Sex != nil {
		updates["sex"] = *req.Sex
	}

	// 如果提供了新密码，更新密码
	if req.Password != "" {
		// 使用bcrypt对密码进行加密
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code": 500,
				"msg":  "密码加密失败",
			})
			return
		}
		updates["password"] = string(hashedPassword)
	}

	// 如果没有需要更新的内容
	if len(updates) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "没有需要更新的内容",
		})
		return
	}

	// 更新管理员信息
	if err := database.DB.Model(&user).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "更新管理员信息失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "更新成功",
	})
}

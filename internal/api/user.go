package api

import (
	"exam-system/internal/model"
	"exam-system/internal/pkg/database"
	"exam-system/internal/service"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type UpdateProfileRequest struct {
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
	OpenID   string `json:"open_id"`
}

func GetUserProfile(c *gin.Context) {
	userId := c.GetUint("userId")
	user, err := service.User.GetProfile(userId)
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
			"id":       user.ID,
			"username": user.Username,
			"nickname": user.Nickname,
			"avatar":   user.Avatar,
			"open_id":  user.OpenID,
		},
	})
}

func UpdateUserProfile(c *gin.Context) {
	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "参数错误",
		})
		return
	}

	userId := c.GetUint("userId")
	err := service.User.UpdateProfile(userId, req.Nickname, req.Avatar)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "更新成功",
	})
}

// GetTokenExpireTime 获取用户token的过期时间
func GetTokenExpireTime(c *gin.Context) {
	// 获取当前用户ID
	userId := c.GetUint("userId")

	// 获取token过期时间
	expireTime, err := service.User.GetTokenExpireTime(userId)
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
			"expire_time":           expireTime,
			"expire_time_formatted": time.Unix(expireTime, 0).Format("2006-01-02 15:04:05"),
		},
	})
}

// CreateFeedbackRequest 创建反馈请求
type CreateFeedbackRequest struct {
	FeedbackContent string `json:"feedback_content" binding:"required"`
}

// GetUserFeedbacks 获取当前用户的反馈列表
func GetUserFeedbacks(c *gin.Context) {
	// 从上下文中获取用户ID
	userId := c.GetUint("userId")

	// 获取分页参数
	pageStr := c.DefaultQuery("page", "1")
	sizeStr := c.DefaultQuery("size", "10")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	size, err := strconv.Atoi(sizeStr)
	if err != nil || size < 1 || size > 100 {
		size = 10
	}

	// 查询用户的反馈列表
	var feedbacks []model.UserFeedback
	var total int64

	db := database.DB.Model(&model.UserFeedback{}).Where("user_id = ?", userId)

	// 获取总数
	if err := db.Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "获取反馈总数失败",
		})
		return
	}

	// 获取反馈列表
	if err := db.Order("created_at DESC").
		Offset((page - 1) * size).
		Limit(size).
		Find(&feedbacks).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "获取反馈列表失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": gin.H{
			"total": total,
			"items": feedbacks,
		},
	})
}

// CreateUserFeedback 创建用户反馈
func CreateUserFeedback(c *gin.Context) {
	// 从上下文中获取用户ID
	userId := c.GetUint("userId")

	var req CreateFeedbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "参数错误",
		})
		return
	}

	// 创建反馈
	feedback := model.UserFeedback{
		UserID:          userId,
		FeedbackContent: req.FeedbackContent,
		Status:          0, // 默认为未确认状态
	}

	if err := database.DB.Create(&feedback).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "创建反馈失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "反馈提交成功",
		"data": feedback,
	})
}

package admin

import (
	"exam-system/internal/model"
	"exam-system/internal/pkg/database"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// FeedbackQuery 反馈查询参数
type FeedbackQuery struct {
	Page      int    `form:"page,default=1"`
	Size      int    `form:"size,default=10"`
	Username  string `form:"username"`
	Status    *int   `form:"status"`
	StartTime string `form:"start_time"`
	EndTime   string `form:"end_time"`
}

// UpdateFeedbackRequest 更新反馈请求
type UpdateFeedbackRequest struct {
	Status       *int   `json:"status"`
	ReplyContent string `json:"reply_content"`
}

// GetAllFeedbacks 获取所有用户反馈
func GetAllFeedbacks(c *gin.Context) {
	var query FeedbackQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "参数错误",
		})
		return
	}

	// 设置默认值
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.Size <= 0 {
		query.Size = 10
	}

	db := database.DB.Model(&model.UserFeedback{}).Preload("User")

	// 构建查询条件
	if query.Username != "" {
		db = db.Joins("JOIN users ON user_feedbacks.user_id = users.id").
			Where("users.username LIKE ? OR users.nickname LIKE ?", "%"+query.Username+"%", "%"+query.Username+"%")
	}

	// 只有当显式提供status参数时才过滤
	if query.Status != nil {
		db = db.Where("status = ?", *query.Status)
	}

	if query.StartTime != "" {
		db = db.Where("created_at >= ?", query.StartTime)
	}
	if query.EndTime != "" {
		db = db.Where("created_at <= ?", query.EndTime)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "获取反馈总数失败",
		})
		return
	}

	var feedbacks []model.UserFeedback
	if err := db.Order("created_at DESC").
		Offset((query.Page - 1) * query.Size).
		Limit(query.Size).
		Find(&feedbacks).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "获取反馈列表失败",
		})
		return
	}

	// 处理返回数据
	feedbackList := make([]gin.H, 0)
	for _, feedback := range feedbacks {
		// 构建用户信息
		userInfo := gin.H{
			"id":       feedback.User.ID,
			"username": feedback.User.Username,
			"nickname": feedback.User.Nickname,
		}

		feedbackList = append(feedbackList, gin.H{
			"id":               feedback.ID,
			"user":             userInfo,
			"feedback_content": feedback.FeedbackContent,
			"status":           feedback.Status,
			"reply_content":    feedback.ReplyContent,
			"created_at":       feedback.CreatedAt,
			"updated_at":       feedback.UpdatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": gin.H{
			"total": total,
			"items": feedbackList,
		},
	})
}

// GetFeedback 获取单个反馈详情
func GetFeedback(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "参数错误",
		})
		return
	}

	var feedback model.UserFeedback
	if err := database.DB.Preload("User").First(&feedback, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code": 404,
			"msg":  "反馈不存在",
		})
		return
	}

	// 构建用户信息
	userInfo := gin.H{
		"id":       feedback.User.ID,
		"username": feedback.User.Username,
		"nickname": feedback.User.Nickname,
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": gin.H{
			"id":               feedback.ID,
			"user":             userInfo,
			"feedback_content": feedback.FeedbackContent,
			"status":           feedback.Status,
			"reply_content":    feedback.ReplyContent,
			"created_at":       feedback.CreatedAt,
			"updated_at":       feedback.UpdatedAt,
		},
	})
}

// UpdateFeedback 更新反馈状态和回复
func UpdateFeedback(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "参数错误",
		})
		return
	}

	var req UpdateFeedbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "参数错误",
		})
		return
	}

	// 构建更新内容
	updates := make(map[string]interface{})
	if req.Status != nil {
		updates["status"] = *req.Status
	}
	if req.ReplyContent != "" {
		updates["reply_content"] = req.ReplyContent
	}

	// 如果没有需要更新的内容
	if len(updates) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "没有需要更新的内容",
		})
		return
	}

	// 更新反馈
	result := database.DB.Model(&model.UserFeedback{}).Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "更新反馈失败",
		})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"code": 404,
			"msg":  "反馈不存在",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "更新成功",
	})
}

// DeleteFeedback 删除反馈
func DeleteFeedback(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "参数错误",
		})
		return
	}

	result := database.DB.Delete(&model.UserFeedback{}, id)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "删除反馈失败",
		})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"code": 404,
			"msg":  "反馈不存在",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "删除成功",
	})
}

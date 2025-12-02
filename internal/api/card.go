package api

import (
	"exam-system/internal/model"
	"exam-system/internal/pkg/database"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// RedeemCardRequest 兑换卡券请求
type RedeemCardRequest struct {
	CardNo   string `json:"card_no" binding:"required"`
	CourseID uint   `json:"course_id" binding:"required"`
}

// RedeemCard 兑换卡券
func RedeemCard(c *gin.Context) {
	// 获取用户ID
	userID, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code": 401,
			"msg":  "未登录或登录已过期",
		})
		return
	}

	var req RedeemCardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "参数错误",
		})
		return
	}

	// 验证卡券是否存在
	var card model.Card
	if err := database.DB.Where("card_no = ?", req.CardNo).First(&card).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "卡券不存在或已被使用",
		})
		return
	}

	// 验证卡券是否过期
	expireTime := card.CreatedAt.AddDate(0, 0, card.ExpireDays)
	if time.Now().After(expireTime) {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "卡券已过期",
		})
		return
	}

	// 验证卡券是否已被全部兑换
	if card.Used >= card.Total {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "卡券已被全部兑换",
		})
		return
	}

	// 验证课程是否存在
	var course model.Course
	if err := database.DB.First(&course, req.CourseID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "课程不存在",
		})
		return
	}

	// 验证卡券是否可以兑换指定课程
	if card.CourseID != nil && *card.CourseID > 0 && *card.CourseID != req.CourseID {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "该卡券只能兑换指定课程",
		})
		return
	}

	// 验证用户是否已经购买过该课程
	var orderCount int64
	database.DB.Model(&model.Order{}).
		Where("user_id = ? AND course_id = ? AND status = ?", userID, req.CourseID, "paid").
		Count(&orderCount)
	if orderCount > 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "您已经购买过该课程",
		})
		return
	}

	// 开始事务
	tx := database.DB.Begin()

	// 创建订单
	now := time.Now()
	orderExpireTime := now.AddDate(0, 0, card.ExpireDays)
	orderNo := now.Format("20060102150405") + strconv.Itoa(int(userID.(uint)))
	order := model.Order{
		OrderNo:     orderNo,
		UserID:      userID.(uint),
		CourseID:    req.CourseID,
		Amount:      card.Amount,
		Status:      "paid", // 直接设置为已支付
		PaymentType: "card", // 支付方式为卡券
		PayTime:     &now,
		ExpireTime:  &orderExpireTime,
	}

	if err := tx.Create(&order).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "创建订单失败",
		})
		return
	}

	// 创建卡券兑换记录
	record := model.CardRecord{
		CardID:   card.ID,
		CardNo:   card.CardNo,
		UserID:   userID.(uint),
		OrderID:  order.ID,
		OrderNo:  order.OrderNo,
		CourseID: req.CourseID,
		Amount:   card.Amount,
	}

	if err := tx.Create(&record).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "创建兑换记录失败",
		})
		return
	}

	// 更新卡券使用次数
	if err := tx.Model(&card).Update("used", gorm.Expr("used + 1")).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "更新卡券使用次数失败",
		})
		return
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "提交事务失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "兑换成功",
		"data": gin.H{
			"order_id":  order.ID,
			"order_no":  order.OrderNo,
			"course_id": order.CourseID,
			"amount":    order.Amount,
			"status":    order.Status,
		},
	})
}

package admin

import (
	"exam-system/internal/model"
	"exam-system/internal/pkg/database"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// OrderQuery 订单查询参数
type OrderQuery struct {
	Page        int    `form:"page,default=1"`
	Size        int    `form:"size,default=10"`
	OrderNo     string `form:"order_no"`
	Username    string `form:"username"`
	UserID      uint   `form:"user_id"`
	Status      string `form:"status"`
	PaymentType string `form:"payment_type"`
	StartTime   string `form:"start_time"`
	EndTime     string `form:"end_time"`
}

// GetOrders 获取订单列表
func GetOrders(c *gin.Context) {
	var query OrderQuery
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

	db := database.DB.Model(&model.Order{}).Debug() // 添加 Debug 模式打印 SQL

	// 构建查询条件
	if query.OrderNo != "" {
		db = db.Where("order_no LIKE ?", "%"+query.OrderNo+"%")
	}
	if query.Status != "" {
		db = db.Where("status = ?", query.Status)
	}
	if query.PaymentType != "" {
		db = db.Where("payment_type = ?", query.PaymentType)
	}
	// 按用户ID查询
	if query.UserID > 0 {
		db = db.Where("user_id = ?", query.UserID)
	}

	// 时间范围查询
	if query.StartTime != "" {
		startTime, err := time.ParseInLocation("2006-01-02 15:04:05", query.StartTime, time.Local)
		if err == nil {
			db = db.Where("orders.created_at >= ?", startTime)
		} else {
			// 打印时间解析错误
			c.Set("time_parse_error_start", err.Error())
		}
	}
	if query.EndTime != "" {
		endTime, err := time.ParseInLocation("2006-01-02 15:04:05", query.EndTime, time.Local)
		if err == nil {
			db = db.Where("orders.created_at <= ?", endTime)
		} else {
			// 打印时间解析错误
			c.Set("time_parse_error_end", err.Error())
		}
	}

	// 用户名查询（需要连接用户表）
	if query.Username != "" {
		db = db.Joins("JOIN users ON orders.user_id = users.id").
			Where("users.username LIKE ? OR users.nickname LIKE ?",
				"%"+query.Username+"%", "%"+query.Username+"%")
	}

	var total int64
	countDb := db.Session(&gorm.Session{}) // 创建新会话用于计数，避免影响主查询
	if err := countDb.Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "获取订单总数失败",
		})
		return
	}

	var orders []model.Order
	err := db.Preload("User").
		Order("created_at DESC").
		Offset((query.Page - 1) * query.Size).
		Limit(query.Size).
		Find(&orders).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "获取订单列表失败",
		})
		return
	}

	// 处理返回数据
	orderList := make([]gin.H, 0)
	for _, order := range orders {
		// 查询课程信息
		var course model.Course
		courseName := "未知课程"
		if err := database.DB.First(&course, order.CourseID).Error; err == nil {
			courseName = course.CategoryLevel2 + "-" + course.Name
		}

		// 构建用户信息
		userInfo := gin.H{
			"id":       order.User.ID,
			"username": order.User.Username,
			"nickname": order.User.Nickname,
		}

		orderList = append(orderList, gin.H{
			"id":           order.ID,
			"order_no":     order.OrderNo,
			"user":         userInfo,
			"course_id":    order.CourseID,
			"course_name":  courseName,
			"amount":       order.Amount,
			"status":       order.Status,
			"payment_type": order.PaymentType,
			"pay_time":     order.PayTime,
			"expire_time":  order.ExpireTime,
			"created_at":   order.CreatedAt,
		})
	}

	// 在响应中添加查询参数，便于排查
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": gin.H{
			"total": total,
			"items": orderList,
			"query": query, // 返回查询参数
		},
	})
}

// GetOrder 获取单个订单
func GetOrder(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "参数错误",
		})
		return
	}

	var order model.Order
	if err := database.DB.Preload("User").First(&order, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code": 404,
			"msg":  "订单不存在",
		})
		return
	}

	// 查询课程信息
	var course model.Course
	courseName := "未知课程"
	if err = database.DB.First(&course, order.CourseID).Error; err == nil {
		courseName = course.CategoryLevel2 + "-" + course.Name
	}

	// 构建用户信息
	userInfo := gin.H{
		"id":       order.User.ID,
		"username": order.User.Username,
		"nickname": order.User.Nickname,
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": gin.H{
			"id":           order.ID,
			"order_no":     order.OrderNo,
			"user":         userInfo,
			"course_id":    order.CourseID,
			"course_name":  courseName,
			"amount":       order.Amount,
			"status":       order.Status,
			"payment_type": order.PaymentType,
			"pay_time":     order.PayTime,
			"expire_time":  order.ExpireTime,
			"created_at":   order.CreatedAt,
		},
	})
}

// CreateOrderRequest 创建订单请求
type CreateOrderRequest struct {
	UserID   uint    `json:"user_id" binding:"required"`
	CourseID uint    `json:"course_id" binding:"required"`
	Amount   float64 `json:"amount" binding:"required"`
}

// CreateOrder 创建订单
func CreateOrder(c *gin.Context) {
	var req CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "参数错误",
		})
		return
	}

	// 检查用户是否存在
	var user model.User
	if err := database.DB.First(&user, req.UserID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "用户不存在",
		})
		return
	}

	// 检查课程是否存在
	var course model.Course
	if err := database.DB.First(&course, req.CourseID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "课程不存在",
		})
		return
	}

	// 创建订单
	order := model.Order{
		OrderNo:  time.Now().Format("20060102150405") + strconv.Itoa(int(req.UserID)),
		UserID:   req.UserID,
		CourseID: req.CourseID,
		Amount:   req.Amount,
		Status:   "pending",
	}

	if err := database.DB.Create(&order).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "创建订单失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": gin.H{
			"id":       order.ID,
			"order_no": order.OrderNo,
		},
	})
}

// UpdateOrderRequest 更新订单请求
type UpdateOrderRequest struct {
	Status      string  `json:"status"`
	PaymentType string  `json:"payment_type"`
	PayTime     string  `json:"pay_time"`
	ExpireTime  string  `json:"expire_time"`
	Amount      float64 `json:"amount"`
}

// UpdateOrder 更新订单
func UpdateOrder(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "参数错误",
		})
		return
	}

	var req UpdateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "参数错误",
		})
		return
	}

	updates := make(map[string]interface{})

	if req.Status != "" {
		updates["status"] = req.Status
	}
	if req.PaymentType != "" {
		updates["payment_type"] = req.PaymentType
	}
	if req.PayTime != "" {
		payTime, err := time.ParseInLocation("2006-01-02 15:04:05.999", req.PayTime, time.Local)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"code": 400,
				"msg":  "支付时间格式错误",
			})
			return
		}
		updates["pay_time"] = payTime
	}
	if req.ExpireTime != "" {
		expireTime, err := time.ParseInLocation("2006-01-02 15:04:05.999", req.ExpireTime, time.Local)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"code": 400,
				"msg":  "过期时间格式错误",
			})
			return
		}
		updates["expire_time"] = expireTime
	}
	if req.Amount > 0 {
		updates["amount"] = req.Amount
	}

	result := database.DB.Model(&model.Order{}).Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "更新订单失败",
		})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"code": 404,
			"msg":  "订单不存在",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "更新成功",
	})
}

// DeleteOrder 删除订单
func DeleteOrder(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "参数错误",
		})
		return
	}

	result := database.DB.Delete(&model.Order{}, id)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "删除订单失败",
		})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"code": 404,
			"msg":  "订单不存在",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "删除成功",
	})
}

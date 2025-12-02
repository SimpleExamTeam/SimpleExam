package service

import (
	"errors"
	"exam-system/internal/model"
	"exam-system/internal/pkg/database"
	"fmt"
	"math/rand"
	"time"
)

var Order = new(OrderService)

type OrderService struct{}

// 订单列表项
type OrderListItem struct {
	ID         uint      `json:"id"`
	OrderNo    string    `json:"order_no"`
	CourseID   uint      `json:"course_id"`
	CourseName string    `json:"course_name"`
	Price      float64   `json:"price"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
}

// 订单详情
type OrderDetail struct {
	ID          uint       `json:"id"`
	OrderNo     string     `json:"order_no"`
	CourseID    uint       `json:"course_id"`
	CourseName  string     `json:"course_name"`
	Price       float64    `json:"price"`
	Status      string     `json:"status"`
	PaymentType string     `json:"payment_type"`
	PayTime     *time.Time `json:"pay_time"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

func (s *OrderService) Create(userId, courseId uint) (*model.Order, error) {
	// 检查课程是否存在
	var course model.Course
	err := database.DB.First(&course, courseId).Error
	if err != nil {
		return nil, errors.New("课程不存在")
	}

	// 检查是否已购买且未过期
	var count int64
	database.DB.Model(&model.Order{}).
		Where("user_id = ? AND course_id = ? AND status = ?", userId, courseId, "paid").
		Where("(expire_time IS NULL OR expire_time > ?)", time.Now()).
		Count(&count)
	if count > 0 {
		return nil, errors.New("已购买该课程")
	}

	// 计算过期时间
	var expireTime *time.Time
	if course.ExpireDays > 0 {
		t := time.Now().AddDate(0, 0, course.ExpireDays)
		expireTime = &t
	}

	// 创建订单时只设置必要的字段，不设置 pay_time
	order := &model.Order{
		OrderNo:    generateOrderNo(),
		UserID:     userId,
		CourseID:   courseId,
		Amount:     course.Price,
		Status:     "pending",
		ExpireTime: expireTime, // 设置过期时间
		// 不设置 PaymentType 和 PayTime，让它们保持 NULL
	}

	err = database.DB.Create(order).Error
	if err != nil {
		return nil, err
	}

	return order, nil
}

func (s *OrderService) GetList(userId uint) ([]map[string]interface{}, error) {
	var orders []model.Order
	err := database.DB.Where("user_id = ?", userId).Find(&orders).Error
	if err != nil {
		return nil, err
	}

	var result []map[string]interface{}
	for _, order := range orders {
		// 查询课程信息
		var course model.Course
		courseName := "未知课程" // 默认值
		if err := database.DB.First(&course, order.CourseID).Error; err == nil {
			// 拼接课程名称：CategoryLevel2-Name
			courseName = course.CategoryLevel2 + "-" + course.Name
		}

		// 构建包含 CourseName 的响应，使用小写加下划线的字段名
		result = append(result, map[string]interface{}{
			"id":           order.ID,
			"order_no":     order.OrderNo,
			"user_id":      order.UserID,
			"course_id":    order.CourseID,
			"course_name":  courseName, // 添加课程名称
			"amount":       order.Amount,
			"status":       order.Status,
			"payment_type": order.PaymentType,
			"pay_time":     order.PayTime,
			"created_at":   order.CreatedAt,
			"updated_at":   order.UpdatedAt,
			"deleted_at":   order.DeletedAt,
		})
	}

	return result, nil
}

func (s *OrderService) GetDetail(userId, orderId uint) (map[string]interface{}, error) {
	var order model.Order
	err := database.DB.Where("id = ? AND user_id = ?", orderId, userId).First(&order).Error
	if err != nil {
		return nil, errors.New("订单不存在")
	}

	// 查询课程信息
	var course model.Course
	courseName := "未知课程" // 默认值
	if err := database.DB.First(&course, order.CourseID).Error; err == nil {
		// 拼接课程名称：CategoryLevel2-Name
		courseName = course.CategoryLevel2 + "-" + course.Name
	}

	// 构建包含 CourseName 的响应，使用小写加下划线的字段名
	result := map[string]interface{}{
		"id":           order.ID,
		"order_no":     order.OrderNo,
		"user_id":      order.UserID,
		"course_id":    order.CourseID,
		"course_name":  courseName, // 添加课程名称
		"amount":       order.Amount,
		"status":       order.Status,
		"payment_type": order.PaymentType,
		"pay_time":     order.PayTime,
		"created_at":   order.CreatedAt,
		"updated_at":   order.UpdatedAt,
		"deleted_at":   order.DeletedAt,
	}

	return result, nil
}

// 生成订单号
func generateOrderNo() string {
	return time.Now().Format("20060102150405") + fmt.Sprintf("%06d", rand.Intn(1000000))
}

// 获取订单列表
func (s *OrderService) GetOrderList(userId uint, page, size int) ([]OrderListItem, int64, error) {
	var orders []model.Order
	var total int64

	offset := (page - 1) * size
	if err := database.DB.Model(&model.Order{}).Where("user_id = ?", userId).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := database.DB.Where("user_id = ?", userId).Order("created_at desc").Offset(offset).Limit(size).Find(&orders).Error; err != nil {
		return nil, 0, err
	}

	var result []OrderListItem
	for _, order := range orders {
		// 查询课程信息
		var course model.Course
		if err := database.DB.First(&course, order.CourseID).Error; err == nil {
			// 拼接课程名称：CategoryLevel2-Name
			courseName := course.CategoryLevel2 + "-" + course.Name

			result = append(result, OrderListItem{
				ID:         order.ID,
				OrderNo:    order.OrderNo,
				CourseID:   order.CourseID,
				CourseName: courseName,   // 设置拼接后的课程名称
				Price:      order.Amount, // 使用 Amount 字段
				Status:     order.Status, // Status 已经是 string 类型
				CreatedAt:  order.CreatedAt,
			})
		} else {
			// 如果课程不存在，仍然添加订单，但课程名称为空
			result = append(result, OrderListItem{
				ID:         order.ID,
				OrderNo:    order.OrderNo,
				CourseID:   order.CourseID,
				CourseName: "未知课程",       // 课程不存在时的默认值
				Price:      order.Amount, // 使用 Amount 字段
				Status:     order.Status, // Status 已经是 string 类型
				CreatedAt:  order.CreatedAt,
			})
		}
	}

	return result, total, nil
}

// 获取订单详情
func (s *OrderService) GetOrderDetail(userId uint, orderId uint) (*OrderDetail, error) {
	var order model.Order
	if err := database.DB.Where("id = ? AND user_id = ?", orderId, userId).First(&order).Error; err != nil {
		return nil, errors.New("订单不存在")
	}

	// 查询课程信息
	var course model.Course
	courseName := "未知课程" // 默认值
	if err := database.DB.First(&course, order.CourseID).Error; err == nil {
		// 拼接课程名称：CategoryLevel2-Name
		courseName = course.CategoryLevel2 + "-" + course.Name
	}

	return &OrderDetail{
		ID:          order.ID,
		OrderNo:     order.OrderNo,
		CourseID:    order.CourseID,
		CourseName:  courseName,   // 设置拼接后的课程名称
		Price:       order.Amount, // 使用 Amount 字段
		Status:      order.Status, // Status 已经是 string 类型
		PaymentType: order.PaymentType,
		PayTime:     order.PayTime,
		CreatedAt:   order.CreatedAt,
		UpdatedAt:   order.UpdatedAt,
	}, nil
}

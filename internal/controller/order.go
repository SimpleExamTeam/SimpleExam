package controller

import (
	"exam-system/internal/service"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type OrderController struct{}

// 获取订单列表
func (c *OrderController) GetList(ctx *gin.Context) {
	// 添加调试日志
	fmt.Println("=== OrderController.GetList called ===")

	// 获取当前用户ID
	userId := ctx.GetUint("userId")
	fmt.Println("userId:", userId)

	// 获取分页参数
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(ctx.DefaultQuery("size", "10"))
	fmt.Println("page:", page, "size:", size)

	// 调用新的 GetOrderList 方法
	orders, total, err := service.Order.GetOrderList(userId, page, size)

	// 详细打印每个订单的内容
	fmt.Println("=== Orders details ===")
	for i, order := range orders {
		fmt.Printf("Order %d: ID=%d, OrderNo=%s, CourseID=%d, CourseName=%s\n",
			i, order.ID, order.OrderNo, order.CourseID, order.CourseName)
	}

	fmt.Println("total:", total)
	fmt.Println("err:", err)

	if err != nil {
		ctx.JSON(http.StatusOK, gin.H{
			"code": 500,
			"msg":  err.Error(),
		})
		return
	}

	// 手动构建响应
	var responseData []map[string]interface{}
	for _, order := range orders {
		responseData = append(responseData, map[string]interface{}{
			"id":          order.ID,
			"order_no":    order.OrderNo,
			"course_id":   order.CourseID,
			"course_name": order.CourseName,
			"price":       order.Price,
			"status":      order.Status,
			"created_at":  order.CreatedAt,
		})
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": responseData,
	})
}

// 获取订单详情
func (c *OrderController) GetDetail(ctx *gin.Context) {
	// 获取当前用户ID
	userId := ctx.GetUint("userId")

	// 获取订单ID
	orderId, _ := strconv.Atoi(ctx.Param("id"))

	// 调用新的 GetOrderDetail 方法
	order, err := service.Order.GetOrderDetail(userId, uint(orderId))
	if err != nil {
		ctx.JSON(http.StatusOK, gin.H{
			"code": 500,
			"msg":  err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": order,
	})
}

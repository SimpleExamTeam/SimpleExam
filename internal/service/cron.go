package service

import (
	"exam-system/internal/model"
	"exam-system/internal/pkg/database"
	"fmt"
	"time"
)

// CronService 定时任务服务
type CronService struct {
	stopChan chan struct{}
}

var Cron = &CronService{
	stopChan: make(chan struct{}),
}

// Start 启动定时任务
func (s *CronService) Start() {
	go s.handleExpiredOrders()
}

// Stop 停止定时任务
func (s *CronService) Stop() {
	close(s.stopChan)
}

// handleExpiredOrders 处理过期订单
func (s *CronService) handleExpiredOrders() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// 查找所有pending和unpaid状态的订单
			var orders []model.Order
			if err := database.DB.Where("status IN (?) AND created_at <= ?",
				[]string{"pending", "unpaid"},
				time.Now().Add(-15*time.Minute)).
				Find(&orders).Error; err != nil {
				fmt.Printf("查询过期订单失败: %v\n", err)
				continue
			}

			// 更新过期订单状态
			for _, order := range orders {
				if err := database.DB.Model(&order).Update("status", "cancelled").Error; err != nil {
					fmt.Printf("更新订单 %s 状态失败: %v\n", order.OrderNo, err)
					continue
				}
				fmt.Printf("订单 %s 已过期，状态已更新为cancelled\n", order.OrderNo)
			}

		case <-s.stopChan:
			return
		}
	}
}

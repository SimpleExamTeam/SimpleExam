package service

import (
	"exam-system/internal/model"
	"exam-system/internal/pkg/database"
	"time"
)

var Statistics = new(StatisticsService)

type StatisticsService struct{}

// 销售统计响应
type SalesStatistics struct {
	TotalSales  float64                `json:"total_sales"`  // 总销售额
	TotalOrders int64                  `json:"total_orders"` // 总订单数
	TimeData    []SalesStatisticsPoint `json:"time_data"`    // 时间维度数据
}

// 销售统计数据点
type SalesStatisticsPoint struct {
	TimePoint string  `json:"time_point"` // 时间点，如 2023-01-01 或 2023-01 或 2023
	Sales     float64 `json:"sales"`      // 销售额
	Orders    int64   `json:"orders"`     // 订单数
}

// 时间维度类型
type TimeDimension string

const (
	DimensionDay   TimeDimension = "day"
	DimensionMonth TimeDimension = "month"
	DimensionYear  TimeDimension = "year"
)

// SystemInfo 系统信息统计
type SystemInfo struct {
	TotalUsers         int64            `json:"total_users"`          // 总用户数
	TotalCourses       int64            `json:"total_courses"`        // 总课程数
	OrdersCount        map[string]int64 `json:"orders_count"`         // 不同状态的订单数
	TotalIncome        float64          `json:"total_income"`         // 总收入（已支付订单）
	CurrentMonthIncome float64          `json:"current_month_income"` // 当月收入
	LastMonthIncome    float64          `json:"last_month_income"`    // 上月收入
}

// GetSalesStatistics 获取销售统计数据
func (s *StatisticsService) GetSalesStatistics(startTime, endTime time.Time, dimension TimeDimension) (*SalesStatistics, error) {
	// 初始化响应
	result := &SalesStatistics{
		TimeData: make([]SalesStatisticsPoint, 0),
	}

	// 查询总销售额和订单数
	var totalStats struct {
		TotalSales  float64
		TotalOrders int64
	}

	err := database.DB.Model(&model.Order{}).
		Where("status = ? AND pay_time BETWEEN ? AND ?", "paid", startTime, endTime).
		Select("SUM(amount) as total_sales, COUNT(*) as total_orders").
		Scan(&totalStats).Error
	if err != nil {
		return nil, err
	}

	result.TotalSales = totalStats.TotalSales
	result.TotalOrders = totalStats.TotalOrders

	// 根据维度查询时间序列数据
	var timeFormat string
	switch dimension {
	case DimensionDay:
		timeFormat = "%Y-%m-%d"
	case DimensionMonth:
		timeFormat = "%Y-%m"
	case DimensionYear:
		timeFormat = "%Y"
	default:
		timeFormat = "%Y-%m-%d"
	}

	// 查询按时间维度分组的数据
	type TimeStats struct {
		TimePoint  string
		Sales      float64
		OrderCount int64
	}

	var timeStats []TimeStats
	err = database.DB.Model(&model.Order{}).
		Where("status = ? AND pay_time BETWEEN ? AND ?", "paid", startTime, endTime).
		Select("DATE_FORMAT(pay_time, ?) as time_point, SUM(amount) as sales, COUNT(*) as order_count", timeFormat).
		Group("time_point").
		Order("time_point ASC").
		Scan(&timeStats).Error
	if err != nil {
		return nil, err
	}

	// 转换为响应格式
	for _, stat := range timeStats {
		result.TimeData = append(result.TimeData, SalesStatisticsPoint{
			TimePoint: stat.TimePoint,
			Sales:     stat.Sales,
			Orders:    stat.OrderCount,
		})
	}

	return result, nil
}

// GetSystemInfo 获取系统信息统计
func (s *StatisticsService) GetSystemInfo() (*SystemInfo, error) {
	result := &SystemInfo{
		OrdersCount: make(map[string]int64),
	}

	// 查询总用户数
	if err := database.DB.Model(&model.User{}).Count(&result.TotalUsers).Error; err != nil {
		return nil, err
	}

	// 查询总课程数
	if err := database.DB.Model(&model.Course{}).Count(&result.TotalCourses).Error; err != nil {
		return nil, err
	}

	// 查询不同状态的订单数
	type OrderStatusCount struct {
		Status string
		Count  int64
	}
	var orderStatusCounts []OrderStatusCount
	if err := database.DB.Model(&model.Order{}).
		Select("status, COUNT(*) as count").
		Group("status").
		Find(&orderStatusCounts).Error; err != nil {
		return nil, err
	}

	// 统计总订单数和各状态订单数
	var totalOrders int64
	for _, item := range orderStatusCounts {
		result.OrdersCount[item.Status] = item.Count
		totalOrders += item.Count
	}
	result.OrdersCount["total"] = totalOrders

	// 查询总收入（已支付订单）
	var totalIncome struct {
		Income float64
	}
	if err := database.DB.Model(&model.Order{}).
		Where("status = ?", "paid").
		Select("SUM(amount) as income").
		Scan(&totalIncome).Error; err != nil {
		return nil, err
	}
	result.TotalIncome = totalIncome.Income

	// 计算当月时间范围
	now := time.Now()
	currentMonthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	currentMonthEnd := currentMonthStart.AddDate(0, 1, 0).Add(-time.Second)

	// 查询当月收入
	var currentMonthIncome struct {
		Income float64
	}
	if err := database.DB.Model(&model.Order{}).
		Where("status = ? AND pay_time BETWEEN ? AND ?", "paid", currentMonthStart, currentMonthEnd).
		Select("SUM(amount) as income").
		Scan(&currentMonthIncome).Error; err != nil {
		return nil, err
	}
	result.CurrentMonthIncome = currentMonthIncome.Income

	// 计算上月时间范围
	lastMonthStart := currentMonthStart.AddDate(0, -1, 0)
	lastMonthEnd := currentMonthStart.Add(-time.Second)

	// 查询上月收入
	var lastMonthIncome struct {
		Income float64
	}
	if err := database.DB.Model(&model.Order{}).
		Where("status = ? AND pay_time BETWEEN ? AND ?", "paid", lastMonthStart, lastMonthEnd).
		Select("SUM(amount) as income").
		Scan(&lastMonthIncome).Error; err != nil {
		return nil, err
	}
	result.LastMonthIncome = lastMonthIncome.Income

	return result, nil
}

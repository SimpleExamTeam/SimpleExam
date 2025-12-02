package admin

import (
	"exam-system/internal/service"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// GetSalesStatistics 获取销售统计数据
func GetSalesStatistics(c *gin.Context) {
	// 获取查询参数
	dimension := c.DefaultQuery("dimension", "day") // 默认按天统计
	startTimeStr := c.Query("start_time")           // 开始时间
	endTimeStr := c.Query("end_time")               // 结束时间

	// 处理时间维度
	var timeDimension service.TimeDimension
	switch dimension {
	case "day":
		timeDimension = service.DimensionDay
	case "month":
		timeDimension = service.DimensionMonth
	case "year":
		timeDimension = service.DimensionYear
	default:
		timeDimension = service.DimensionDay
	}

	// 处理时间范围
	var startTime, endTime time.Time
	var err error

	// 如果没有提供开始时间，默认为当前时间前30天
	if startTimeStr == "" {
		startTime = time.Now().AddDate(0, 0, -30)
	} else {
		startTime, err = time.ParseInLocation("2006-01-02 15:04:05", startTimeStr, time.Local)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"code": 400,
				"msg":  "开始时间格式错误，正确格式为：2006-01-02 15:04:05",
			})
			return
		}
	}

	// 如果没有提供结束时间，默认为当前时间
	if endTimeStr == "" {
		endTime = time.Now()
	} else {
		endTime, err = time.ParseInLocation("2006-01-02 15:04:05", endTimeStr, time.Local)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"code": 400,
				"msg":  "结束时间格式错误，正确格式为：2006-01-02 15:04:05",
			})
			return
		}
	}

	// 获取销售统计数据
	statistics, err := service.Statistics.GetSalesStatistics(startTime, endTime, timeDimension)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "获取销售统计数据失败：" + err.Error(),
		})
		return
	}

	// 返回结果
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": gin.H{
			"statistics": statistics,
			"query": gin.H{
				"dimension":  dimension,
				"start_time": startTime.Format("2006-01-02 15:04:05"),
				"end_time":   endTime.Format("2006-01-02 15:04:05"),
			},
		},
	})
}

// GetSystemInfo 获取系统信息统计数据
func GetSystemInfo(c *gin.Context) {
	// 获取系统信息统计数据
	systemInfo, err := service.Statistics.GetSystemInfo()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "获取系统信息统计数据失败：" + err.Error(),
		})
		return
	}

	// 返回结果
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": systemInfo,
	})
}

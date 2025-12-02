package admin

import (
	"exam-system/internal/middleware"
	"exam-system/internal/model"
	"exam-system/internal/pkg/database"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// Login 管理员登录
func Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "参数错误",
		})
		return
	}

	// 记录登录日志信息
	loginLog := model.AdminLoginLog{
		Username:  req.Username,
		IP:        c.ClientIP(),
		UserAgent: c.GetHeader("User-Agent"),
		LoginTime: time.Now(),
		IsSuccess: false, // 默认为失败，成功时再更新
	}

	// 查询用户
	var user model.User
	if err := database.DB.Where("username = ?", req.Username).First(&user).Error; err != nil {
		loginLog.FailReason = "用户不存在"
		database.DB.Create(&loginLog)

		c.JSON(http.StatusUnauthorized, gin.H{
			"code": 401,
			"msg":  "用户名或密码错误",
		})
		return
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		loginLog.FailReason = "密码错误"
		database.DB.Create(&loginLog)

		c.JSON(http.StatusUnauthorized, gin.H{
			"code": 401,
			"msg":  "用户名或密码错误",
		})
		return
	}

	// 验证是否是管理员
	if !user.IsAdmin {
		loginLog.FailReason = "非管理员用户"
		database.DB.Create(&loginLog)

		c.JSON(http.StatusForbidden, gin.H{
			"code": 403,
			"msg":  "无管理员权限",
		})
		return
	}

	// 生成 token
	token, err := middleware.GenerateToken(user.ID)
	if err != nil {
		loginLog.FailReason = "生成token失败"
		database.DB.Create(&loginLog)

		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "生成token失败",
		})
		return
	}

	// 登录成功，更新日志
	loginLog.IsSuccess = true
	loginLog.FailReason = ""
	database.DB.Create(&loginLog)

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": gin.H{
			"token": token,
			"user": gin.H{
				"id":       user.ID,
				"username": user.Username,
				"nickname": user.Nickname,
				"avatar":   user.Avatar,
			},
		},
	})
}

// GetLoginLogs 获取管理员登录日志
func GetLoginLogs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))
	username := c.Query("username")
	status := c.Query("status") // success 或 fail
	startTime := c.Query("start_time")
	endTime := c.Query("end_time")

	// 创建查询
	db := database.DB.Model(&model.AdminLoginLog{})

	// 构建查询条件
	if username != "" {
		db = db.Where("username LIKE ?", "%"+username+"%")
	}
	if status == "success" {
		db = db.Where("is_success = ?", true)
	} else if status == "fail" {
		db = db.Where("is_success = ?", false)
	}

	// 时间范围过滤
	if startTime != "" {
		start, err := time.ParseInLocation("2006-01-02 15:04:05", startTime, time.Local)
		if err == nil {
			db = db.Where("login_time >= ?", start)
		}
	}
	if endTime != "" {
		end, err := time.ParseInLocation("2006-01-02 15:04:05", endTime, time.Local)
		if err == nil {
			db = db.Where("login_time <= ?", end)
		}
	}

	// 统计总数
	var total int64
	if err := db.Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "获取登录日志总数失败",
		})
		return
	}

	// 分页查询
	var logs []model.AdminLoginLog
	if err := db.Order("login_time DESC").
		Offset((page - 1) * size).
		Limit(size).
		Find(&logs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "获取登录日志失败",
		})
		return
	}

	// 构建返回数据
	var items []gin.H
	for _, log := range logs {
		items = append(items, gin.H{
			"id":          log.ID,
			"username":    log.Username,
			"ip":          log.IP,
			"user_agent":  log.UserAgent,
			"is_success":  log.IsSuccess,
			"fail_reason": log.FailReason,
			"login_time":  log.LoginTime,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": gin.H{
			"total": total,
			"items": items,
		},
	})
}

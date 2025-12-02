package api

import (
	"exam-system/internal/middleware"
	"exam-system/internal/model"
	"exam-system/internal/pkg/database"
	"exam-system/internal/service"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// 微信登录请求结构体
type WXLoginRequest struct {
	Code string `json:"code" binding:"required"`
}

// 微信用户信息请求结构体
type WXUserInfoRequest struct {
	UserInfo service.WXUserInfo `json:"userInfo" binding:"required"`
}

// 微信登录
func WXLogin(c *gin.Context) {
	var req WXLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "参数错误",
		})
		return
	}

	// 调用微信登录服务
	user, token, err := service.WeChat.Login(req.Code)
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

// 更新微信用户信息
func UpdateWXUserInfo(c *gin.Context) {
	var req WXUserInfoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "参数错误",
		})
		return
	}

	// 获取当前用户ID
	userId := c.GetUint("userId")

	// 调用更新用户信息服务
	err := service.WeChat.UpdateUserInfo(userId, req.UserInfo)
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

// 获取微信网页授权URL
func GetWXOAuthURL(c *gin.Context) {
	// 获取回调状态(可选)
	state := c.Query("state")
	if state == "" {
		state = "STATE"
	}

	// 获取授权URL
	url, err := service.WeChat.GetOAuthURL(state)
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
			"url": url,
		},
	})
}

// 微信网页授权回调
func WXOAuthCallback(c *gin.Context) {
	// 获取授权码
	code := c.Query("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "缺少授权码",
		})
		return
	}

	// 获取状态(可选)
	state := c.Query("state")

	// 使用授权码登录
	user, token, err := service.WeChat.LoginByOAuth(code, state)
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

// 管理员微信登录
func WXAdminLogin(c *gin.Context) {
	var req WXLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "参数错误",
		})
		return
	}

	// 记录登录日志信息
	loginLog := model.AdminLoginLog{
		Username:  "微信小程序登录", // 微信登录没有用户名，使用固定标识
		IP:        c.ClientIP(),
		UserAgent: c.GetHeader("User-Agent"),
		LoginTime: time.Now(),
		IsSuccess: false, // 默认为失败，成功时再更新
	}

	// 调用管理员微信登录服务
	user, token, err := service.WeChat.AdminLogin(req.Code)
	if err != nil {
		loginLog.FailReason = err.Error()
		database.DB.Create(&loginLog)

		c.JSON(http.StatusUnauthorized, gin.H{
			"code": 401,
			"msg":  err.Error(),
		})
		return
	}

	// 登录成功，更新日志
	loginLog.Username = user.Nickname // 记录用户昵称到username字段
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
				"is_admin": user.IsAdmin,
			},
		},
	})
}

// 管理员微信网页授权回调
func WXAdminOAuthCallback(c *gin.Context) {
	// 获取授权码
	code := c.Query("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "缺少授权码",
		})
		return
	}

	// 获取状态(可选)
	state := c.Query("state")

	// 记录登录日志信息
	loginLog := model.AdminLoginLog{
		Username:  "微信网页授权登录", // 微信登录没有用户名，使用固定标识
		IP:        c.ClientIP(),
		UserAgent: c.GetHeader("User-Agent"),
		LoginTime: time.Now(),
		IsSuccess: false, // 默认为失败，成功时再更新
	}

	// 使用授权码进行管理员登录
	user, token, err := service.WeChat.AdminLoginByOAuth(code, state)
	if err != nil {
		loginLog.FailReason = err.Error()
		database.DB.Create(&loginLog)

		c.JSON(http.StatusUnauthorized, gin.H{
			"code": 401,
			"msg":  err.Error(),
		})
		return
	}

	// 登录成功，更新日志
	loginLog.Username = user.Nickname // 记录用户昵称到username字段
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
				"is_admin": user.IsAdmin,
			},
		},
	})
}

// 获取管理员微信网页授权URL
func GetWXAdminOAuthURL(c *gin.Context) {
	// 获取回调状态(可选)
	state := c.Query("state")
	if state == "" {
		state = "ADMIN_STATE"
	}

	// 获取管理员授权URL
	url, err := service.WeChat.GetAdminOAuthURL(state)
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
			"url": url,
		},
	})
}

// 检查管理员二维码状态
func CheckAdminQRCodeStatus(c *gin.Context) {
	sceneStr := c.Query("scene_str")
	if sceneStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "scene_str不能为空",
		})
		return
	}

	qrcode, err := service.QRCode.CheckAdmin(sceneStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  err.Error(),
		})
		return
	}

	data := gin.H{
		"status": qrcode.Status,
	}

	// 如果已确认，返回管理员token
	if qrcode.Status == "confirmed" {
		var user model.User
		if err := database.DB.Where("open_id = ?", qrcode.OpenID).First(&user).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code": 500,
				"msg":  "用户不存在",
			})
			return
		}

		// 验证是否是管理员
		if !user.IsAdmin {
			c.JSON(http.StatusForbidden, gin.H{
				"code": 403,
				"msg":  "无管理员权限",
			})
			return
		}

		token, err := middleware.GenerateToken(user.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code": 500,
				"msg":  "生成token失败",
			})
			return
		}
		data["token"] = token
		data["user"] = gin.H{
			"id":       user.ID,
			"username": user.Username,
			"nickname": user.Nickname,
			"avatar":   user.Avatar,
			"is_admin": user.IsAdmin,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": data,
	})
}

// 创建管理员登录二维码
func CreateAdminLoginQRCode(c *gin.Context) {
	qrcode, qrcodeURL, err := service.QRCode.CreateAdmin()
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
			"scene_str": qrcode.SceneStr,
			"url":       qrcodeURL,
		},
	})
}

// 管理员扫码登录授权回调
func AdminQRCodeCallback(c *gin.Context) {
	code := c.Query("code")
	state := c.Query("state") // state就是scene_str

	if code == "" || state == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "参数错误",
		})
		return
	}

	// 记录登录日志信息
	loginLog := model.AdminLoginLog{
		Username:  "微信扫码登录", // 微信登录没有用户名，使用固定标识
		IP:        c.ClientIP(),
		UserAgent: c.GetHeader("User-Agent"),
		LoginTime: time.Now(),
		IsSuccess: false, // 默认为失败，成功时再更新
	}

	// 先检查二维码状态
	qrcode, err := service.QRCode.CheckAdmin(state)
	if err != nil {
		loginLog.FailReason = "检查二维码状态失败: " + err.Error()
		database.DB.Create(&loginLog)

		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "检查二维码状态失败",
		})
		return
	}

	// 如果二维码已经确认，直接返回成功
	if qrcode.Status == "confirmed" {
		// 查找用户信息
		var user model.User
		if err := database.DB.Where("open_id = ?", qrcode.OpenID).First(&user).Error; err != nil {
			loginLog.FailReason = "用户不存在"
			database.DB.Create(&loginLog)

			c.JSON(http.StatusInternalServerError, gin.H{
				"code": 500,
				"msg":  "用户不存在",
			})
			return
		}

		// 验证管理员权限
		if !user.IsAdmin {
			loginLog.FailReason = "无管理员权限"
			database.DB.Create(&loginLog)

			c.JSON(http.StatusForbidden, gin.H{
				"code": 403,
				"msg":  "无管理员权限",
			})
			return
		}

		// 生成新的token
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
		loginLog.Username = user.Nickname // 记录用户昵称到username字段
		loginLog.IsSuccess = true
		loginLog.FailReason = ""
		database.DB.Create(&loginLog)

		c.JSON(http.StatusOK, gin.H{
			"code": 200,
			"msg":  "登录成功",
			"data": gin.H{
				"token": token,
				"user": gin.H{
					"id":       user.ID,
					"username": user.Username,
					"nickname": user.Nickname,
					"avatar":   user.Avatar,
					"is_admin": user.IsAdmin,
				},
			},
		})
		return
	}

	// 使用管理员微信登录服务处理授权码
	user, token, err := service.WeChat.AdminLoginByOAuth(code, state)
	if err != nil {
		loginLog.FailReason = err.Error()
		database.DB.Create(&loginLog)

		c.JSON(http.StatusUnauthorized, gin.H{
			"code": 401,
			"msg":  err.Error(),
		})
		return
	}

	// 更新二维码状态为confirmed
	if err := service.QRCode.UpdateStatus(state, "confirmed", user.OpenID); err != nil {
		loginLog.FailReason = "更新二维码状态失败"
		database.DB.Create(&loginLog)

		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "更新二维码状态失败",
		})
		return
	}

	// 登录成功，更新日志
	loginLog.Username = user.Nickname // 记录用户昵称到username字段
	loginLog.IsSuccess = true
	loginLog.FailReason = ""
	database.DB.Create(&loginLog)

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "登录成功",
		"data": gin.H{
			"token": token,
			"user": gin.H{
				"id":       user.ID,
				"username": user.Username,
				"nickname": user.Nickname,
				"avatar":   user.Avatar,
				"is_admin": user.IsAdmin,
			},
		},
	})
}

package api

import (
	"encoding/json"
	"exam-system/internal/config"
	"exam-system/internal/model"
	"exam-system/internal/pkg/database"
	"exam-system/internal/service"
	"fmt"
	"net/http"

	"exam-system/internal/middleware"

	"github.com/gin-gonic/gin"
)

// 创建登录二维码
func CreateLoginQRCode(c *gin.Context) {
	qrcode, qrcodeURL, err := service.QRCode.Create()
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

// 检查二维码状态
func CheckQRCodeStatus(c *gin.Context) {
	sceneStr := c.Query("scene_str")
	if sceneStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "scene_str不能为空",
		})
		return
	}

	qrcode, err := service.QRCode.Check(sceneStr)
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

	// 如果已确认，返回用户token
	if qrcode.Status == "confirmed" {
		var user model.User
		if err := database.DB.Where("open_id = ?", qrcode.OpenID).First(&user).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code": 500,
				"msg":  "用户不存在",
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
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": data,
	})
}

// 扫码登录授权回调
func QRCodeCallback(c *gin.Context) {
	fmt.Printf("=== 开始处理扫码登录回调 ===\n")

	code := c.Query("code")
	state := c.Query("state") // state就是scene_str
	fmt.Printf("收到参数: code=%s, state=%s\n", code, state)

	if code == "" || state == "" {
		fmt.Println("错误: 参数为空")
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "参数错误",
		})
		return
	}

	// 先检查二维码状态
	qrcode, err := service.QRCode.Check(state)
	if err != nil {
		fmt.Printf("错误: 检查二维码状态失败: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "检查二维码状态失败",
		})
		return
	}

	// 如果二维码已经确认，直接返回成功
	if qrcode.Status == "confirmed" {
		fmt.Printf("二维码已确认，直接返回成功: state=%s\n", state)
		// 查找用户信息
		var user model.User
		if err := database.DB.Where("open_id = ?", qrcode.OpenID).First(&user).Error; err != nil {
			fmt.Printf("错误: 查找用户失败: %v\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"code": 500,
				"msg":  "用户不存在",
			})
			return
		}

		// 生成新的token
		token, err := middleware.GenerateToken(user.ID)
		if err != nil {
			fmt.Printf("错误: 生成token失败: %v\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"code": 500,
				"msg":  "生成token失败",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"code": 200,
			"msg":  "登录成功",
			"data": gin.H{
				"token": token,
				"user": gin.H{
					"id":       user.ID,
					"nickname": user.Nickname,
					"avatar":   user.Avatar,
				},
			},
		})
		return
	}

	// 获取微信配置
	cfg := config.GlobalConfig.WeChat
	fmt.Printf("微信配置: AppID=%s, QRCodeCallback=%s\n", cfg.AppID, cfg.QRCodeCallback)

	// 通过code获取access_token和openid
	accessTokenURL := fmt.Sprintf("https://api.weixin.qq.com/sns/oauth2/access_token?"+
		"appid=%s&"+
		"secret=%s&"+
		"code=%s&"+
		"grant_type=authorization_code",
		cfg.AppID,
		cfg.AppSecret,
		code,
	)
	fmt.Printf("请求access_token URL: %s\n", accessTokenURL)

	// 发送请求获取access_token
	resp, err := http.Get(accessTokenURL)
	if err != nil {
		fmt.Printf("错误: 获取access_token失败: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "获取access_token失败",
		})
		return
	}
	defer resp.Body.Close()

	// 解析响应
	var result struct {
		AccessToken string `json:"access_token"`
		OpenID      string `json:"openid"`
		ErrCode     int    `json:"errcode"`
		ErrMsg      string `json:"errmsg"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		fmt.Printf("错误: 解析access_token响应失败: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "解析access_token响应失败",
		})
		return
	}

	fmt.Printf("access_token响应: %+v\n", result)

	// 检查是否有错误
	if result.ErrCode != 0 {
		fmt.Printf("错误: 获取access_token失败: %s\n", result.ErrMsg)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  fmt.Sprintf("获取access_token失败: %s", result.ErrMsg),
		})
		return
	}

	// 获取用户信息
	userInfoURL := fmt.Sprintf("https://api.weixin.qq.com/sns/userinfo?"+
		"access_token=%s&"+
		"openid=%s&"+
		"lang=zh_CN",
		result.AccessToken,
		result.OpenID,
	)
	fmt.Printf("请求用户信息URL: %s\n", userInfoURL)

	// 发送请求获取用户信息
	resp, err = http.Get(userInfoURL)
	if err != nil {
		fmt.Printf("错误: 获取用户信息失败: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "获取用户信息失败",
		})
		return
	}
	defer resp.Body.Close()

	// 解析用户信息
	var userInfo struct {
		OpenID   string `json:"openid"`
		Nickname string `json:"nickname"`
		Sex      int    `json:"sex"`
		Province string `json:"province"`
		City     string `json:"city"`
		Country  string `json:"country"`
		Avatar   string `json:"headimgurl"`
		UnionID  string `json:"unionid"`
		ErrCode  int    `json:"errcode"`
		ErrMsg   string `json:"errmsg"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		fmt.Printf("错误: 解析用户信息失败: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "解析用户信息失败",
		})
		return
	}

	fmt.Printf("用户信息响应: %+v\n", userInfo)

	// 检查是否有错误
	if userInfo.ErrCode != 0 {
		fmt.Printf("错误: 获取用户信息失败: %s\n", userInfo.ErrMsg)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  fmt.Sprintf("获取用户信息失败: %s", userInfo.ErrMsg),
		})
		return
	}

	// 更新或创建用户
	var user model.User
	if err := database.DB.Where("open_id = ?", userInfo.OpenID).First(&user).Error; err != nil {
		fmt.Printf("用户不存在，创建新用户: OpenID=%s\n", userInfo.OpenID)
		// 用户不存在，创建新用户
		user = model.User{
			OpenID:   userInfo.OpenID,
			UnionID:  userInfo.UnionID,
			Nickname: userInfo.Nickname,
			Avatar:   userInfo.Avatar,
			Sex:      userInfo.Sex,
			Country:  userInfo.Country,
			Province: userInfo.Province,
			City:     userInfo.City,
		}
		if err := database.DB.Create(&user).Error; err != nil {
			fmt.Printf("错误: 创建用户失败: %v\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"code": 500,
				"msg":  "创建用户失败",
			})
			return
		}
		fmt.Printf("新用户创建成功: ID=%d\n", user.ID)
	} else {
		fmt.Printf("用户已存在，更新信息: ID=%d, OpenID=%s\n", user.ID, user.OpenID)
		// 用户存在，更新信息
		user.Nickname = userInfo.Nickname
		user.UnionID = userInfo.UnionID
		user.Avatar = userInfo.Avatar
		user.Sex = userInfo.Sex
		user.Country = userInfo.Country
		user.Province = userInfo.Province
		user.City = userInfo.City
		if err := database.DB.Save(&user).Error; err != nil {
			fmt.Printf("错误: 更新用户信息失败: %v\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"code": 500,
				"msg":  "更新用户信息失败",
			})
			return
		}
		fmt.Printf("用户信息更新成功: ID=%d\n", user.ID)
	}

	// 更新二维码状态为confirmed
	fmt.Printf("更新二维码状态: state=%s, openID=%s\n", state, userInfo.OpenID)
	if err := service.QRCode.UpdateStatus(state, "confirmed", userInfo.OpenID); err != nil {
		fmt.Printf("错误: 更新二维码状态失败: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "更新二维码状态失败",
		})
		return
	}
	fmt.Println("二维码状态更新成功")

	// 生成token
	token, err := middleware.GenerateToken(user.ID)
	if err != nil {
		fmt.Printf("错误: 生成token失败: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "生成token失败",
		})
		return
	}
	fmt.Printf("生成token成功: %s\n", token[:10]) // 只显示前10位

	// 返回成功响应
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "登录成功",
		"data": gin.H{
			"token": token,
			"user": gin.H{
				"id":       user.ID,
				"nickname": user.Nickname,
				"avatar":   user.Avatar,
			},
		},
	})
}

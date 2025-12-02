package service

import (
	"exam-system/internal/config"
	"exam-system/internal/model"
	"exam-system/internal/pkg/database"
	"fmt"
	"time"

	"github.com/google/uuid"
)

var QRCode = new(QRCodeService)

type QRCodeService struct{}

// 创建二维码
func (s *QRCodeService) Create() (*model.QRCode, string, error) {
	// 生成唯一的场景值
	sceneStr := uuid.New().String()

	// 创建二维码记录
	qrcode := &model.QRCode{
		SceneStr: sceneStr,
		Status:   "pending",
	}

	if err := database.DB.Create(qrcode).Error; err != nil {
		return nil, "", fmt.Errorf("创建二维码记录失败: %v", err)
	}

	// 获取微信配置
	cfg := config.GlobalConfig.WeChat

	// 生成二维码URL
	qrcodeURL := fmt.Sprintf("https://open.weixin.qq.com/connect/oauth2/authorize?"+
		"appid=%s&"+
		"redirect_uri=%s&"+
		"response_type=code&"+
		"scope=snsapi_userinfo&"+
		"state=%s#wechat_redirect",
		cfg.AppID,
		cfg.QRCodeCallback,
		sceneStr)

	return qrcode, qrcodeURL, nil
}

// 创建管理员二维码
func (s *QRCodeService) CreateAdmin() (*model.QRCode, string, error) {
	// 生成唯一的场景值
	sceneStr := uuid.New().String()

	// 创建二维码记录
	qrcode := &model.QRCode{
		SceneStr: sceneStr,
		Status:   "pending",
	}

	if err := database.DB.Create(qrcode).Error; err != nil {
		return nil, "", fmt.Errorf("创建二维码记录失败: %v", err)
	}

	// 获取微信配置
	cfg := config.GlobalConfig.WeChat

	// 确定管理员回调地址
	callbackURL := cfg.AdminQRCodeCallback
	if callbackURL == "" {
		// 如果没有配置管理员专用回调地址，使用普通回调地址
		callbackURL = cfg.QRCodeCallback
	}

	// 生成二维码URL
	qrcodeURL := fmt.Sprintf("https://open.weixin.qq.com/connect/oauth2/authorize?"+
		"appid=%s&"+
		"redirect_uri=%s&"+
		"response_type=code&"+
		"scope=snsapi_userinfo&"+
		"state=%s#wechat_redirect",
		cfg.AppID,
		callbackURL,
		sceneStr)

	return qrcode, qrcodeURL, nil
}

// 检查二维码状态
func (s *QRCodeService) Check(sceneStr string) (*model.QRCode, error) {
	var qrcode model.QRCode
	if err := database.DB.Where("scene_str = ?", sceneStr).First(&qrcode).Error; err != nil {
		return nil, fmt.Errorf("二维码不存在")
	}

	// 检查是否过期（5分钟）
	if time.Since(qrcode.CreatedAt) > 5*time.Minute && qrcode.Status == "pending" {
		// 更新状态为过期
		qrcode.Status = "expired"
		database.DB.Save(&qrcode)
		return &qrcode, nil
	}

	return &qrcode, nil
}

// 检查管理员二维码状态
func (s *QRCodeService) CheckAdmin(sceneStr string) (*model.QRCode, error) {
	var qrcode model.QRCode
	if err := database.DB.Where("scene_str = ?", sceneStr).First(&qrcode).Error; err != nil {
		return nil, fmt.Errorf("二维码不存在")
	}

	// 检查是否过期（5分钟）
	if time.Since(qrcode.CreatedAt) > 5*time.Minute && qrcode.Status == "pending" {
		// 更新状态为过期
		qrcode.Status = "expired"
		database.DB.Save(&qrcode)
		return &qrcode, nil
	}

	// 如果二维码已确认，需要验证用户是否是管理员
	if qrcode.Status == "confirmed" && qrcode.OpenID != "" {
		var user model.User
		if err := database.DB.Where("open_id = ?", qrcode.OpenID).First(&user).Error; err != nil {
			// 如果用户不存在，将二维码状态设为过期
			qrcode.Status = "expired"
			database.DB.Save(&qrcode)
			return &qrcode, fmt.Errorf("用户不存在")
		}

		// 如果用户不是管理员，将二维码状态设为过期
		if !user.IsAdmin {
			qrcode.Status = "expired"
			database.DB.Save(&qrcode)
			return &qrcode, fmt.Errorf("用户无管理员权限")
		}
	}

	return &qrcode, nil
}

// 更新二维码状态
func (s *QRCodeService) UpdateStatus(sceneStr string, status string, openID string) error {
	result := database.DB.Model(&model.QRCode{}).
		Where("scene_str = ?", sceneStr).
		Updates(map[string]interface{}{
			"status":  status,
			"open_id": openID,
		})

	if result.Error != nil {
		return fmt.Errorf("更新状态失败: %v", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("二维码不存在")
	}

	return nil
}

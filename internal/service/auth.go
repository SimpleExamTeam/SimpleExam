package service

import (
	"errors"
	"exam-system/internal/config"
	"exam-system/internal/middleware"
	"exam-system/internal/model"
	"exam-system/internal/pkg/database"
	"exam-system/internal/pkg/logger"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var Auth = new(AuthService)

type AuthService struct{}

func (s *AuthService) Login(username, password string) (string, *model.User, error) {
	var user model.User
	if err := database.DB.Where("username = ?", username).First(&user).Error; err != nil {
		return "", nil, errors.New("用户不存在")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", nil, errors.New("密码错误")
	}

	// 使用中间件中的 GenerateToken 函数
	token, err := middleware.GenerateToken(user.ID)
	if err != nil {
		return "", nil, err
	}

	return token, &user, nil
}

func (s *AuthService) Register(username, password, nickname string) (*model.User, error) {
	// 检查用户名是否已存在
	var count int64
	database.DB.Model(&model.User{}).Where("username = ?", username).Count(&count)
	if count > 0 {
		return nil, errors.New("用户名已存在")
	}

	// 加密密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &model.User{
		Username: username,
		Password: string(hashedPassword),
		Nickname: nickname,
	}

	if err := database.DB.Create(user).Error; err != nil {
		return nil, err
	}

	return user, nil
}

// EnsureDefaultAdmin 检查并在缺失时创建默认管理员账号
func (s *AuthService) EnsureDefaultAdmin() error {
	var user model.User
	err := database.DB.Where("username = ?", "admin").First(&user).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("查询管理员账号失败: %w", err)
	}

	if errors.Is(err, gorm.ErrRecordNotFound) {
		// 创建默认管理员账号
		hashedPassword, hashErr := bcrypt.GenerateFromPassword([]byte("simple_exam"), bcrypt.DefaultCost)
		if hashErr != nil {
			return fmt.Errorf("生成管理员默认密码失败: %w", hashErr)
		}

		admin := &model.User{
			Username: "admin",
			Password: string(hashedPassword),
			Nickname: "管理员",
			IsAdmin:  true,
		}

		if createErr := database.DB.Create(admin).Error; createErr != nil {
			return fmt.Errorf("创建默认管理员账号失败: %w", createErr)
		}

		logger.Infof("默认管理员账号已创建，用户名: %s", admin.Username)
		return nil
	}

	// 如果存在但未标记为管理员，进行修正
	if !user.IsAdmin {
		if updateErr := database.DB.Model(&model.User{}).
			Where("id = ?", user.ID).
			Update("is_admin", true).Error; updateErr != nil {
			return fmt.Errorf("更新管理员标识失败: %w", updateErr)
		}
		logger.Infof("账号 %s 已标记为管理员", user.Username)
	}

	return nil
}

// ResetPassword 通过用户名重置密码
func (s *AuthService) ResetPassword(username, password string) error {
	if username == "" || password == "" {
		return errors.New("用户名或密码不能为空")
	}

	var user model.User
	if err := database.DB.Where("username = ?", username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("用户不存在: %s", username)
		}
		return fmt.Errorf("查询用户失败: %w", err)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("加密密码失败: %w", err)
	}

	if err := database.DB.Model(&model.User{}).
		Where("id = ?", user.ID).
		Update("password", string(hashedPassword)).Error; err != nil {
		return fmt.Errorf("更新密码失败: %w", err)
	}

	logger.Infof("用户 %s 的密码已被重置", username)
	return nil
}

// GenerateToken 生成JWT token
func (s *AuthService) GenerateToken(openID string) (string, error) {
	// 查找用户
	var user model.User
	if err := database.DB.Where("open_id = ?", openID).First(&user).Error; err != nil {
		return "", fmt.Errorf("用户不存在")
	}

	// 创建JWT Claims
	claims := jwt.MapClaims{
		"user_id": user.ID,
		"open_id": user.OpenID,
		"exp":     time.Now().Add(time.Duration(config.GlobalConfig.JWT.ExpireTime) * time.Second).Unix(),
	}

	// 生成token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(config.GlobalConfig.JWT.Secret))
	if err != nil {
		return "", fmt.Errorf("生成token失败: %v", err)
	}

	return signedToken, nil
}

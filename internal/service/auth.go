package service

import (
	"errors"
	"exam-system/internal/config"
	"exam-system/internal/middleware"
	"exam-system/internal/model"
	"exam-system/internal/pkg/database"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt"
	"golang.org/x/crypto/bcrypt"
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

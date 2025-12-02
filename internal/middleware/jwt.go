package middleware

import (
	"exam-system/internal/config"
	"exam-system/internal/model"
	"exam-system/internal/pkg/database"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
)

func JWT() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取JWT配置
		if config.GlobalConfig == nil {
			fmt.Printf("配置未初始化\n")
			c.JSON(http.StatusInternalServerError, gin.H{
				"code": 500,
				"msg":  "系统错误，无法验证身份",
			})
			c.Abort()
			return
		}
		jwtConfig := config.GlobalConfig.JWT

		authHeader := c.GetHeader("Authorization")

		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code": 401,
				"msg":  "未登录或token已过期",
			})
			c.Abort()
			return
		}

		// 获取token
		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code": 401,
				"msg":  "token格式错误",
			})
			c.Abort()
			return
		}

		tokenStr := parts[1]

		// 解析token
		claims, err := parseToken(tokenStr, jwtConfig.Secret)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code": 401,
				"msg":  "token无效: " + err.Error(),
			})
			c.Abort()
			return
		}

		// 检查用户是否存在且未被删除
		var user model.User
		if err := database.DB.Unscoped().First(&user, claims.UserID).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code": 401,
				"msg":  "用户不存在或已被删除",
			})
			c.Abort()
			return
		}

		// 检查用户是否已被删除
		if !user.DeletedAt.Time.IsZero() {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code": 401,
				"msg":  "用户已被删除",
			})
			c.Abort()
			return
		}

		// 将用户信息保存到上下文
		c.Set("userId", claims.UserID)
		c.Next()
	}
}

type Claims struct {
	UserID uint `json:"userId"`
	jwt.StandardClaims
}

func parseToken(tokenString string, secretKey string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// 验证签名方法
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("不支持的签名方法: %v", token.Header["alg"])
		}
		return []byte(secretKey), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("无效的token")
}

// GenerateToken 生成JWT token
func GenerateToken(userId uint) (string, error) {
	// 获取配置
	if config.GlobalConfig == nil {
		return "", fmt.Errorf("配置未初始化")
	}
	jwtConfig := config.GlobalConfig.JWT

	// 计算过期时间
	expireSeconds := jwtConfig.ExpireTime
	expireTime := time.Now().Add(time.Duration(expireSeconds) * time.Second)

	fmt.Printf("生成token，用户ID: %d，过期时间: %v（%d秒后）\n",
		userId, expireTime, expireSeconds)

	claims := Claims{
		UserID: userId,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expireTime.Unix(),
			IssuedAt:  time.Now().Unix(),
		},
	}

	// 创建token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// 签名并返回token字符串
	tokenStr, err := token.SignedString([]byte(jwtConfig.Secret))
	if err == nil {
		fmt.Printf("生成的token: %s...\n", tokenStr[:10]) // 只显示前10位
	}
	return tokenStr, err
}

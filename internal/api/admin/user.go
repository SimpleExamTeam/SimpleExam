package admin

import (
	"exam-system/internal/model"
	"exam-system/internal/pkg/database"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// GetUsers 获取用户列表
func GetUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))
	keyword := c.Query("keyword")
	isAdmin := c.Query("is_admin")

	var users []model.User
	var total int64
	query := database.DB.Model(&model.User{})

	// 关键字搜索
	if keyword != "" {
		query = query.Where("username LIKE ? OR nickname LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}

	// 管理员筛选
	if isAdmin != "" {
		if isAdmin == "true" {
			query = query.Where("is_admin = ?", true)
		} else if isAdmin == "false" {
			query = query.Where("is_admin = ?", false)
		}
	}

	// 统计总数
	if err := query.Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "获取用户总数失败",
		})
		return
	}

	// 分页查询
	err := query.Order("created_at DESC").Offset((page - 1) * size).Limit(size).Find(&users).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "获取用户列表失败",
		})
		return
	}

	// 处理返回数据，去除敏感信息
	var userList []gin.H
	for _, user := range users {
		userList = append(userList, gin.H{
			"id":         user.ID,
			"username":   user.Username,
			"nickname":   user.Nickname,
			"avatar":     user.Avatar,
			"sex":        user.Sex,
			"country":    user.Country,
			"province":   user.Province,
			"city":       user.City,
			"is_admin":   user.IsAdmin,
			"open_id":    user.OpenID,
			"union_id":   user.UnionID,
			"created_at": user.CreatedAt,
			"updated_at": user.UpdatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": gin.H{
			"total": total,
			"items": userList,
		},
	})
}

// GetUser 获取单个用户
func GetUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "参数错误",
		})
		return
	}

	var user model.User
	if err := database.DB.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code": 404,
			"msg":  "用户不存在",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": gin.H{
			"id":         user.ID,
			"username":   user.Username,
			"nickname":   user.Nickname,
			"avatar":     user.Avatar,
			"sex":        user.Sex,
			"country":    user.Country,
			"province":   user.Province,
			"city":       user.City,
			"is_admin":   user.IsAdmin,
			"open_id":    user.OpenID,
			"union_id":   user.UnionID,
			"created_at": user.CreatedAt,
			"updated_at": user.UpdatedAt,
		},
	})
}

// CreateUserRequest 创建用户请求
type CreateUserRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Nickname string `json:"nickname" binding:"required"`
	Avatar   string `json:"avatar"`
	Sex      int    `json:"sex"`
	Country  string `json:"country"`
	Province string `json:"province"`
	City     string `json:"city"`
	IsAdmin  bool   `json:"is_admin"`
}

// CreateUser 创建用户
func CreateUser(c *gin.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "参数错误",
		})
		return
	}

	// 检查用户名是否已存在
	var count int64
	if err := database.DB.Model(&model.User{}).Where("username = ?", req.Username).Count(&count).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "检查用户名失败",
		})
		return
	}

	if count > 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "用户名已存在",
		})
		return
	}

	// 加密密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "密码加密失败",
		})
		return
	}

	user := model.User{
		Username: req.Username,
		Password: string(hashedPassword),
		Nickname: req.Nickname,
		Avatar:   req.Avatar,
		Sex:      req.Sex,
		Country:  req.Country,
		Province: req.Province,
		City:     req.City,
		IsAdmin:  req.IsAdmin,
	}

	if err := database.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "创建用户失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": gin.H{
			"id": user.ID,
		},
	})
}

// UpdateUserRequest 更新用户请求
type UpdateUserRequest struct {
	Nickname string `json:"nickname"`
	Password string `json:"password"`
	Avatar   string `json:"avatar"`
	Sex      *int   `json:"sex"`
	Country  string `json:"country"`
	Province string `json:"province"`
	City     string `json:"city"`
	IsAdmin  *bool  `json:"is_admin"`
}

// UpdateUser 更新用户
func UpdateUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "参数错误",
		})
		return
	}

	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "参数错误",
		})
		return
	}

	updates := make(map[string]interface{})

	if req.Nickname != "" {
		updates["nickname"] = req.Nickname
	}
	if req.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code": 500,
				"msg":  "密码加密失败",
			})
			return
		}
		updates["password"] = string(hashedPassword)
	}
	if req.Avatar != "" {
		updates["avatar"] = req.Avatar
	}
	if req.Sex != nil {
		updates["sex"] = *req.Sex
	}
	if req.Country != "" {
		updates["country"] = req.Country
	}
	if req.Province != "" {
		updates["province"] = req.Province
	}
	if req.City != "" {
		updates["city"] = req.City
	}
	if req.IsAdmin != nil {
		updates["is_admin"] = *req.IsAdmin
	}

	result := database.DB.Model(&model.User{}).Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "更新用户失败",
		})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"code": 404,
			"msg":  "用户不存在",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "更新成功",
	})
}

// DeleteUser 删除用户
func DeleteUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "参数错误",
		})
		return
	}

	// 检查是否有关联的订单
	var orderCount int64
	if err := database.DB.Model(&model.Order{}).Where("user_id = ?", id).Count(&orderCount).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "检查关联订单失败",
		})
		return
	}

	if orderCount > 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "该用户已有订单关联，无法删除",
		})
		return
	}

	// 检查是否有关联的考试记录
	var examCount int64
	if err := database.DB.Model(&model.ExamRecord{}).Where("user_id = ?", id).Count(&examCount).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 400,
			"msg":  "检查关联考试记录失败",
		})
		return
	}

	if examCount > 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "该用户已有考试记录，无法删除",
		})
		return
	}

	result := database.DB.Delete(&model.User{}, id)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "删除用户失败",
		})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"code": 404,
			"msg":  "用户不存在",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "删除成功",
	})
}

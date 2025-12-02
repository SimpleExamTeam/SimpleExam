package admin

import (
	"exam-system/internal/model"
	"exam-system/internal/pkg/database"
	"exam-system/internal/utils"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// CardQuery 卡券查询参数
type CardQuery struct {
	Page     int    `form:"page,default=1"`
	Size     int    `form:"size,default=10"`
	CardNo   string `form:"card_no"`
	CourseID uint   `form:"course_id"`
}

// GetCards 获取卡券列表
func GetCards(c *gin.Context) {
	var query CardQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "参数错误",
		})
		return
	}

	// 设置默认值
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.Size <= 0 {
		query.Size = 10
	}

	db := database.DB.Model(&model.Card{})

	// 构建查询条件
	if query.CardNo != "" {
		db = db.Where("card_no LIKE ?", "%"+query.CardNo+"%")
	}
	if query.CourseID > 0 {
		db = db.Where("course_id = ?", query.CourseID)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "获取卡券总数失败",
		})
		return
	}

	var cards []model.Card
	if err := db.Order("id DESC").
		Offset((query.Page - 1) * query.Size).
		Limit(query.Size).
		Find(&cards).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "获取卡券列表失败",
		})
		return
	}

	// 处理返回数据
	cardList := make([]gin.H, 0)
	for _, card := range cards {
		// 查询课程信息
		var courseName string = "全部课程"
		if card.CourseID != nil && *card.CourseID > 0 {
			var course model.Course
			if err := database.DB.First(&course, *card.CourseID).Error; err == nil {
				courseName = course.Name
			}
		}

		// 计算过期时间
		expireTime := card.CreatedAt.AddDate(0, 0, card.ExpireDays)
		isExpired := time.Now().After(expireTime)

		cardList = append(cardList, gin.H{
			"id":          card.ID,
			"card_no":     card.CardNo,
			"course_id":   card.CourseID,
			"course_name": courseName,
			"amount":      card.Amount,
			"total":       card.Total,
			"used":        card.Used,
			"expire_days": card.ExpireDays,
			"expire_time": expireTime,
			"is_expired":  isExpired,
			"created_at":  card.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": gin.H{
			"total": total,
			"items": cardList,
		},
	})
}

// GetCard 获取单个卡券
func GetCard(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "参数错误",
		})
		return
	}

	var card model.Card
	if err := database.DB.First(&card, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code": 404,
			"msg":  "卡券不存在",
		})
		return
	}

	// 查询课程信息
	var courseName string = "全部课程"
	if card.CourseID != nil && *card.CourseID > 0 {
		var course model.Course
		if err = database.DB.First(&course, *card.CourseID).Error; err == nil {
			courseName = course.Name
		}
	}

	// 计算过期时间
	expireTime := card.CreatedAt.AddDate(0, 0, card.ExpireDays)
	isExpired := time.Now().After(expireTime)

	// 查询兑换记录
	var records []model.CardRecord
	database.DB.Where("card_id = ?", card.ID).Find(&records)

	// 处理兑换记录
	recordList := make([]gin.H, 0)
	for _, record := range records {
		var user model.User
		database.DB.First(&user, record.UserID)

		recordList = append(recordList, gin.H{
			"id":         record.ID,
			"user_id":    record.UserID,
			"username":   user.Username,
			"nickname":   user.Nickname,
			"order_id":   record.OrderID,
			"order_no":   record.OrderNo,
			"course_id":  record.CourseID,
			"amount":     record.Amount,
			"created_at": record.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": gin.H{
			"id":          card.ID,
			"card_no":     card.CardNo,
			"course_id":   card.CourseID,
			"course_name": courseName,
			"amount":      card.Amount,
			"total":       card.Total,
			"used":        card.Used,
			"expire_days": card.ExpireDays,
			"expire_time": expireTime,
			"is_expired":  isExpired,
			"created_at":  card.CreatedAt,
			"records":     recordList,
		},
	})
}

// CreateCardRequest 创建卡券请求
type CreateCardRequest struct {
	CourseID   *uint   `json:"course_id"`
	Amount     float64 `json:"amount"`
	Total      int     `json:"total" binding:"required"`
	ExpireDays int     `json:"expire_days" binding:"required"`
}

// CreateCard 创建卡券
func CreateCard(c *gin.Context) {
	var req CreateCardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "参数错误",
		})
		return
	}

	// 验证参数
	if req.Total <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "可兑换总数必须大于0",
		})
		return
	}

	if req.ExpireDays <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "有效期天数必须大于0",
		})
		return
	}

	// 验证课程是否存在
	if req.CourseID != nil && *req.CourseID > 0 {
		var course model.Course
		if err := database.DB.First(&course, *req.CourseID).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"code": 400,
				"msg":  "课程不存在",
			})
			return
		}
	}

	// 生成卡券编号
	cardNo := generateCardNo()

	// 创建卡券
	card := model.Card{
		CardNo:     cardNo,
		CourseID:   req.CourseID,
		Amount:     req.Amount,
		Total:      req.Total,
		ExpireDays: req.ExpireDays,
		Used:       0,
	}

	if err := database.DB.Create(&card).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "创建卡券失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": gin.H{
			"id":      card.ID,
			"card_no": card.CardNo,
		},
	})
}

// UpdateCardRequest 更新卡券请求
type UpdateCardRequest struct {
	CourseID   *uint   `json:"course_id"`
	Amount     float64 `json:"amount"`
	Total      int     `json:"total"`
	ExpireDays int     `json:"expire_days"`
}

// UpdateCard 更新卡券
func UpdateCard(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "参数错误",
		})
		return
	}

	var req UpdateCardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "参数错误",
		})
		return
	}

	// 验证卡券是否存在
	var card model.Card
	if err := database.DB.First(&card, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code": 404,
			"msg":  "卡券不存在",
		})
		return
	}

	// 验证参数
	if req.Total > 0 && req.Total < card.Used {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "可兑换总数不能小于已兑换数量",
		})
		return
	}

	// 验证课程是否存在
	if req.CourseID != nil && *req.CourseID > 0 {
		var course model.Course
		if err := database.DB.First(&course, *req.CourseID).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"code": 400,
				"msg":  "课程不存在",
			})
			return
		}
	}

	// 构建更新内容
	updates := make(map[string]interface{})
	updates["course_id"] = req.CourseID
	updates["amount"] = req.Amount
	if req.Total > 0 {
		updates["total"] = req.Total
	}
	if req.ExpireDays > 0 {
		updates["expire_days"] = req.ExpireDays
	}

	// 更新卡券
	if err := database.DB.Model(&card).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "更新卡券失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "更新成功",
	})
}

// DeleteCard 删除卡券
func DeleteCard(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "参数错误",
		})
		return
	}

	// 验证卡券是否存在
	var card model.Card
	if err := database.DB.First(&card, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code": 404,
			"msg":  "卡券不存在",
		})
		return
	}

	// 删除卡券
	if err := database.DB.Delete(&card).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "删除卡券失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "删除成功",
	})
}

// CardRecordQuery 卡券兑换记录查询参数
type CardRecordQuery struct {
	Page     int    `form:"page,default=1"`
	Size     int    `form:"size,default=10"`
	CardNo   string `form:"card_no"`
	Username string `form:"username"`
	CourseID uint   `form:"course_id"`
}

// GetAllCardRecords 获取所有卡券兑换记录
func GetAllCardRecords(c *gin.Context) {
	var query CardRecordQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "参数错误",
		})
		return
	}

	// 设置默认值
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.Size <= 0 {
		query.Size = 10
	}

	db := database.DB.Model(&model.CardRecord{})

	// 构建查询条件
	if query.CardNo != "" {
		var card model.Card
		if err := database.DB.Where("card_no LIKE ?", "%"+query.CardNo+"%").First(&card).Error; err == nil {
			db = db.Where("card_id = ?", card.ID)
		}
	}
	if query.Username != "" {
		var users []model.User
		database.DB.Where("username LIKE ? OR nickname LIKE ?", "%"+query.Username+"%", "%"+query.Username+"%").Find(&users)
		if len(users) > 0 {
			userIDs := make([]uint, len(users))
			for i, user := range users {
				userIDs[i] = user.ID
			}
			db = db.Where("user_id IN ?", userIDs)
		}
	}
	if query.CourseID > 0 {
		db = db.Where("course_id = ?", query.CourseID)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "获取兑换记录总数失败",
		})
		return
	}

	var records []model.CardRecord
	if err := db.Order("id DESC").
		Offset((query.Page - 1) * query.Size).
		Limit(query.Size).
		Find(&records).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "获取兑换记录列表失败",
		})
		return
	}

	// 处理返回数据
	recordList := make([]gin.H, 0)
	for _, record := range records {
		var user model.User
		database.DB.First(&user, record.UserID)

		var card model.Card
		database.DB.First(&card, record.CardID)

		var course model.Course
		database.DB.First(&course, record.CourseID)

		recordList = append(recordList, gin.H{
			"id":          record.ID,
			"card_id":     record.CardID,
			"card_no":     card.CardNo,
			"user_id":     record.UserID,
			"username":    user.Username,
			"nickname":    user.Nickname,
			"order_id":    record.OrderID,
			"order_no":    record.OrderNo,
			"course_id":   record.CourseID,
			"course_name": course.Name,
			"amount":      record.Amount,
			"created_at":  record.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": gin.H{
			"total": total,
			"items": recordList,
		},
	})
}

// GetCardRecords 获取指定卡券的兑换记录
func GetCardRecords(c *gin.Context) {
	cardID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "参数错误",
		})
		return
	}

	// 验证卡券是否存在
	var card model.Card
	if err := database.DB.First(&card, cardID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code": 404,
			"msg":  "卡券不存在",
		})
		return
	}

	// 获取兑换记录
	var records []model.CardRecord
	if err := database.DB.Where("card_id = ?", cardID).Order("id DESC").Find(&records).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "获取兑换记录失败",
		})
		return
	}

	// 处理返回数据
	recordList := make([]gin.H, 0)
	for _, record := range records {
		var user model.User
		database.DB.First(&user, record.UserID)

		var course model.Course
		database.DB.First(&course, record.CourseID)

		recordList = append(recordList, gin.H{
			"id":          record.ID,
			"user_id":     record.UserID,
			"username":    user.Username,
			"nickname":    user.Nickname,
			"order_id":    record.OrderID,
			"order_no":    record.OrderNo,
			"course_id":   record.CourseID,
			"course_name": course.CategoryLevel2 + "-" + course.Name,
			"amount":      record.Amount,
			"created_at":  record.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": recordList,
	})
}

// 生成卡券编号，18位数字+字母混合
func generateCardNo() string {
	rand.Seed(time.Now().UnixNano())
	return utils.GenerateRandomString(18)
}

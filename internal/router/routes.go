package router

import (
	"exam-system/internal/api"
	"exam-system/internal/api/admin"
	"exam-system/internal/middleware"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// SetupRoutes 配置所有路由
func SetupRoutes(r *gin.Engine, userFS, adminFS http.FileSystem) {
	// 健康检查接口（不需要任何中间件）
	r.GET("/api/v1/health", api.SimpleHealthCheck)

	// 设置静态文件路由 - 直接访问静态资源文件
	r.StaticFS("/static/user", userFS)
	r.StaticFS("/static/admin", adminFS)

	// 设置前端SPA路由
	setupSPARoutes(r, userFS, adminFS)

	// 用户API路由
	setupAPIRoutes(r)

	// 管理员API路由
	setupAdminAPIRoutes(r)
}

// setupAPIRoutes 设置用户API路由
func setupAPIRoutes(r *gin.Engine) {
	// API路由
	apiGroup := r.Group("/api/v1")
	apiGroup.Use(middleware.Logger())
	apiGroup.Use(middleware.Recovery())
	apiGroup.Use(middleware.Cors())

	// 认证相关
	auth := apiGroup.Group("/auth")
	{
		auth.POST("/login", api.Login)
		auth.POST("/register", api.Register)
		auth.POST("/wx/login", api.WXLogin)
	}

	// 微信相关（不需要认证）
	wx := apiGroup.Group("/wechat")
	{
		// 微信网页授权
		wx.GET("/oauth/url", api.GetWXOAuthURL)        // 获取授权URL
		wx.GET("/oauth/callback", api.WXOAuthCallback) // 授权回调

		// 微信扫码登录
		wx.POST("/qrcode/create", api.CreateLoginQRCode) // 创建登录二维码
		wx.GET("/qrcode/check", api.CheckQRCodeStatus)   // 检查二维码状态
		wx.GET("/qrcode/callback", api.QRCodeCallback)   // 扫码登录回调
	}

	// 支付回调（不需要认证）
	apiGroup.POST("/payments/notify", api.PaymentNotify)
	// 退款回调（不需要认证）
	apiGroup.POST("/payments/refund/notify", api.RefundNotify)

	// 需要认证的路由
	authorized := apiGroup.Group("/")
	authorized.Use(middleware.JWT())
	{
		// 支付相关
		payments := authorized.Group("/payments")
		{
			payments.POST("/create", api.CreatePayment)
			payments.GET("/query/:order_no", api.QueryPayment)
			payments.POST("/cancel/:order_no", api.CancelPayment)
			payments.POST("/redeem-card", api.RedeemCard) // 兑换卡券
		}

		// 用户相关
		user := authorized.Group("/user")
		{
			user.GET("/profile", api.GetUserProfile)
			user.PUT("/profile/update", api.UpdateUserProfile)
			user.POST("/wx/update-info", api.UpdateWXUserInfo)
			user.GET("/token/expire-time", api.GetTokenExpireTime)

			// 用户反馈
			user.GET("/feedback", api.GetUserFeedbacks)    // 获取用户的反馈列表
			user.POST("/feedback", api.CreateUserFeedback) // 提交反馈
		}

		// 课程相关
		course := authorized.Group("/courses")
		{
			course.GET("", api.GetCourseCategories)
			course.GET("/category/:id", api.GetCategoryDetail)
			course.GET("/:id", api.GetCourseDetail)
			course.GET("/:id/exam", api.GetCourseExam)
			course.POST("/:id/exam/submit", api.SubmitCourseExam)
		}

		// 题目相关
		question := authorized.Group("/questions")
		{
			question.GET("/:course_id", api.GetCourseQuestions)
		}

		// 练习相关
		practice := authorized.Group("/practice")
		{
			practice.GET("/wrong-questions", api.GetWrongQuestionsStats)
			practice.GET("/wrong-questions/:course_id", api.GetWrongQuestionsByCourse)
			practice.POST("/submit", api.SubmitPractice)
		}

		// 考试相关
		exam := authorized.Group("/exams")
		{
			exam.GET("/result", api.GetExamResults)
		}

		// 订单相关
		order := authorized.Group("/orders")
		{
			order.POST("", api.CreateOrder)
			order.GET("", api.GetOrders)
			order.GET("/:id", api.GetOrderDetail)
		}
	}
}

// setupAdminAPIRoutes 设置管理员API路由
func setupAdminAPIRoutes(r *gin.Engine) {
	// 管理端 API 路由
	adminGroup := r.Group("/api/v1/admin")
	adminGroup.Use(middleware.Logger())
	adminGroup.Use(middleware.Recovery())
	adminGroup.Use(middleware.Cors())

	// 管理员登录
	adminGroup.POST("/login", admin.Login)

	// 管理员微信登录（不需要认证）
	adminWX := adminGroup.Group("/wechat")
	{
		adminWX.POST("/login", api.WXAdminLogin)                 // 管理员微信小程序登录
		adminWX.GET("/oauth/url", api.GetWXAdminOAuthURL)        // 获取管理员微信网页授权URL
		adminWX.GET("/oauth/callback", api.WXAdminOAuthCallback) // 管理员微信网页授权回调

		// 管理员扫码登录
		adminWX.POST("/qrcode/create", api.CreateAdminLoginQRCode) // 创建管理员登录二维码
		adminWX.GET("/qrcode/check", api.CheckAdminQRCodeStatus)   // 检查管理员二维码状态
		adminWX.GET("/qrcode/callback", api.AdminQRCodeCallback)   // 管理员扫码登录回调
	}

	// 需要管理员权限的路由
	authorized := adminGroup.Group("/")
	authorized.Use(middleware.JWT())
	authorized.Use(middleware.AdminAuth())
	{
		// 系统管理
		system := authorized.Group("/system")
		{
			system.GET("/login-logs", admin.GetLoginLogs)             // 获取登录日志
			system.GET("/sales-statistics", admin.GetSalesStatistics) // 获取销售统计数据
			system.GET("/system-info", admin.GetSystemInfo)           // 获取系统信息统计数据

			// 管理员个人信息
			system.GET("/profile", admin.GetAdminProfile)    // 获取当前管理员个人信息
			system.PUT("/profile", admin.UpdateAdminProfile) // 更新当前管理员个人信息
		}

		// 用户管理
		users := authorized.Group("/users")
		{
			users.GET("", admin.GetUsers)          // 获取用户列表
			users.GET("/:id", admin.GetUser)       // 获取单个用户
			users.POST("", admin.CreateUser)       // 创建用户
			users.PUT("/:id", admin.UpdateUser)    // 更新用户
			users.DELETE("/:id", admin.DeleteUser) // 删除用户

			// 用户反馈管理
			feedback := users.Group("/feedback")
			{
				feedback.GET("", admin.GetAllFeedbacks)       // 获取所有反馈
				feedback.GET("/:id", admin.GetFeedback)       // 获取单个反馈
				feedback.PUT("/:id", admin.UpdateFeedback)    // 更新反馈（如回复、更改状态）
				feedback.DELETE("/:id", admin.DeleteFeedback) // 删除反馈
			}
		}

		// 订单管理
		orders := authorized.Group("/orders")
		{
			orders.GET("", admin.GetOrders)                  // 获取订单列表
			orders.GET("/:id", admin.GetOrder)               // 获取单个订单
			orders.POST("", admin.CreateOrder)               // 创建订单
			orders.PUT("/:id", admin.UpdateOrder)            // 更新订单
			orders.DELETE("/:id", admin.DeleteOrder)         // 删除订单
			orders.POST("/refund", api.RefundPayment)        // 申请退款
			orders.GET("/refund/:order_no", api.QueryRefund) // 查询退款状态
		}

		// 课程管理
		courses := authorized.Group("/courses")
		{
			courses.GET("", admin.GetCourses)          // 获取课程列表
			courses.GET("/:id", admin.GetCourse)       // 获取单个课程
			courses.POST("", admin.CreateCourse)       // 创建课程
			courses.PUT("/:id", admin.UpdateCourse)    // 更新课程
			courses.DELETE("/:id", admin.DeleteCourse) // 删除课程
		}

		// 题库管理
		questions := authorized.Group("/questions")
		{
			questions.GET("", admin.GetQuestions)                       // 获取题目列表
			questions.GET("/:id", admin.GetQuestion)                    // 获取单个题目
			questions.POST("", admin.CreateQuestion)                    // 创建题目
			questions.PUT("/:id", admin.UpdateQuestion)                 // 更新题目
			questions.DELETE("/:id", admin.DeleteQuestion)              // 删除题目
			questions.POST("/batch-delete", admin.BatchDeleteQuestions) // 批量删除题目
			questions.GET("/export", admin.ExportQuestions)             // 导出题库
			questions.POST("/import", admin.ImportQuestions)            // 导入题库
		}

		// 卡券管理
		cards := authorized.Group("/cards")
		{
			cards.GET("", admin.GetCards)                   // 获取卡券列表
			cards.GET("/records", admin.GetAllCardRecords)  // 获取所有卡券兑换记录
			cards.GET("/:id/records", admin.GetCardRecords) // 获取指定卡券的兑换记录
			cards.GET("/:id", admin.GetCard)                // 获取单个卡券
			cards.POST("", admin.CreateCard)                // 创建卡券
			cards.PUT("/:id", admin.UpdateCard)             // 更新卡券
			cards.DELETE("/:id", admin.DeleteCard)          // 删除卡券
		}
	}
}

// setupSPARoutes 设置前端SPA路由
func setupSPARoutes(r *gin.Engine, userFS, adminFS http.FileSystem) {
	// 直接处理管理员页面请求 - 使用GET方法明确匹配路径
	r.GET("/admin", serveAdminIndex(adminFS))
	r.GET("/admin/*path", serveAdminPath(adminFS))

	// 用户前端 - 处理根路径请求，这个放在最后
	r.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path

		// 如果是API请求，直接跳过
		if strings.HasPrefix(path, "/api/") ||
			strings.HasPrefix(path, "/static/") {
			c.Next()
			return
		}

		// 使用用户端静态文件系统
		serveUserFile(c, path, userFS)
	})
}

// serveAdminIndex 提供管理员首页
func serveAdminIndex(adminFS http.FileSystem) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Content-Type", "text/html; charset=utf-8")
		file, err := adminFS.Open("/index.html")
		if err != nil {
			c.String(http.StatusNotFound, "管理员页面不存在")
			return
		}
		defer file.Close()

		http.ServeContent(c.Writer, c.Request, "index.html", time.Now(), file.(io.ReadSeeker))
	}
}

// serveAdminPath 提供管理员其他路径
func serveAdminPath(adminFS http.FileSystem) gin.HandlerFunc {
	fileServer := http.FileServer(adminFS)
	return func(c *gin.Context) {
		path := c.Param("path")

		// 移除前导斜杠
		if path != "" && path[0] == '/' {
			path = path[1:]
		}

		// 检查文件是否存在
		f, err := adminFS.Open(path)
		if err == nil {
			// 文件存在，关闭文件并提供服务
			f.Close()
			c.Request.URL.Path = "/" + path
			fileServer.ServeHTTP(c.Writer, c.Request)
			return
		}

		// 文件不存在，返回index.html
		c.Request.URL.Path = "/"
		fileServer.ServeHTTP(c.Writer, c.Request)
	}
}

// serveUserFile 提供用户端文件
func serveUserFile(c *gin.Context, path string, userFS http.FileSystem) {
	fileServer := http.FileServer(userFS)

	// 检查文件是否存在
	f, err := userFS.Open(path)
	if err == nil {
		// 文件存在，关闭文件并提供服务
		f.Close()
		fileServer.ServeHTTP(c.Writer, c.Request)
		return
	}

	// 文件不存在，返回index.html
	c.Request.URL.Path = "/"
	fileServer.ServeHTTP(c.Writer, c.Request)
}

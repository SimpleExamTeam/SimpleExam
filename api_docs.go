package main

// _apiHealthCheck doc
// @Summary      健康检查
// @Description  简单健康检查，用于Docker健康检查和负载均衡器
// @Tags         系统
// @Success      200  {object}  map[string]string  "{"status":"ok"}"
// @Router       /health [get]
func _apiHealthCheck() {}

// _apiLogin doc
// @Summary      用户登录
// @Description  使用用户名和密码登录
// @Tags         认证
// @Accept       json
// @Produce      json
// @Param        body  body      api.LoginRequest  true  "登录请求"
// @Success      200   {object}  map[string]any    "登录成功"
// @Router       /auth/login [post]
func _apiLogin() {}

// _apiRegister doc
// @Summary      用户注册
// @Description  注册新用户
// @Tags         认证
// @Accept       json
// @Produce      json
// @Param        body  body      api.RegisterRequest  true  "注册请求"
// @Success      200   {object}  map[string]any       "注册成功"
// @Router       /auth/register [post]
func _apiRegister() {}

// _apiWXLogin doc
// @Summary      微信小程序登录
// @Description  使用微信小程序code登录
// @Tags         认证
// @Accept       json
// @Produce      json
// @Param        body  body      api.WXLoginRequest  true  "微信登录请求"
// @Success      200   {object}  map[string]any      "登录成功"
// @Router       /auth/wx/login [post]
func _apiWXLogin() {}

// _apiGetWXOAuthURL doc
// @Summary      获取微信网页授权URL
// @Description  获取微信网页授权URL
// @Tags         微信
// @Produce      json
// @Param        state  query  string  false  "自定义状态参数"
// @Success      200    {object}  map[string]any  "授权URL"
// @Router       /wechat/oauth/url [get]
func _apiGetWXOAuthURL() {}

// _apiWXOAuthCallback doc
// @Summary      微信网页授权回调
// @Description  微信网页授权回调处理
// @Tags         微信
// @Produce      json
// @Param        code   query  string  true  "授权code"
// @Param        state  query  string  false  "自定义状态参数"
// @Success      200    {object}  map[string]any  "登录成功"
// @Router       /wechat/oauth/callback [get]
func _apiWXOAuthCallback() {}

// _apiUpdateWXUserInfo doc
// @Summary      更新微信用户信息
// @Description  更新微信用户信息
// @Tags         用户
// @Accept       json
// @Produce      json
// @Param        body  body      api.WXUserInfoRequest  true  "微信用户信息"
// @Success      200   {object}  map[string]any        "更新成功"
// @Router       /user/wx/update-info [post]
// @Security     BearerAuth
func _apiUpdateWXUserInfo() {}

// _apiCreateLoginQRCode doc
// @Summary      创建登录二维码
// @Description  创建微信扫码登录二维码
// @Tags         微信
// @Produce      json
// @Success      200  {object}  map[string]any  "二维码信息"
// @Router       /wechat/qrcode/create [post]
func _apiCreateLoginQRCode() {}

// _apiCheckQRCodeStatus doc
// @Summary      检查二维码状态
// @Description  检查微信扫码登录二维码状态
// @Tags         微信
// @Produce      json
// @Param        scene_str  query  string  true  "场景值"
// @Success      200        {object}  map[string]any  "二维码状态"
// @Router       /wechat/qrcode/check [get]
func _apiCheckQRCodeStatus() {}

// _apiQRCodeCallback doc
// @Summary      扫码登录回调
// @Description  微信扫码登录回调处理
// @Tags         微信
// @Produce      json
// @Param        code   query  string  true  "授权code"
// @Param        state  query  string  false  "自定义状态参数"
// @Success      200    {object}  map[string]any  "登录成功"
// @Router       /wechat/qrcode/callback [get]
func _apiQRCodeCallback() {}

// _apiGetUserProfile doc
// @Summary      获取用户个人信息
// @Description  获取当前登录用户的个人信息
// @Tags         用户
// @Produce      json
// @Success      200  {object}  map[string]any  "用户信息"
// @Router       /user/profile [get]
// @Security     BearerAuth
func _apiGetUserProfile() {}

// _apiUpdateUserProfile doc
// @Summary      更新用户个人信息
// @Description  更新当前登录用户的个人信息
// @Tags         用户
// @Accept       json
// @Produce      json
// @Param        body  body      api.UpdateProfileRequest  true  "更新信息"
// @Success      200   {object}  map[string]any            "更新成功"
// @Router       /user/profile/update [put]
// @Security     BearerAuth
func _apiUpdateUserProfile() {}

// _apiGetTokenExpireTime doc
// @Summary      获取Token过期时间
// @Description  获取当前用户Token的过期时间
// @Tags         用户
// @Produce      json
// @Success      200  {object}  map[string]any  "过期时间"
// @Router       /user/token/expire-time [get]
// @Security     BearerAuth
func _apiGetTokenExpireTime() {}

// _apiGetUserFeedbacks doc
// @Summary      获取用户反馈列表
// @Description  获取当前用户的反馈列表
// @Tags         反馈
// @Produce      json
// @Param        page  query  int  false  "页码(默认1)"
// @Param        size  query  int  false  "每页条数(默认10)"
// @Success      200   {object}  map[string]any  "反馈列表"
// @Router       /user/feedback [get]
// @Security     BearerAuth
func _apiGetUserFeedbacks() {}

// _apiCreateUserFeedback doc
// @Summary      提交用户反馈
// @Description  提交用户反馈
// @Tags         反馈
// @Accept       json
// @Produce      json
// @Param        body  body      api.CreateFeedbackRequest  true  "反馈内容"
// @Success      200   {object}  map[string]any             "提交成功"
// @Router       /user/feedback [post]
// @Security     BearerAuth
func _apiCreateUserFeedback() {}

// _apiGetCourseCategories doc
// @Summary      获取课程分类列表
// @Description  获取所有课程分类
// @Tags         课程
// @Produce      json
// @Success      200  {object}  map[string]any  "分类列表"
// @Router       /courses [get]
// @Security     BearerAuth
func _apiGetCourseCategories() {}

// _apiGetCategoryDetail doc
// @Summary      获取课程分类详情
// @Description  获取指定课程分类的详情
// @Tags         课程
// @Produce      json
// @Param        id   path  int  true  "分类ID"
// @Success      200  {object}  map[string]any  "分类详情"
// @Router       /courses/category/{id} [get]
// @Security     BearerAuth
func _apiGetCategoryDetail() {}

// _apiGetCourseDetail doc
// @Summary      获取课程详情
// @Description  获取指定课程的详情
// @Tags         课程
// @Produce      json
// @Param        id   path  int  true  "课程ID"
// @Success      200  {object}  map[string]any  "课程详情"
// @Router       /courses/{id} [get]
// @Security     BearerAuth
func _apiGetCourseDetail() {}

// _apiGetCourseExam doc
// @Summary      获取课程考试
// @Description  获取指定课程的考试题目
// @Tags         课程
// @Produce      json
// @Param        id   path  int  true  "课程ID"
// @Success      200  {object}  map[string]any  "考试题目"
// @Router       /courses/{id}/exam [get]
// @Security     BearerAuth
func _apiGetCourseExam() {}

// _apiSubmitCourseExam doc
// @Summary      提交课程考试
// @Description  提交指定课程的考试答案
// @Tags         课程
// @Accept       json
// @Produce      json
// @Param        id    path  int           true  "课程ID"
// @Param        body  body  map[string]any  true  "考试答案"
// @Success      200   {object}  map[string]any  "考试结果"
// @Router       /courses/{id}/exam/submit [post]
// @Security     BearerAuth
func _apiSubmitCourseExam() {}

// _apiGetCourseQuestions doc
// @Summary      获取课程题目
// @Description  获取指定课程的题目列表
// @Tags         题目
// @Produce      json
// @Param        course_id  path   int     true  "课程ID"
// @Param        type       query  string  false  "题目类型(如:single_choice)"
// @Success      200        {object}  map[string]any  "题目列表"
// @Router       /questions/{course_id} [get]
// @Security     BearerAuth
func _apiGetCourseQuestions() {}

// _apiGetWrongQuestionsStats doc
// @Summary      获取错题统计
// @Description  获取错题统计信息
// @Tags         练习
// @Produce      json
// @Success      200  {object}  map[string]any  "错题统计"
// @Router       /practice/wrong-questions [get]
// @Security     BearerAuth
func _apiGetWrongQuestionsStats() {}

// _apiGetWrongQuestionsByCourse doc
// @Summary      获取课程错题
// @Description  获取指定课程的错题列表
// @Tags         练习
// @Produce      json
// @Param        course_id  path  int  true  "课程ID"
// @Success      200        {object}  map[string]any  "错题列表"
// @Router       /practice/wrong-questions/{course_id} [get]
// @Security     BearerAuth
func _apiGetWrongQuestionsByCourse() {}

// _apiClearWrongQuestions doc
// @Summary      清空错题
// @Description  清空当前用户的全部错题记录
// @Tags         练习
// @Produce      json
// @Success      200  {object}  map[string]any  "清空成功"
// @Router       /practice/wrong-questions [delete]
// @Security     BearerAuth
func _apiClearWrongQuestions() {}

// _apiSubmitPractice doc
// @Summary      提交练习
// @Description  提交练习答案
// @Tags         练习
// @Accept       json
// @Produce      json
// @Param        body  body  map[string]any  true  "练习答案"
// @Success      200   {object}  map[string]any  "练习结果"
// @Router       /practice/submit [post]
// @Security     BearerAuth
func _apiSubmitPractice() {}

// _apiGenerateExplanation doc
// @Summary      生成题目解析
// @Description  生成指定题目的AI解析
// @Tags         练习
// @Accept       json
// @Produce      json
// @Param        id    path  int                            true  "题目ID"
// @Param        body  body  api.GenerateExplanationRequest  true  "生成参数"
// @Success      200   {object}  map[string]any              "解析结果"
// @Router       /practice/question/{id}/explanation [post]
// @Security     BearerAuth
func _apiGenerateExplanation() {}

// _apiGetExamResults doc
// @Summary      获取考试结果
// @Description  获取当前用户的考试结果列表
// @Tags         考试
// @Produce      json
// @Success      200  {object}  map[string]any  "考试结果"
// @Router       /exams/result [get]
// @Security     BearerAuth
func _apiGetExamResults() {}

// _apiCreatePayment doc
// @Summary      创建支付
// @Description  创建支付订单
// @Tags         支付
// @Accept       json
// @Produce      json
// @Param        body  body      api.CreatePaymentRequest  true  "支付请求"
// @Success      200   {object}  map[string]any            "支付结果"
// @Router       /payments/create [post]
// @Security     BearerAuth
func _apiCreatePayment() {}

// _apiPaymentNotify doc
// @Summary      支付回调
// @Description  微信支付回调通知处理
// @Tags         支付
// @Accept       json
// @Produce      json
// @Success      200  {object}  map[string]any  "回调处理结果"
// @Router       /payments/notify [post]
func _apiPaymentNotify() {}

// _apiQueryPayment doc
// @Summary      查询支付
// @Description  查询支付订单状态
// @Tags         支付
// @Produce      json
// @Param        order_no  path  string  true  "订单号"
// @Success      200       {object}  map[string]any  "支付状态"
// @Router       /payments/query/{order_no} [get]
// @Security     BearerAuth
func _apiQueryPayment() {}

// _apiCancelPayment doc
// @Summary      取消支付
// @Description  取消支付订单
// @Tags         支付
// @Produce      json
// @Param        order_no  path  string  true  "订单号"
// @Success      200       {object}  map[string]any  "取消成功"
// @Router       /payments/cancel/{order_no} [post]
// @Security     BearerAuth
func _apiCancelPayment() {}

// _apiRedeemCard doc
// @Summary      兑换卡券
// @Description  使用卡券兑换课程
// @Tags         支付
// @Accept       json
// @Produce      json
// @Param        body  body      api.RedeemCardRequest  true  "兑换请求"
// @Success      200   {object}  map[string]any         "兑换成功"
// @Router       /payments/redeem-card [post]
// @Security     BearerAuth
func _apiRedeemCard() {}

// _apiRefundPayment doc
// @Summary      申请退款
// @Description  申请退款
// @Tags         支付
// @Accept       json
// @Produce      json
// @Param        body  body      api.RefundRequest  true  "退款请求"
// @Success      200   {object}  map[string]any     "退款结果"
// @Router       /admin/orders/refund [post]
// @Security     BearerAuth
func _apiRefundPayment() {}

// _apiQueryRefund doc
// @Summary      查询退款
// @Description  查询退款状态
// @Tags         支付
// @Produce      json
// @Param        order_no  path  string  true  "订单号"
// @Success      200       {object}  map[string]any  "退款状态"
// @Router       /admin/orders/refund/{order_no} [get]
// @Security     BearerAuth
func _apiQueryRefund() {}

// _apiRefundNotify doc
// @Summary      退款回调
// @Description  微信退款回调通知处理
// @Tags         支付
// @Accept       xml
// @Produce      xml
// @Success      200  {object}  map[string]any  "回调处理结果"
// @Router       /payments/refund/notify [post]
func _apiRefundNotify() {}

// _apiCreateOrder doc
// @Summary      创建订单
// @Description  创建订单
// @Tags         订单
// @Accept       json
// @Produce      json
// @Param        body  body  map[string]any  true  "创建订单请求"
// @Success      200   {object}  map[string]any  "订单信息"
// @Router       /orders [post]
// @Security     BearerAuth
func _apiCreateOrder() {}

// _apiGetOrders doc
// @Summary      获取订单列表
// @Description  获取当前用户的订单列表
// @Tags         订单
// @Produce      json
// @Success      200  {object}  map[string]any  "订单列表"
// @Router       /orders [get]
// @Security     BearerAuth
func _apiGetOrders() {}

// _apiGetOrderDetail doc
// @Summary      获取订单详情
// @Description  获取订单详情
// @Tags         订单
// @Produce      json
// @Param        id   path  int  true  "订单ID"
// @Success      200  {object}  map[string]any  "订单详情"
// @Router       /orders/{id} [get]
// @Security     BearerAuth
func _apiGetOrderDetail() {}

// _apiWXAdminLogin doc
// @Summary      管理员微信小程序登录
// @Description  管理员使用微信小程序code登录
// @Tags         管理员-认证
// @Accept       json
// @Produce      json
// @Param        body  body      api.WXLoginRequest  true  "微信登录请求"
// @Success      200   {object}  map[string]any      "登录成功"
// @Router       /admin/wechat/login [post]
func _apiWXAdminLogin() {}

// _apiGetWXAdminOAuthURL doc
// @Summary      获取管理员微信网页授权URL
// @Description  获取管理员微信网页授权URL
// @Tags         管理员-微信
// @Produce      json
// @Param        state  query  string  false  "自定义状态参数"
// @Success      200    {object}  map[string]any  "授权URL"
// @Router       /admin/wechat/oauth/url [get]
func _apiGetWXAdminOAuthURL() {}

// _apiWXAdminOAuthCallback doc
// @Summary      管理员微信网页授权回调
// @Description  管理员微信网页授权回调处理
// @Tags         管理员-微信
// @Produce      json
// @Param        code   query  string  true  "授权code"
// @Param        state  query  string  false  "自定义状态参数"
// @Success      200    {object}  map[string]any  "登录成功"
// @Router       /admin/wechat/oauth/callback [get]
func _apiWXAdminOAuthCallback() {}

// _apiCreateAdminLoginQRCode doc
// @Summary      创建管理员登录二维码
// @Description  创建管理员扫码登录二维码
// @Tags         管理员-微信
// @Produce      json
// @Success      200  {object}  map[string]any  "二维码信息"
// @Router       /admin/wechat/qrcode/create [post]
func _apiCreateAdminLoginQRCode() {}

// _apiCheckAdminQRCodeStatus doc
// @Summary      检查管理员二维码状态
// @Description  检查管理员扫码登录二维码状态
// @Tags         管理员-微信
// @Produce      json
// @Param        scene_str  query  string  true  "场景值"
// @Success      200        {object}  map[string]any  "二维码状态"
// @Router       /admin/wechat/qrcode/check [get]
func _apiCheckAdminQRCodeStatus() {}

// _apiAdminQRCodeCallback doc
// @Summary      管理员扫码登录回调
// @Description  管理员扫码登录回调处理
// @Tags         管理员-微信
// @Produce      json
// @Param        code   query  string  true  "授权code"
// @Param        state  query  string  false  "自定义状态参数"
// @Success      200    {object}  map[string]any  "登录成功"
// @Router       /admin/wechat/qrcode/callback [get]
func _apiAdminQRCodeCallback() {}

// _adminLogin doc
// @Summary      管理员登录
// @Description  管理员使用用户名和密码登录
// @Tags         管理员-认证
// @Accept       json
// @Produce      json
// @Param        body  body      admin.LoginRequest  true  "登录请求"
// @Success      200   {object}  map[string]any      "登录成功"
// @Router       /admin/login [post]
func _adminLogin() {}

// _adminGetLoginLogs doc
// @Summary      获取登录日志
// @Description  获取管理员登录日志列表
// @Tags         管理员-系统
// @Produce      json
// @Param        page       query  int     false  "页码(默认1)"
// @Param        size       query  int     false  "每页条数(默认10)"
// @Param        username   query  string  false  "用户名"
// @Param        status     query  int     false  "状态"
// @Param        start_time query  string  false  "开始时间"
// @Param        end_time   query  string  false  "结束时间"
// @Success      200        {object}  map[string]any  "登录日志"
// @Router       /admin/system/login-logs [get]
// @Security     BearerAuth
func _adminGetLoginLogs() {}

// _adminGetSalesStatistics doc
// @Summary      获取销售统计
// @Description  获取销售统计数据
// @Tags         管理员-系统
// @Produce      json
// @Param        dimension  query  string  false  "统计维度(按日/月/年)"
// @Param        start_time query  string  false  "开始时间"
// @Param        end_time   query  string  false  "结束时间"
// @Success      200        {object}  map[string]any  "销售统计"
// @Router       /admin/system/sales-statistics [get]
// @Security     BearerAuth
func _adminGetSalesStatistics() {}

// _adminGetSystemInfo doc
// @Summary      获取系统信息
// @Description  获取系统信息统计数据
// @Tags         管理员-系统
// @Produce      json
// @Success      200  {object}  map[string]any  "系统信息"
// @Router       /admin/system/system-info [get]
// @Security     BearerAuth
func _adminGetSystemInfo() {}

// _adminGetAdminProfile doc
// @Summary      获取管理员个人信息
// @Description  获取当前管理员个人信息
// @Tags         管理员-系统
// @Produce      json
// @Success      200  {object}  map[string]any  "管理员信息"
// @Router       /admin/system/profile [get]
// @Security     BearerAuth
func _adminGetAdminProfile() {}

// _adminUpdateAdminProfile doc
// @Summary      更新管理员个人信息
// @Description  更新当前管理员个人信息
// @Tags         管理员-系统
// @Accept       json
// @Produce      json
// @Param        body  body      admin.UpdateProfileRequest  true  "更新信息"
// @Success      200   {object}  map[string]any              "更新成功"
// @Router       /admin/system/profile [put]
// @Security     BearerAuth
func _adminUpdateAdminProfile() {}

// _adminGetUsers doc
// @Summary      获取用户列表
// @Description  获取用户列表(管理员)
// @Tags         管理员-用户管理
// @Produce      json
// @Param        page     query  int     false  "页码(默认1)"
// @Param        size     query  int     false  "每页条数(默认10)"
// @Param        keyword  query  string  false  "搜索关键词"
// @Param        is_admin query  bool    false  "是否管理员"
// @Success      200      {object}  map[string]any  "用户列表"
// @Router       /admin/users [get]
// @Security     BearerAuth
func _adminGetUsers() {}

// _adminGetUser doc
// @Summary      获取单个用户
// @Description  获取单个用户详情(管理员)
// @Tags         管理员-用户管理
// @Produce      json
// @Param        id   path  int  true  "用户ID"
// @Success      200  {object}  map[string]any  "用户详情"
// @Router       /admin/users/{id} [get]
// @Security     BearerAuth
func _adminGetUser() {}

// _adminCreateUser doc
// @Summary      创建用户
// @Description  创建用户(管理员)
// @Tags         管理员-用户管理
// @Accept       json
// @Produce      json
// @Param        body  body      admin.CreateUserRequest  true  "创建用户请求"
// @Success      200   {object}  map[string]any           "创建成功"
// @Router       /admin/users [post]
// @Security     BearerAuth
func _adminCreateUser() {}

// _adminUpdateUser doc
// @Summary      更新用户
// @Description  更新用户信息(管理员)
// @Tags         管理员-用户管理
// @Accept       json
// @Produce      json
// @Param        id    path  int                    true  "用户ID"
// @Param        body  body  admin.UpdateUserRequest  true  "更新用户请求"
// @Success      200   {object}  map[string]any     "更新成功"
// @Router       /admin/users/{id} [put]
// @Security     BearerAuth
func _adminUpdateUser() {}

// _adminDeleteUser doc
// @Summary      删除用户
// @Description  删除用户(管理员)
// @Tags         管理员-用户管理
// @Produce      json
// @Param        id   path  int  true  "用户ID"
// @Success      200  {object}  map[string]any  "删除成功"
// @Router       /admin/users/{id} [delete]
// @Security     BearerAuth
func _adminDeleteUser() {}

// _adminGetAllFeedbacks doc
// @Summary      获取所有反馈
// @Description  获取所有用户反馈(管理员)
// @Tags         管理员-反馈管理
// @Produce      json
// @Param        page       query  int     false  "页码(默认1)"
// @Param        size       query  int     false  "每页条数(默认10)"
// @Param        username   query  string  false  "用户名"
// @Param        status     query  int     false  "状态"
// @Param        start_time query  string  false  "开始时间"
// @Param        end_time   query  string  false  "结束时间"
// @Success      200        {object}  map[string]any  "反馈列表"
// @Router       /admin/users/feedback [get]
// @Security     BearerAuth
func _adminGetAllFeedbacks() {}

// _adminGetFeedback doc
// @Summary      获取单个反馈
// @Description  获取单个反馈详情(管理员)
// @Tags         管理员-反馈管理
// @Produce      json
// @Param        id   path  int  true  "反馈ID"
// @Success      200  {object}  map[string]any  "反馈详情"
// @Router       /admin/users/feedback/{id} [get]
// @Security     BearerAuth
func _adminGetFeedback() {}

// _adminUpdateFeedback doc
// @Summary      更新反馈
// @Description  更新反馈(管理员回复等)
// @Tags         管理员-反馈管理
// @Accept       json
// @Produce      json
// @Param        id    path  int                          true  "反馈ID"
// @Param        body  body  admin.UpdateFeedbackRequest  true  "更新反馈请求"
// @Success      200   {object}  map[string]any           "更新成功"
// @Router       /admin/users/feedback/{id} [put]
// @Security     BearerAuth
func _adminUpdateFeedback() {}

// _adminDeleteFeedback doc
// @Summary      删除反馈
// @Description  删除反馈(管理员)
// @Tags         管理员-反馈管理
// @Produce      json
// @Param        id   path  int  true  "反馈ID"
// @Success      200  {object}  map[string]any  "删除成功"
// @Router       /admin/users/feedback/{id} [delete]
// @Security     BearerAuth
func _adminDeleteFeedback() {}

// _adminGetOrders doc
// @Summary      获取订单列表(管理员)
// @Description  获取所有订单列表(管理员)
// @Tags         管理员-订单管理
// @Produce      json
// @Param        page         query  int     false  "页码(默认1)"
// @Param        size         query  int     false  "每页条数(默认10)"
// @Param        order_no     query  string  false  "订单号"
// @Param        username     query  string  false  "用户名"
// @Param        user_id      query  int     false  "用户ID"
// @Param        status       query  string  false  "状态"
// @Param        payment_type  query string  false  "支付类型"
// @Param        start_time   query  string  false  "开始时间"
// @Param        end_time     query  string  false  "结束时间"
// @Success      200          {object}  map[string]any  "订单列表"
// @Router       /admin/orders [get]
// @Security     BearerAuth
func _adminGetOrders() {}

// _adminGetOrder doc
// @Summary      获取单个订单
// @Description  获取单个订单详情(管理员)
// @Tags         管理员-订单管理
// @Produce      json
// @Param        id   path  int  true  "订单ID"
// @Success      200  {object}  map[string]any  "订单详情"
// @Router       /admin/orders/{id} [get]
// @Security     BearerAuth
func _adminGetOrder() {}

// _adminCreateOrder doc
// @Summary      创建订单(管理员)
// @Description  管理员创建订单
// @Tags         管理员-订单管理
// @Accept       json
// @Produce      json
// @Param        body  body      admin.CreateOrderRequest  true  "创建订单请求"
// @Success      200   {object}  map[string]any            "创建成功"
// @Router       /admin/orders [post]
// @Security     BearerAuth
func _adminCreateOrder() {}

// _adminUpdateOrder doc
// @Summary      更新订单
// @Description  更新订单信息(管理员)
// @Tags         管理员-订单管理
// @Accept       json
// @Produce      json
// @Param        id    path  int                      true  "订单ID"
// @Param        body  body  admin.UpdateOrderRequest  true  "更新订单请求"
// @Success      200   {object}  map[string]any       "更新成功"
// @Router       /admin/orders/{id} [put]
// @Security     BearerAuth
func _adminUpdateOrder() {}

// _adminDeleteOrder doc
// @Summary      删除订单
// @Description  删除订单(管理员)
// @Tags         管理员-订单管理
// @Produce      json
// @Param        id   path  int  true  "订单ID"
// @Success      200  {object}  map[string]any  "删除成功"
// @Router       /admin/orders/{id} [delete]
// @Security     BearerAuth
func _adminDeleteOrder() {}

// _adminGetCourses doc
// @Summary      获取课程列表(管理员)
// @Description  获取课程列表(管理员)
// @Tags         管理员-课程管理
// @Produce      json
// @Param        page     query  int     false  "页码(默认1)"
// @Param        size     query  int     false  "每页条数(默认10)"
// @Param        keyword  query  string  false  "搜索关键词"
// @Success      200      {object}  map[string]any  "课程列表"
// @Router       /admin/courses [get]
// @Security     BearerAuth
func _adminGetCourses() {}

// _adminGetCourse doc
// @Summary      获取单个课程
// @Description  获取单个课程详情(管理员)
// @Tags         管理员-课程管理
// @Produce      json
// @Param        id   path  int  true  "课程ID"
// @Success      200  {object}  map[string]any  "课程详情"
// @Router       /admin/courses/{id} [get]
// @Security     BearerAuth
func _adminGetCourse() {}

// _adminCreateCourse doc
// @Summary      创建课程
// @Description  创建课程(管理员)
// @Tags         管理员-课程管理
// @Accept       json
// @Produce      json
// @Param        body  body      admin.CreateCourseRequest  true  "创建课程请求"
// @Success      200   {object}  map[string]any             "创建成功"
// @Router       /admin/courses [post]
// @Security     BearerAuth
func _adminCreateCourse() {}

// _adminUpdateCourse doc
// @Summary      更新课程
// @Description  更新课程信息(管理员)
// @Tags         管理员-课程管理
// @Accept       json
// @Produce      json
// @Param        id    path  int                      true  "课程ID"
// @Param        body  body  admin.UpdateCourseRequest  true  "更新课程请求"
// @Success      200   {object}  map[string]any       "更新成功"
// @Router       /admin/courses/{id} [put]
// @Security     BearerAuth
func _adminUpdateCourse() {}

// _adminDeleteCourse doc
// @Summary      删除课程
// @Description  删除课程(管理员)
// @Tags         管理员-课程管理
// @Produce      json
// @Param        id   path  int  true  "课程ID"
// @Success      200  {object}  map[string]any  "删除成功"
// @Router       /admin/courses/{id} [delete]
// @Security     BearerAuth
func _adminDeleteCourse() {}

// _adminGetQuestions doc
// @Summary      获取题目列表
// @Description  获取题目列表(管理员)
// @Tags         管理员-题库管理
// @Produce      json
// @Param        page      query  int     false  "页码(默认1)"
// @Param        size      query  int     false  "每页条数(默认10)"
// @Param        type      query  string  false  "题目类型"
// @Param        question  query  string  false  "题目内容搜索"
// @Param        course_id query  int     false  "课程ID"
// @Success      200       {object}  map[string]any  "题目列表"
// @Router       /admin/questions [get]
// @Security     BearerAuth
func _adminGetQuestions() {}

// _adminGetQuestion doc
// @Summary      获取单个题目
// @Description  获取单个题目详情(管理员)
// @Tags         管理员-题库管理
// @Produce      json
// @Param        id   path  int  true  "题目ID"
// @Success      200  {object}  map[string]any  "题目详情"
// @Router       /admin/questions/{id} [get]
// @Security     BearerAuth
func _adminGetQuestion() {}

// _adminCreateQuestion doc
// @Summary      创建题目
// @Description  创建题目(管理员)
// @Tags         管理员-题库管理
// @Accept       json
// @Produce      json
// @Param        body  body      admin.QuestionRequest  true  "创建题目请求"
// @Success      200   {object}  map[string]any         "创建成功"
// @Router       /admin/questions [post]
// @Security     BearerAuth
func _adminCreateQuestion() {}

// _adminUpdateQuestion doc
// @Summary      更新题目
// @Description  更新题目(管理员)
// @Tags         管理员-题库管理
// @Accept       json
// @Produce      json
// @Param        id    path  int                    true  "题目ID"
// @Param        body  body  admin.QuestionRequest  true  "更新题目请求"
// @Success      200   {object}  map[string]any     "更新成功"
// @Router       /admin/questions/{id} [put]
// @Security     BearerAuth
func _adminUpdateQuestion() {}

// _adminDeleteQuestion doc
// @Summary      删除题目
// @Description  删除题目(管理员)
// @Tags         管理员-题库管理
// @Produce      json
// @Param        id   path  int  true  "题目ID"
// @Success      200  {object}  map[string]any  "删除成功"
// @Router       /admin/questions/{id} [delete]
// @Security     BearerAuth
func _adminDeleteQuestion() {}

// _adminBatchDeleteQuestions doc
// @Summary      批量删除题目
// @Description  批量删除题目(管理员)
// @Tags         管理员-题库管理
// @Accept       json
// @Produce      json
// @Param        body  body  map[string]any  true  "批量删除请求(IDs)"
// @Success      200   {object}  map[string]any  "删除成功"
// @Router       /admin/questions/batch-delete [post]
// @Security     BearerAuth
func _adminBatchDeleteQuestions() {}

// _adminClearQuestionsByCourse doc
// @Summary      清空课程题目
// @Description  一键清空指定课程的全部题目
// @Tags         管理员-题库管理
// @Produce      json
// @Param        course_id  path  int  true  "课程ID"
// @Success      200        {object}  map[string]any  "清空成功"
// @Router       /admin/questions/clear-by-course/{course_id} [delete]
// @Security     BearerAuth
func _adminClearQuestionsByCourse() {}

// _adminExportQuestions doc
// @Summary      导出题库
// @Description  导出指定课程的题库(CSV)
// @Tags         管理员-题库管理
// @Produce      text/csv
// @Param        course_id  query  int  false  "课程ID"
// @Success      200        {file}   string  "CSV文件"
// @Router       /admin/questions/export [get]
// @Security     BearerAuth
func _adminExportQuestions() {}

// _adminImportQuestions doc
// @Summary      导入题库
// @Description  从CSV文件导入题库
// @Tags         管理员-题库管理
// @Accept       multipart/form-data
// @Produce      json
// @Param        file  formData  file  true  "CSV文件"
// @Success      200   {object}  map[string]any  "导入结果"
// @Router       /admin/questions/import [post]
// @Security     BearerAuth
func _adminImportQuestions() {}

// _adminGetCards doc
// @Summary      获取卡券列表
// @Description  获取卡券列表(管理员)
// @Tags         管理员-卡券管理
// @Produce      json
// @Param        page      query  int     false  "页码(默认1)"
// @Param        size      query  int     false  "每页条数(默认10)"
// @Param        card_no   query  string  false  "卡券号"
// @Param        course_id query  int     false  "课程ID"
// @Success      200       {object}  map[string]any  "卡券列表"
// @Router       /admin/cards [get]
// @Security     BearerAuth
func _adminGetCards() {}

// _adminGetCard doc
// @Summary      获取单个卡券
// @Description  获取单个卡券详情(管理员)
// @Tags         管理员-卡券管理
// @Produce      json
// @Param        id   path  int  true  "卡券ID"
// @Success      200  {object}  map[string]any  "卡券详情"
// @Router       /admin/cards/{id} [get]
// @Security     BearerAuth
func _adminGetCard() {}

// _adminCreateCard doc
// @Summary      创建卡券
// @Description  创建卡券(管理员)
// @Tags         管理员-卡券管理
// @Accept       json
// @Produce      json
// @Param        body  body      admin.CreateCardRequest  true  "创建卡券请求"
// @Success      200   {object}  map[string]any           "创建成功"
// @Router       /admin/cards [post]
// @Security     BearerAuth
func _adminCreateCard() {}

// _adminUpdateCard doc
// @Summary      更新卡券
// @Description  更新卡券(管理员)
// @Tags         管理员-卡券管理
// @Accept       json
// @Produce      json
// @Param        id    path  int                    true  "卡券ID"
// @Param        body  body  admin.UpdateCardRequest  true  "更新卡券请求"
// @Success      200   {object}  map[string]any     "更新成功"
// @Router       /admin/cards/{id} [put]
// @Security     BearerAuth
func _adminUpdateCard() {}

// _adminDeleteCard doc
// @Summary      删除卡券
// @Description  删除卡券(管理员)
// @Tags         管理员-卡券管理
// @Produce      json
// @Param        id   path  int  true  "卡券ID"
// @Success      200  {object}  map[string]any  "删除成功"
// @Router       /admin/cards/{id} [delete]
// @Security     BearerAuth
func _adminDeleteCard() {}

// _adminGetAllCardRecords doc
// @Summary      获取所有卡券兑换记录
// @Description  获取所有卡券兑换记录(管理员)
// @Tags         管理员-卡券管理
// @Produce      json
// @Param        page      query  int     false  "页码(默认1)"
// @Param        size      query  int     false  "每页条数(默认10)"
// @Param        card_no   query  string  false  "卡券号"
// @Param        username  query  string  false  "用户名"
// @Param        course_id query  int     false  "课程ID"
// @Success      200       {object}  map[string]any  "兑换记录"
// @Router       /admin/cards/records [get]
// @Security     BearerAuth
func _adminGetAllCardRecords() {}

// _adminGetCardRecords doc
// @Summary      获取指定卡券的兑换记录
// @Description  获取指定卡券的兑换记录(管理员)
// @Tags         管理员-卡券管理
// @Produce      json
// @Param        id   path  int  true  "卡券ID"
// @Success      200  {object}  map[string]any  "兑换记录"
// @Router       /admin/cards/{id}/records [get]
// @Security     BearerAuth
func _adminGetCardRecords() {}

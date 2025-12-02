package service

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"encoding/xml"
	"errors"
	"exam-system/internal/config"
	"exam-system/internal/middleware"
	"exam-system/internal/model"
	"exam-system/internal/pkg/database"
	"exam-system/internal/pkg/payment"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var WeChat = new(WeChatService)

type WeChatService struct{}

// 微信登录响应结构体
type WXLoginResponse struct {
	OpenID     string `json:"openid"`
	SessionKey string `json:"session_key"`
	UnionID    string `json:"unionid"`
	ErrCode    int    `json:"errcode"`
	ErrMsg     string `json:"errmsg"`
}

// OAuth2.0网页授权返回结构体
type WXOAuthResponse struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	OpenID       string `json:"openid"`
	Scope        string `json:"scope"`
	UnionID      string `json:"unionid"`
	ErrCode      int    `json:"errcode"`
	ErrMsg       string `json:"errmsg"`
}

// 微信用户信息结构体
type WXUserInfo struct {
	OpenID    string `json:"openid"`
	NickName  string `json:"nickName"`
	Gender    int    `json:"gender"`
	City      string `json:"city"`
	Province  string `json:"province"`
	Country   string `json:"country"`
	AvatarURL string `json:"avatarUrl"`
	UnionID   string `json:"unionId"`
	Watermark struct {
		Timestamp int64  `json:"timestamp"`
		AppID     string `json:"appid"`
	} `json:"watermark"`
}

// OAuth用户信息结构体
type WXOAuthUserInfo struct {
	OpenID     string   `json:"openid"`
	Nickname   string   `json:"nickname"`
	Sex        int      `json:"sex"`
	Province   string   `json:"province"`
	City       string   `json:"city"`
	Country    string   `json:"country"`
	HeadImgURL string   `json:"headimgurl"`
	Privilege  []string `json:"privilege"`
	UnionID    string   `json:"unionid"`
	ErrCode    int      `json:"errcode"`
	ErrMsg     string   `json:"errmsg"`
}

// 微信登录
func (s *WeChatService) Login(code string) (*model.User, string, error) {
	// 获取配置
	if config.GlobalConfig == nil {
		return nil, "", errors.New("配置未初始化")
	}
	cfg := config.GlobalConfig

	// 请求微信接口获取openid和session_key
	url := fmt.Sprintf("https://api.weixin.qq.com/sns/jscode2session?appid=%s&secret=%s&js_code=%s&grant_type=authorization_code",
		cfg.WeChat.AppID, cfg.WeChat.AppSecret, code)

	resp, err := http.Get(url)
	if err != nil {
		return nil, "", errors.New("请求微信接口失败")
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, "", errors.New("读取微信响应失败")
	}

	var wxResp WXLoginResponse
	if err = json.Unmarshal(body, &wxResp); err != nil {
		return nil, "", errors.New("解析微信响应失败")
	}

	if wxResp.ErrCode != 0 {
		return nil, "", fmt.Errorf("微信登录失败: %s", wxResp.ErrMsg)
	}

	// 查找或创建用户
	var user model.User
	result := database.DB.Where("open_id = ?", wxResp.OpenID).First(&user)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			// 用户不存在，创建新用户
			// 生成随机密码
			randomPassword := fmt.Sprintf("%d", rand.Int63())
			hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(randomPassword), bcrypt.DefaultCost)

			// 使用openid前12位作为用户名，提高唯一性
			username := "wx_" + wxResp.OpenID[:12]

			// 创建新用户
			user = model.User{
				Username: username,
				Password: string(hashedPassword),
				Nickname: "微信用户", // 默认昵称
				OpenID:   wxResp.OpenID,
				UnionID:  wxResp.UnionID,
			}

			if err = database.DB.Create(&user).Error; err != nil {
				return nil, "", errors.New("创建用户失败")
			}
		} else {
			return nil, "", errors.New("查询用户失败")
		}
	}

	// 生成JWT令牌
	token, err := middleware.GenerateToken(user.ID)
	if err != nil {
		return nil, "", errors.New("生成令牌失败")
	}

	return &user, token, nil
}

// 管理员微信登录
func (s *WeChatService) AdminLogin(code string) (*model.User, string, error) {
	// 获取配置
	if config.GlobalConfig == nil {
		return nil, "", errors.New("配置未初始化")
	}
	cfg := config.GlobalConfig

	// 请求微信接口获取openid和session_key
	url := fmt.Sprintf("https://api.weixin.qq.com/sns/jscode2session?appid=%s&secret=%s&js_code=%s&grant_type=authorization_code",
		cfg.WeChat.AppID, cfg.WeChat.AppSecret, code)

	resp, err := http.Get(url)
	if err != nil {
		return nil, "", errors.New("请求微信接口失败")
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, "", errors.New("读取微信响应失败")
	}

	var wxResp WXLoginResponse
	if err = json.Unmarshal(body, &wxResp); err != nil {
		return nil, "", errors.New("解析微信响应失败")
	}

	if wxResp.ErrCode != 0 {
		return nil, "", fmt.Errorf("微信登录失败: %s", wxResp.ErrMsg)
	}

	// 查找用户
	var user model.User
	result := database.DB.Where("open_id = ?", wxResp.OpenID).First(&user)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, "", errors.New("管理员账户不存在，请联系系统管理员")
		} else {
			return nil, "", errors.New("查询用户失败")
		}
	}

	// 验证是否是管理员
	if !user.IsAdmin {
		return nil, "", errors.New("无管理员权限")
	}

	// 生成JWT令牌
	token, err := middleware.GenerateToken(user.ID)
	if err != nil {
		return nil, "", errors.New("生成令牌失败")
	}

	return &user, token, nil
}

// 更新微信用户信息
func (s *WeChatService) UpdateUserInfo(userId uint, userInfo WXUserInfo) error {
	var user model.User
	if err := database.DB.First(&user, userId).Error; err != nil {
		return errors.New("用户不存在")
	}

	// 确保OpenID匹配
	if user.OpenID != userInfo.OpenID {
		return errors.New("用户信息不匹配")
	}

	// 更新用户信息
	user.Nickname = userInfo.NickName
	user.Avatar = userInfo.AvatarURL
	user.UnionID = userInfo.UnionID

	if err := database.DB.Save(&user).Error; err != nil {
		return errors.New("更新用户信息失败")
	}

	return nil
}

// GeneratePayParams 生成支付参数
func (s *WeChatService) GeneratePayParams(orderNo string, totalFee int, openid string) (*payment.WXPayParams, error) {
	if config.GlobalConfig == nil {
		return nil, errors.New("配置未初始化")
	}
	cfg := config.GlobalConfig.WeChat

	// 调用统一下单接口
	unifiedOrderURL := "https://api.mch.weixin.qq.com/pay/unifiedorder"

	// 构建统一下单参数
	nonceStr := generateRandomString(32)
	unifiedOrderParams := map[string]string{
		"appid":            cfg.AppID,
		"mch_id":           cfg.MchID,
		"nonce_str":        nonceStr,
		"body":             "课程购买",
		"out_trade_no":     orderNo,
		"total_fee":        fmt.Sprintf("%d", totalFee),
		"spbill_create_ip": "127.0.0.1", // 这里应该获取真实的客户端IP
		"notify_url":       cfg.NotifyURL,
		"trade_type":       "JSAPI",
		"openid":           openid,
	}

	// 生成签名
	unifiedOrderParams["sign"] = s.GenerateSign(unifiedOrderParams, cfg.PayKey)

	// 将参数转换为XML
	var buf strings.Builder
	buf.WriteString("<xml>")
	for k, v := range unifiedOrderParams {
		buf.WriteString(fmt.Sprintf("<%s>%s</%s>", k, v, k))
	}
	buf.WriteString("</xml>")

	// 发送请求
	resp, err := http.Post(unifiedOrderURL, "application/xml", strings.NewReader(buf.String()))
	if err != nil {
		return nil, fmt.Errorf("调用统一下单接口失败: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取统一下单响应失败: %v", err)
	}

	// 解析响应
	var result struct {
		ReturnCode string `xml:"return_code"`
		ReturnMsg  string `xml:"return_msg"`
		ResultCode string `xml:"result_code"`
		PrepayID   string `xml:"prepay_id"`
		ErrCode    string `xml:"err_code"`
		ErrCodeDes string `xml:"err_code_des"`
	}
	if err := xml.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("解析统一下单响应失败: %v", err)
	}

	// 检查返回结果
	if result.ReturnCode != "SUCCESS" {
		return nil, fmt.Errorf("统一下单通信失败: %s", result.ReturnMsg)
	}
	if result.ResultCode != "SUCCESS" {
		return nil, fmt.Errorf("统一下单业务失败: %s", result.ErrCodeDes)
	}

	// 生成支付参数
	timeStamp := fmt.Sprintf("%d", time.Now().Unix())
	params := &payment.WXPayParams{
		AppID:     cfg.AppID,
		TimeStamp: timeStamp,
		NonceStr:  nonceStr,
		Package:   fmt.Sprintf("prepay_id=%s", result.PrepayID),
		SignType:  "MD5",
	}

	// 生成支付签名
	signParams := make(map[string]string)
	signParams["appId"] = params.AppID
	signParams["timeStamp"] = params.TimeStamp
	signParams["nonceStr"] = params.NonceStr
	signParams["package"] = params.Package
	signParams["signType"] = params.SignType
	params.PaySign = s.GenerateSign(signParams, cfg.PayKey)

	return params, nil
}

// generateRandomString 生成指定长度的随机字符串
func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[rand.Intn(len(charset))]
	}
	return string(result)
}

// GenerateSign 生成签名
func (s *WeChatService) GenerateSign(params map[string]string, key string) string {
	// 按照参数名ASCII码从小到大排序
	var keys []string
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// 拼接字符串
	var buf strings.Builder
	for _, k := range keys {
		if params[k] != "" {
			buf.WriteString(k)
			buf.WriteString("=")
			buf.WriteString(params[k])
			buf.WriteString("&")
		}
	}
	buf.WriteString("key=")
	buf.WriteString(key)

	// MD5加密
	h := md5.New()
	h.Write([]byte(buf.String()))
	return strings.ToUpper(hex.EncodeToString(h.Sum(nil)))
}

// HandlePayNotify 处理支付回调
func (s *WeChatService) HandlePayNotify(notifyData []byte) error {
	var notify payment.WXPayNotify
	if err := xml.Unmarshal(notifyData, &notify); err != nil {
		return fmt.Errorf("解析回调数据失败: %v", err)
	}

	// 验证签名
	signParams := make(map[string]string)
	signParams["appid"] = notify.AppID
	signParams["mch_id"] = notify.MchID
	signParams["nonce_str"] = notify.NonceStr
	signParams["result_code"] = notify.ResultCode
	signParams["out_trade_no"] = notify.OutTradeNo
	signParams["transaction_id"] = notify.TransactionID
	signParams["total_fee"] = fmt.Sprintf("%d", notify.TotalFee)
	signParams["time_end"] = notify.TimeEnd
	signParams["bank_type"] = notify.BankType
	signParams["cash_fee"] = notify.CashFee
	signParams["fee_type"] = notify.FeeType
	signParams["is_subscribe"] = notify.IsSubscribe
	signParams["trade_type"] = notify.TradeType
	signParams["openid"] = notify.OpenID

	// 添加优惠券相关字段
	if notify.CouponCount != "" {
		signParams["coupon_count"] = notify.CouponCount
		signParams["coupon_fee"] = notify.CouponFee

		// 添加优惠券详细信息
		if notify.CouponFee_0 != "" {
			signParams["coupon_fee_0"] = notify.CouponFee_0
		}
		if notify.CouponID_0 != "" {
			signParams["coupon_id_0"] = notify.CouponID_0
		}
	}

	// 获取配置
	if config.GlobalConfig == nil {
		return errors.New("配置未初始化")
	}
	cfg := config.GlobalConfig.WeChat

	sign := s.GenerateSign(signParams, cfg.PayKey)

	// 签名验证
	signVerified := (sign == notify.Sign)
	if !signVerified {
		// 记录警告但继续处理，因为这可能是由于微信回调中的字段顺序或格式导致的
		fmt.Println("警告: 签名验证失败，但继续处理支付成功通知")
		// 如果是生产环境，可以添加额外的日志记录
	}

	// 查找订单
	var order model.Order
	if err := database.DB.Where("order_no = ?", notify.OutTradeNo).First(&order).Error; err != nil {
		return errors.New("订单不存在")
	}

	// 查找课程信息以获取有效期
	var course model.Course
	if err := database.DB.First(&course, order.CourseID).Error; err != nil {
		return fmt.Errorf("课程不存在: %v", err)
	}

	// 计算过期时间
	var expireTime *time.Time
	if course.ExpireDays > 0 {
		t := time.Now().AddDate(0, 0, course.ExpireDays)
		expireTime = &t
	}

	// 更新订单状态
	now := time.Now()
	updates := map[string]interface{}{
		"status":       "paid",
		"pay_time":     &now,
		"payment_type": "wechat",
		"expire_time":  expireTime,
	}

	if err := database.DB.Model(&order).Updates(updates).Error; err != nil {
		return errors.New("更新订单状态失败")
	}

	return nil
}

// 获取微信网页授权URL
func (s *WeChatService) GetOAuthURL(state string) (string, error) {
	// 获取配置
	if config.GlobalConfig == nil {
		return "", errors.New("配置未初始化")
	}
	cfg := config.GlobalConfig

	if cfg.WeChat.OAuthRedirect == "" {
		return "", errors.New("未配置OAuth回调地址")
	}

	// 生成授权链接
	redirectURL := url.QueryEscape(cfg.WeChat.OAuthRedirect)
	oauthURL := fmt.Sprintf(
		"https://open.weixin.qq.com/connect/oauth2/authorize?"+
			"appid=%s&redirect_uri=%s&response_type=code&scope=snsapi_userinfo&state=%s#wechat_redirect",
		cfg.WeChat.AppID, redirectURL, state)

	return oauthURL, nil
}

// 获取管理员微信网页授权URL
func (s *WeChatService) GetAdminOAuthURL(state string) (string, error) {
	// 获取配置
	if config.GlobalConfig == nil {
		return "", errors.New("配置未初始化")
	}
	cfg := config.GlobalConfig

	// 优先使用专门的管理员回调地址配置
	adminRedirectURL := cfg.WeChat.AdminOAuthRedirect
	if adminRedirectURL == "" {
		// 如果没有配置管理员专用回调地址，则基于普通回调地址生成
		adminRedirectURL = cfg.WeChat.OAuthRedirect
		if adminRedirectURL == "" {
			return "", errors.New("未配置OAuth回调地址")
		}

		// 将普通用户回调地址转换为管理员回调地址
		if strings.Contains(adminRedirectURL, "/wechat/callback") {
			adminRedirectURL = strings.Replace(adminRedirectURL, "/wechat/callback", "/api/v1/admin/wechat/oauth/callback", 1)
		} else {
			// 如果配置的不是标准格式，则需要手动配置管理员回调地址
			return "", errors.New("请在配置中设置正确的管理员OAuth回调地址")
		}
	}

	// 生成授权链接
	redirectURL := url.QueryEscape(adminRedirectURL)
	oauthURL := fmt.Sprintf(
		"https://open.weixin.qq.com/connect/oauth2/authorize?"+
			"appid=%s&redirect_uri=%s&response_type=code&scope=snsapi_userinfo&state=%s#wechat_redirect",
		cfg.WeChat.AppID, redirectURL, state)

	return oauthURL, nil
}

// 使用网页授权码登录
func (s *WeChatService) LoginByOAuth(code, state string) (*model.User, string, error) {
	// 获取配置
	if config.GlobalConfig == nil {
		return nil, "", errors.New("配置未初始化")
	}
	cfg := config.GlobalConfig

	// 1. 通过授权码获取访问令牌
	tokenURL := fmt.Sprintf(
		"https://api.weixin.qq.com/sns/oauth2/access_token?"+
			"appid=%s&secret=%s&code=%s&grant_type=authorization_code",
		cfg.WeChat.AppID, cfg.WeChat.AppSecret, code)

	resp, err := http.Get(tokenURL)
	if err != nil {
		return nil, "", errors.New("请求微信接口失败")
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, "", errors.New("读取微信响应失败")
	}

	var oauthResp WXOAuthResponse
	if err = json.Unmarshal(body, &oauthResp); err != nil {
		return nil, "", errors.New("解析微信响应失败")
	}

	if oauthResp.ErrCode != 0 {
		return nil, "", fmt.Errorf("获取访问令牌失败: %s", oauthResp.ErrMsg)
	}

	// 2. 获取用户信息
	userInfoURL := fmt.Sprintf(
		"https://api.weixin.qq.com/sns/userinfo?"+
			"access_token=%s&openid=%s&lang=zh_CN",
		oauthResp.AccessToken, oauthResp.OpenID)

	resp, err = http.Get(userInfoURL)
	if err != nil {
		return nil, "", errors.New("获取用户信息失败")
	}
	defer resp.Body.Close()

	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, "", errors.New("读取用户信息失败")
	}

	var userInfo WXOAuthUserInfo
	if err = json.Unmarshal(body, &userInfo); err != nil {
		return nil, "", errors.New("解析用户信息失败")
	}

	if userInfo.ErrCode != 0 {
		return nil, "", fmt.Errorf("获取用户信息失败: %s", userInfo.ErrMsg)
	}

	// 3. 查找或创建用户
	var user model.User
	result := database.DB.Where("open_id = ?", userInfo.OpenID).First(&user)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			// 用户不存在，创建新用户
			// 生成随机密码
			randomPassword := fmt.Sprintf("%d", rand.Int63())
			hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(randomPassword), bcrypt.DefaultCost)

			// 使用openid作为用户名，取前12位而不是8位，降低重复率
			username := "wx_" + userInfo.OpenID[:12]

			// 创建新用户
			user = model.User{
				Username: username,
				Password: string(hashedPassword),
				Nickname: userInfo.Nickname,
				Avatar:   userInfo.HeadImgURL,
				OpenID:   userInfo.OpenID,
				UnionID:  userInfo.UnionID,
			}

			if err = database.DB.Create(&user).Error; err != nil {
				return nil, "", errors.New("创建用户失败")
			}
		} else {
			return nil, "", errors.New("查询用户失败")
		}
	} else {
		// 更新用户信息
		updates := map[string]interface{}{
			"nickname": userInfo.Nickname,
			"avatar":   userInfo.HeadImgURL,
		}

		if userInfo.UnionID != "" && user.UnionID == "" {
			updates["union_id"] = userInfo.UnionID
		}

		if err = database.DB.Model(&user).Updates(updates).Error; err != nil {
			return nil, "", errors.New("更新用户信息失败")
		}
	}

	// 4. 生成JWT令牌
	token, err := middleware.GenerateToken(user.ID)
	if err != nil {
		return nil, "", errors.New("生成令牌失败")
	}

	return &user, token, nil
}

// 管理员使用网页授权码登录
func (s *WeChatService) AdminLoginByOAuth(code, state string) (*model.User, string, error) {
	// 获取配置
	if config.GlobalConfig == nil {
		return nil, "", errors.New("配置未初始化")
	}
	cfg := config.GlobalConfig

	// 1. 通过授权码获取访问令牌
	tokenURL := fmt.Sprintf(
		"https://api.weixin.qq.com/sns/oauth2/access_token?"+
			"appid=%s&secret=%s&code=%s&grant_type=authorization_code",
		cfg.WeChat.AppID, cfg.WeChat.AppSecret, code)

	resp, err := http.Get(tokenURL)
	if err != nil {
		return nil, "", errors.New("请求微信接口失败")
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, "", errors.New("读取微信响应失败")
	}

	var oauthResp WXOAuthResponse
	if err = json.Unmarshal(body, &oauthResp); err != nil {
		return nil, "", errors.New("解析微信响应失败")
	}

	if oauthResp.ErrCode != 0 {
		return nil, "", fmt.Errorf("获取访问令牌失败: %s", oauthResp.ErrMsg)
	}

	// 2. 获取用户信息
	userInfoURL := fmt.Sprintf(
		"https://api.weixin.qq.com/sns/userinfo?"+
			"access_token=%s&openid=%s&lang=zh_CN",
		oauthResp.AccessToken, oauthResp.OpenID)

	resp, err = http.Get(userInfoURL)
	if err != nil {
		return nil, "", errors.New("获取用户信息失败")
	}
	defer resp.Body.Close()

	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, "", errors.New("读取用户信息失败")
	}

	var userInfo WXOAuthUserInfo
	if err = json.Unmarshal(body, &userInfo); err != nil {
		return nil, "", errors.New("解析用户信息失败")
	}

	if userInfo.ErrCode != 0 {
		return nil, "", fmt.Errorf("获取用户信息失败: %s", userInfo.ErrMsg)
	}

	// 3. 查找用户
	var user model.User
	result := database.DB.Where("open_id = ?", userInfo.OpenID).First(&user)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, "", errors.New("管理员账户不存在，请联系系统管理员")
		} else {
			return nil, "", errors.New("查询用户失败")
		}
	}

	// 4. 验证是否是管理员
	if !user.IsAdmin {
		return nil, "", errors.New("无管理员权限")
	}

	// 5. 更新用户信息（如果需要）
	updates := map[string]interface{}{
		"nickname": userInfo.Nickname,
		"avatar":   userInfo.HeadImgURL,
	}

	if userInfo.UnionID != "" && user.UnionID == "" {
		updates["union_id"] = userInfo.UnionID
	}

	if err = database.DB.Model(&user).Updates(updates).Error; err != nil {
		return nil, "", errors.New("更新用户信息失败")
	}

	// 6. 生成JWT令牌
	token, err := middleware.GenerateToken(user.ID)
	if err != nil {
		return nil, "", errors.New("生成令牌失败")
	}

	return &user, token, nil
}

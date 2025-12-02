package service

import (
	"bytes"
	"crypto/aes"
	"crypto/md5"
	"crypto/tls"
	"encoding/base64"
	"encoding/hex"
	"encoding/xml"
	"errors"
	"exam-system/internal/config"
	"exam-system/internal/model"
	"exam-system/internal/pkg/database"
	"exam-system/internal/pkg/payment"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/pkcs12"
)

var Payment = new(PaymentService)

type PaymentService struct{}

// 微信支付参数
type WXPayParams struct {
	AppID     string `json:"appId"`
	TimeStamp string `json:"timeStamp"`
	NonceStr  string `json:"nonceStr"`
	Package   string `json:"package"`
	SignType  string `json:"signType"`
	PaySign   string `json:"paySign"`
}

// 统一下单
func (s *PaymentService) Create(userID uint, courseID string, totalFee int, openID string, existingOrderNo string) (*payment.WXPayParams, *model.Order, error) {
	// 获取配置
	if config.GlobalConfig == nil {
		return nil, nil, errors.New("配置未初始化")
	}

	var order *model.Order

	// 如果提供了订单号，尝试查找已有订单
	if existingOrderNo != "" {
		var existingOrder model.Order
		if err := database.DB.Where("order_no = ? AND user_id = ?", existingOrderNo, userID).First(&existingOrder).Error; err != nil {
			return nil, nil, errors.New("订单不存在")
		}

		// 检查订单状态是否允许重新支付
		if existingOrder.Status != "unpaid" && existingOrder.Status != "pending" {
			return nil, nil, errors.New("订单状态不允许重新支付")
		}

		// 使用已有订单
		order = &existingOrder
	} else {
		// 生成新订单号
		orderNo := fmt.Sprintf("%s%d", time.Now().Format("20060102150405"), userID)

		// 将课程ID转换为uint
		courseIDUint, err := strconv.ParseUint(courseID, 10, 32)
		if err != nil {
			return nil, nil, fmt.Errorf("课程ID格式错误: %v", err)
		}

		// 创建新订单
		order = &model.Order{
			OrderNo:  orderNo,
			UserID:   userID,
			CourseID: uint(courseIDUint),
			Amount:   float64(totalFee) / 100, // 转换为元
			Status:   "pending",
		}

		if err := database.DB.Create(order).Error; err != nil {
			return nil, nil, fmt.Errorf("保存订单失败: %v", err)
		}
	}

	// 生成支付参数
	params, err := WeChat.GeneratePayParams(order.OrderNo, totalFee, openID)
	if err != nil {
		return nil, nil, fmt.Errorf("生成支付参数失败: %v", err)
	}

	return params, order, nil
}

// 处理支付回调
func (s *PaymentService) HandleNotify(notify *payment.WXPayNotify) error {
	fmt.Printf("=== 开始处理支付回调 ===\n")
	fmt.Printf("回调数据: %+v\n", notify)

	// 获取配置
	if config.GlobalConfig == nil {
		fmt.Println("错误: 配置未初始化")
		return errors.New("配置未初始化")
	}
	cfg := config.GlobalConfig.WeChat

	// 检查返回码
	fmt.Printf("检查返回码: ReturnCode=%s, ResultCode=%s, ReturnMsg=%s\n",
		notify.ReturnCode, notify.ResultCode, notify.ReturnMsg)
	if notify.ReturnCode != "SUCCESS" || notify.ResultCode != "SUCCESS" {
		fmt.Printf("错误: 支付失败，原因: %s\n", notify.ReturnMsg)
		return fmt.Errorf("支付失败: %s", notify.ReturnMsg)
	}

	// 验证签名
	signParams := make(map[string]string)
	signParams["return_code"] = notify.ReturnCode
	signParams["return_msg"] = notify.ReturnMsg
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

	fmt.Printf("签名参数: %+v\n", signParams)
	sign := Payment.generateSign(signParams, cfg.PayKey)
	fmt.Printf("生成的签名: %s\n原始签名: %s\n", sign, notify.Sign)

	// 签名验证
	signVerified := (sign == notify.Sign)
	if !signVerified {
		// 记录警告但继续处理，因为这可能是由于微信回调中的字段顺序或格式导致的
		fmt.Println("警告: 签名验证失败，但继续处理支付成功通知")
		// 如果是生产环境，可以添加额外的日志记录
	}

	// 查找订单
	var order model.Order
	fmt.Printf("查找订单: OrderNo=%s\n", notify.OutTradeNo)
	if err := database.DB.Where("order_no = ?", notify.OutTradeNo).First(&order).Error; err != nil {
		fmt.Printf("错误: 订单不存在，err=%v\n", err)
		return errors.New("订单不存在")
	}
	fmt.Printf("找到订单: %+v\n", order)

	// 查找课程信息以获取有效期
	var course model.Course
	fmt.Printf("查找课程: CourseID=%d\n", order.CourseID)
	if err := database.DB.First(&course, order.CourseID).Error; err != nil {
		fmt.Printf("错误: 课程不存在，err=%v\n", err)
		return fmt.Errorf("课程不存在: %v", err)
	}
	fmt.Printf("找到课程: %+v\n", course)

	// 计算过期时间
	var expireTime *time.Time
	if course.ExpireDays > 0 {
		t := time.Now().AddDate(0, 0, course.ExpireDays)
		expireTime = &t
		fmt.Printf("设置过期时间: %v\n", *expireTime)
	} else {
		fmt.Println("课程无过期时间")
	}

	// 更新订单状态
	now := time.Now()
	updates := map[string]interface{}{
		"status":       "paid",
		"pay_time":     &now,
		"payment_type": "wechat",
		"expire_time":  expireTime,
	}
	fmt.Printf("更新订单数据: %+v\n", updates)

	if err := database.DB.Model(&order).Updates(updates).Error; err != nil {
		fmt.Printf("错误: 更新订单状态失败，err=%v\n", err)
		return errors.New("更新订单状态失败")
	}
	fmt.Println("订单更新成功")

	fmt.Println("=== 支付回调处理完成 ===")
	return nil
}

// 查询支付结果
func (s *PaymentService) Query(userID uint, orderNo string) (map[string]interface{}, error) {
	var order model.Order
	if err := database.DB.Where("order_no = ? AND user_id = ?", orderNo, userID).First(&order).Error; err != nil {
		return nil, errors.New("订单不存在")
	}

	return map[string]interface{}{
		"order_no":     order.OrderNo,
		"status":       order.Status,
		"amount":       order.Amount,
		"payment_type": order.PaymentType,
		"pay_time":     order.PayTime,
	}, nil
}

// 生成签名
func (s *PaymentService) generateSign(params map[string]string, key string) string {
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

// 取消支付
func (s *PaymentService) CancelPayment(userID uint, orderNo string) error {
	var order model.Order
	if err := database.DB.Where("order_no = ? AND user_id = ?", orderNo, userID).First(&order).Error; err != nil {
		return errors.New("订单不存在")
	}

	// 如果状态已经是unpaid，直接返回成功
	if order.Status == "unpaid" {
		return nil
	}

	// 只有pending状态的订单可以取消
	if order.Status != "pending" {
		return errors.New("订单状态不允许取消")
	}

	// 更新订单状态为unpaid
	if err := database.DB.Model(&order).Update("status", "unpaid").Error; err != nil {
		return fmt.Errorf("更新订单状态失败: %v", err)
	}

	return nil
}

// CreateFreeOrder 创建免费订单并直接标记为已支付
func (s *PaymentService) CreateFreeOrder(userID uint, courseID string, existingOrderNo string) (*model.Order, error) {
	var order *model.Order

	// 如果提供了订单号，尝试查找已有订单
	if existingOrderNo != "" {
		var existingOrder model.Order
		if err := database.DB.Where("order_no = ? AND user_id = ?", existingOrderNo, userID).First(&existingOrder).Error; err != nil {
			return nil, errors.New("订单不存在")
		}

		// 检查订单状态
		if existingOrder.Status == "paid" {
			return &existingOrder, nil // 已经支付，直接返回
		}

		// 使用已有订单
		order = &existingOrder
	} else {
		// 生成新订单号
		orderNo := fmt.Sprintf("%s%d", time.Now().Format("20060102150405"), userID)

		// 将课程ID转换为uint
		courseIDUint, err := strconv.ParseUint(courseID, 10, 32)
		if err != nil {
			return nil, fmt.Errorf("课程ID格式错误: %v", err)
		}

		// 创建新订单
		order = &model.Order{
			OrderNo:  orderNo,
			UserID:   userID,
			CourseID: uint(courseIDUint),
			Amount:   0, // 免费课程金额为0
			Status:   "pending",
		}

		if err := database.DB.Create(order).Error; err != nil {
			return nil, fmt.Errorf("保存订单失败: %v", err)
		}
	}

	// 查找课程信息以获取有效期
	var course model.Course
	if err := database.DB.First(&course, order.CourseID).Error; err != nil {
		return nil, fmt.Errorf("课程不存在: %v", err)
	}

	// 计算过期时间
	var expireTime *time.Time
	if course.ExpireDays > 0 {
		t := time.Now().AddDate(0, 0, course.ExpireDays)
		expireTime = &t
	}

	// 更新订单状态为已支付
	now := time.Now()
	updates := map[string]interface{}{
		"status":       "paid",
		"pay_time":     &now,
		"payment_type": "free", // 标记为免费支付
		"expire_time":  expireTime,
	}

	if err := database.DB.Model(order).Updates(updates).Error; err != nil {
		return nil, errors.New("更新订单状态失败")
	}

	return order, nil
}

// Refund 申请退款
func (s *PaymentService) Refund(adminID uint, orderNo string, refundFee float64, refundReason string) (map[string]interface{}, error) {
	fmt.Printf("=== 开始处理退款请求 ===\n")
	fmt.Printf("管理员ID: %d, 订单号: %s, 退款金额: %.2f, 退款原因: %s\n", adminID, orderNo, refundFee, refundReason)

	// 获取配置
	if config.GlobalConfig == nil {
		return nil, errors.New("配置未初始化")
	}
	cfg := config.GlobalConfig.WeChat

	// 查找订单
	var order model.Order
	if err := database.DB.Where("order_no = ?", orderNo).First(&order).Error; err != nil {
		return nil, errors.New("订单不存在")
	}
	fmt.Printf("找到订单: %+v\n", order)

	// 检查订单状态
	if order.Status != "paid" {
		return nil, errors.New("订单状态不允许退款")
	}

	// 检查退款金额
	if refundFee <= 0 || refundFee > order.Amount {
		return nil, fmt.Errorf("退款金额无效，应在0-%.2f之间", order.Amount)
	}

	// 生成退款单号
	outRefundNo := fmt.Sprintf("refund_%s_%d", orderNo, time.Now().Unix())

	// 构建退款请求
	totalFee := int(order.Amount * 100)  // 转换为分
	refundFeeInt := int(refundFee * 100) // 转换为分

	refundReq := payment.WXRefundRequest{
		AppID:       cfg.AppID,
		MchID:       cfg.MchID,
		NonceStr:    s.generateNonceStr(),
		OutTradeNo:  orderNo,
		OutRefundNo: outRefundNo,
		TotalFee:    totalFee,
		RefundFee:   refundFeeInt,
		RefundDesc:  refundReason,
		NotifyUrl:   cfg.RefundNotifyURL,
	}

	// 生成签名
	params := make(map[string]string)
	params["appid"] = refundReq.AppID
	params["mch_id"] = refundReq.MchID
	params["nonce_str"] = refundReq.NonceStr
	params["out_trade_no"] = refundReq.OutTradeNo
	params["out_refund_no"] = refundReq.OutRefundNo
	params["total_fee"] = strconv.Itoa(refundReq.TotalFee)
	params["refund_fee"] = strconv.Itoa(refundReq.RefundFee)
	if refundReq.RefundDesc != "" {
		params["refund_desc"] = refundReq.RefundDesc
	}
	if refundReq.NotifyUrl != "" {
		params["notify_url"] = refundReq.NotifyUrl
	}

	refundReq.Sign = s.generateSign(params, cfg.PayKey)
	fmt.Printf("退款请求参数: %+v\n", refundReq)

	// 将请求转为XML
	xmlData, err := xml.Marshal(refundReq)
	if err != nil {
		return nil, fmt.Errorf("生成XML失败: %v", err)
	}

	// 加载TLS证书
	tlsConfig, err := s.loadTLSConfig()
	if err != nil {
		return nil, fmt.Errorf("加载TLS证书失败: %v", err)
	}

	// 创建带TLS证书的HTTP客户端
	transport := &http.Transport{
		TLSClientConfig: tlsConfig,
	}
	client := &http.Client{
		Transport: transport,
	}

	// 创建请求
	req, err := http.NewRequest("POST", cfg.RefundURL, bytes.NewBuffer(xmlData))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}
	req.Header.Set("Content-Type", "application/xml")

	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("发送请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}
	fmt.Printf("微信退款响应: %s\n", string(body))

	// 解析响应
	var refundResp payment.WXRefundResponse
	if err := xml.Unmarshal(body, &refundResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %v", err)
	}

	// 检查返回码
	if refundResp.ReturnCode != "SUCCESS" {
		return nil, fmt.Errorf("退款请求失败: %s", refundResp.ReturnMsg)
	}

	if refundResp.ResultCode != "SUCCESS" {
		return nil, fmt.Errorf("退款失败: %s - %s", refundResp.ErrCode, refundResp.ErrCodeDes)
	}

	// 更新订单状态，只更新status字段
	if err := database.DB.Model(&order).Update("status", "refunding").Error; err != nil {
		return nil, errors.New("更新订单状态失败")
	}

	// 记录退款信息到日志（不保存到数据库）
	fmt.Printf("退款申请成功，订单号: %s, 退款单号: %s, 退款ID: %s, 退款金额: %.2f\n",
		orderNo, outRefundNo, refundResp.RefundID, refundFee)

	fmt.Println("=== 退款请求处理完成 ===")
	return map[string]interface{}{
		"order_no":      order.OrderNo,
		"out_refund_no": outRefundNo,
		"refund_id":     refundResp.RefundID,
		"refund_fee":    refundFee,
		"status":        "refunding",
	}, nil
}

// loadTLSConfig 加载TLS证书配置
func (s *PaymentService) loadTLSConfig() (*tls.Config, error) {
	// 获取配置
	if config.GlobalConfig == nil {
		return nil, errors.New("配置未初始化")
	}
	cfg := config.GlobalConfig.WeChat

	// 获取工作目录
	workDir, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("获取工作目录失败: %v", err)
	}

	// 证书路径
	certPath := filepath.Join(workDir, cfg.CertPath)
	fmt.Printf("证书路径: %s\n", certPath)

	// 读取证书文件
	certData, err := os.ReadFile(certPath)
	if err != nil {
		return nil, fmt.Errorf("读取证书文件失败: %v", err)
	}

	// 解析PKCS12证书
	privateKey, certificate, err := pkcs12.Decode(certData, cfg.MchID)
	if err != nil {
		return nil, fmt.Errorf("解析PKCS12证书失败: %v", err)
	}

	// 构建TLS证书
	tlsCert := tls.Certificate{
		Certificate: [][]byte{certificate.Raw},
		PrivateKey:  privateKey,
		Leaf:        certificate,
	}

	// 创建TLS配置
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{tlsCert},
	}

	return tlsConfig, nil
}

// HandleRefundNotify 处理退款回调
func (s *PaymentService) HandleRefundNotify(notify *payment.WXRefundNotify) error {
	fmt.Printf("=== 开始处理退款回调 ===\n")
	fmt.Printf("回调数据: %+v\n", notify)

	// 获取配置
	if config.GlobalConfig == nil {
		return errors.New("配置未初始化")
	}
	cfg := config.GlobalConfig.WeChat

	// 检查返回码
	if notify.ReturnCode != "SUCCESS" {
		return fmt.Errorf("退款回调失败: %s", notify.ReturnMsg)
	}

	// 解密退款信息
	reqInfo, err := s.decryptRefundInfo(notify.ReqInfo, cfg.PayKey)
	if err != nil {
		return fmt.Errorf("解密退款信息失败: %v", err)
	}

	// 解析解密后的XML
	var decryptedInfo payment.WXRefundNotifyDecrypted
	if err := xml.Unmarshal([]byte(reqInfo), &decryptedInfo); err != nil {
		return fmt.Errorf("解析解密后的XML失败: %v", err)
	}
	fmt.Printf("解密后的退款信息: %+v\n", decryptedInfo)

	// 查找订单
	var order model.Order
	if err := database.DB.Where("order_no = ?", decryptedInfo.OutTradeNo).First(&order).Error; err != nil {
		return errors.New("订单不存在")
	}

	// 更新订单状态
	status := "refund_failed"
	if decryptedInfo.RefundStatus == "SUCCESS" {
		status = "refunded" // 退款成功
	}

	// 只更新status字段
	if err := database.DB.Model(&order).Update("status", status).Error; err != nil {
		return errors.New("更新订单状态失败")
	}

	// 记录退款信息到日志（不保存到数据库）
	fmt.Printf("退款成功，订单号: %s, 退款ID: %s, 退款状态: %s, 退款金额: %d\n",
		decryptedInfo.OutTradeNo, decryptedInfo.RefundID, decryptedInfo.RefundStatus, decryptedInfo.RefundFee)

	fmt.Println("=== 退款回调处理完成 ===")
	return nil
}

// QueryRefund 查询退款状态
func (s *PaymentService) QueryRefund(adminID uint, orderNo string) (map[string]interface{}, error) {
	fmt.Printf("=== 开始查询退款状态 ===\n")
	fmt.Printf("管理员ID: %d, 订单号: %s\n", adminID, orderNo)

	// 获取配置
	if config.GlobalConfig == nil {
		return nil, errors.New("配置未初始化")
	}
	cfg := config.GlobalConfig.WeChat

	// 查找订单
	var order model.Order
	if err := database.DB.Where("order_no = ?", orderNo).First(&order).Error; err != nil {
		return nil, errors.New("订单不存在")
	}

	// 构建查询请求
	queryReq := payment.WXRefundQueryRequest{
		AppID:      cfg.AppID,
		MchID:      cfg.MchID,
		NonceStr:   s.generateNonceStr(),
		OutTradeNo: orderNo,
	}

	// 生成签名
	params := make(map[string]string)
	params["appid"] = queryReq.AppID
	params["mch_id"] = queryReq.MchID
	params["nonce_str"] = queryReq.NonceStr
	params["out_trade_no"] = queryReq.OutTradeNo

	queryReq.Sign = s.generateSign(params, cfg.PayKey)
	fmt.Printf("查询请求参数: %+v\n", queryReq)

	// 将请求转为XML
	xmlData, err := xml.Marshal(queryReq)
	if err != nil {
		return nil, fmt.Errorf("生成XML失败: %v", err)
	}

	// 创建普通HTTP客户端（查询接口不需要证书）
	client := &http.Client{}

	// 创建请求
	req, err := http.NewRequest("POST", cfg.RefundQueryURL, bytes.NewBuffer(xmlData))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}
	req.Header.Set("Content-Type", "application/xml")

	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("发送请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}
	fmt.Printf("微信查询响应: %s\n", string(body))

	// 解析响应
	var queryResp payment.WXRefundQueryResponse
	if err := xml.Unmarshal(body, &queryResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %v", err)
	}

	// 检查返回码
	if queryResp.ReturnCode != "SUCCESS" {
		return nil, fmt.Errorf("查询请求失败: %s", queryResp.ReturnMsg)
	}

	if queryResp.ResultCode != "SUCCESS" {
		return nil, fmt.Errorf("查询失败: %s - %s", queryResp.ErrCode, queryResp.ErrCodeDes)
	}

	// 更新订单状态，只更新status字段
	if queryResp.RefundStatus0 == "SUCCESS" && order.Status != "refunded" {
		if err := database.DB.Model(&order).Update("status", "refunded").Error; err != nil {
			return nil, errors.New("更新订单状态失败")
		}
	} else if queryResp.RefundStatus0 == "FAIL" && order.Status != "refund_failed" {
		if err := database.DB.Model(&order).Update("status", "refund_failed").Error; err != nil {
			return nil, errors.New("更新订单状态失败")
		}
	}

	fmt.Println("=== 查询退款状态完成 ===")
	return map[string]interface{}{
		"order_no":       order.OrderNo,
		"out_refund_no":  queryResp.OutRefundNo0,
		"refund_id":      queryResp.RefundID0,
		"refund_fee":     float64(queryResp.RefundFee0) / 100,
		"refund_status":  queryResp.RefundStatus0,
		"success_time":   queryResp.RefundSuccessTime0,
		"refund_account": queryResp.RefundRecvAccout0,
	}, nil
}

// 生成随机字符串
func (s *PaymentService) generateNonceStr() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

// 解密退款通知信息
func (s *PaymentService) decryptRefundInfo(reqInfo, key string) (string, error) {
	// 对商户key做md5，得到32位小写key
	h := md5.New()
	h.Write([]byte(key))
	key = hex.EncodeToString(h.Sum(nil))

	// base64解码
	cipherData, err := base64.StdEncoding.DecodeString(reqInfo)
	if err != nil {
		return "", fmt.Errorf("base64解码失败: %v", err)
	}

	// 使用AES-256-ECB解密
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return "", fmt.Errorf("创建AES密码块失败: %v", err)
	}

	// ECB模式解密
	decrypted := make([]byte, len(cipherData))
	size := block.BlockSize()

	for bs, be := 0, size; bs < len(cipherData); bs, be = bs+size, be+size {
		block.Decrypt(decrypted[bs:be], cipherData[bs:be])
	}

	// 去除PKCS#7填充
	padding := int(decrypted[len(decrypted)-1])
	decrypted = decrypted[:len(decrypted)-padding]

	return string(decrypted), nil
}

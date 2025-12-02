package api

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"exam-system/internal/pkg/payment"
	"exam-system/internal/service"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// 统一下单请求参数
type CreatePaymentRequest struct {
	CourseID string `json:"course_id" binding:"required"`
	TotalFee int    `json:"total_fee" binding:"required"`
	OpenID   string `json:"open_id" binding:"required"`
}

// 微信支付回调通知参数
type WXPayNotifyRequest struct {
	XMLName       xml.Name `xml:"xml"`
	AppID         string   `xml:"appid"`
	MchID         string   `xml:"mch_id"`
	NonceStr      string   `xml:"nonce_str"`
	Sign          string   `xml:"sign"`
	ResultCode    string   `xml:"result_code"`
	OpenID        string   `xml:"openid"`
	OutTradeNo    string   `xml:"out_trade_no"`
	TransactionID string   `xml:"transaction_id"`
	TotalFee      int      `xml:"total_fee"`
	TimeEnd       string   `xml:"time_end"`
}

// RefundRequest 退款请求参数
type RefundRequest struct {
	OrderNo      string  `json:"order_no" binding:"required"`   // 订单号
	RefundFee    float64 `json:"refund_fee" binding:"required"` // 退款金额（元）
	RefundReason string  `json:"refund_reason"`                 // 退款原因
}

// 统一下单
func CreatePayment(c *gin.Context) {
	var req struct {
		CourseID string `json:"course_id" binding:"required"`
		TotalFee int    `json:"total_fee" binding:"required"`
		OpenID   string `json:"open_id" binding:"required"`
		OrderNo  string `json:"order_no,omitempty"` // 可选参数，用于重新发起支付
	}

	// 打印原始请求体
	rawData, _ := ioutil.ReadAll(c.Request.Body)
	c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(rawData))
	fmt.Printf("CreatePayment 原始请求数据: %s\n", string(rawData))

	if err := c.ShouldBindJSON(&req); err != nil {
		fmt.Printf("CreatePayment 参数绑定错误: %v\n", err)

		// 尝试手动解析JSON，以便更灵活地处理
		var manualReq map[string]interface{}
		if jsonErr := json.Unmarshal(rawData, &manualReq); jsonErr == nil {
			fmt.Printf("手动解析JSON成功: %+v\n", manualReq)

			// 检查是否包含所需字段
			courseID, hasCourseID := manualReq["course_id"]
			totalFee, hasTotalFee := manualReq["total_fee"]
			openID, hasOpenID := manualReq["open_id"]

			if hasCourseID && hasTotalFee && hasOpenID {
				// 手动构建请求
				courseIDStr, ok1 := courseID.(string)
				totalFeeFloat, ok2 := totalFee.(float64) // JSON中的数字会被解析为float64
				openIDStr, ok3 := openID.(string)

				if ok1 && ok2 && ok3 {
					totalFeeInt := int(totalFeeFloat)
					fmt.Printf("手动构建请求: CourseID=%s, TotalFee=%d, OpenID=%s\n",
						courseIDStr, totalFeeInt, openIDStr)

					// 继续处理请求
					userID := c.GetUint("userId")

					// 处理金额为0的情况（免费课程）
					if totalFeeInt == 0 {
						orderNo := ""
						if orderNoVal, hasOrderNo := manualReq["order_no"]; hasOrderNo {
							if orderNoStr, ok := orderNoVal.(string); ok {
								orderNo = orderNoStr
							}
						}

						fmt.Println("手动处理: 检测到免费课程，使用 CreateFreeOrder")
						order, err := service.Payment.CreateFreeOrder(userID, courseIDStr, orderNo)
						if err != nil {
							c.JSON(http.StatusInternalServerError, gin.H{
								"code": 500,
								"msg":  err.Error(),
							})
							return
						}

						c.JSON(http.StatusOK, gin.H{
							"code": 200,
							"data": gin.H{
								"orderNo": order.OrderNo,
								"status":  "paid",
								"message": "免费课程已开通",
							},
						})
						return
					}

					// 正常支付流程
					orderNo := ""
					if orderNoVal, hasOrderNo := manualReq["order_no"]; hasOrderNo {
						if orderNoStr, ok := orderNoVal.(string); ok {
							orderNo = orderNoStr
						}
					}

					fmt.Println("手动处理: 执行正常支付流程")
					params, order, err := service.Payment.Create(userID, courseIDStr, totalFeeInt, openIDStr, orderNo)
					if err != nil {
						c.JSON(http.StatusInternalServerError, gin.H{
							"code": 500,
							"msg":  err.Error(),
						})
						return
					}

					c.JSON(http.StatusOK, gin.H{
						"code": 200,
						"data": gin.H{
							"orderNo": order.OrderNo,
							"params":  params,
						},
					})
					return
				}
			}
		}

		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  fmt.Sprintf("参数错误: %v", err),
		})
		return
	}

	fmt.Printf("CreatePayment 请求参数: CourseID=%s, TotalFee=%d, OpenID=%s, OrderNo=%s\n",
		req.CourseID, req.TotalFee, req.OpenID, req.OrderNo)

	userID := c.GetUint("userId")
	fmt.Printf("CreatePayment 用户ID: %d\n", userID)

	// 处理金额为0的情况（免费课程）
	if req.TotalFee == 0 {
		fmt.Println("CreatePayment 检测到免费课程，使用 CreateFreeOrder")
		// 直接创建订单并标记为已支付
		order, err := service.Payment.CreateFreeOrder(userID, req.CourseID, req.OrderNo)
		if err != nil {
			fmt.Printf("CreatePayment 创建免费订单失败: %v\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"code": 500,
				"msg":  err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"code": 200,
			"data": gin.H{
				"orderNo": order.OrderNo,
				"status":  "paid", // 直接标记为已支付
				"message": "免费课程已开通",
			},
		})
		return
	}

	// 正常支付流程
	fmt.Println("CreatePayment 执行正常支付流程")
	params, order, err := service.Payment.Create(userID, req.CourseID, req.TotalFee, req.OpenID, req.OrderNo)
	if err != nil {
		fmt.Printf("CreatePayment 创建支付订单失败: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  err.Error(),
		})
		return
	}

	fmt.Printf("CreatePayment 创建支付订单成功: OrderNo=%s\n", order.OrderNo)
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": gin.H{
			"orderNo": order.OrderNo,
			"params":  params,
		},
	})
}

// 支付回调通知
func PaymentNotify(c *gin.Context) {
	fmt.Println("=== 收到支付回调请求 ===")
	contentType := c.GetHeader("Content-Type")
	fmt.Printf("Content-Type: %s\n", contentType)

	// 读取原始请求数据
	body, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		fmt.Printf("读取请求数据失败: %v\n", err)
		c.JSON(http.StatusOK, gin.H{
			"code": 500,
			"msg":  "读取请求数据失败",
		})
		return
	}
	fmt.Printf("原始请求数据: %s\n", string(body))

	// 尝试解析为JSON格式
	var jsonData struct {
		OrderNo string `json:"order_no"`
		Status  string `json:"status"`
	}
	if err := json.Unmarshal(body, &jsonData); err == nil && jsonData.OrderNo != "" {
		fmt.Println("=== 使用JSON格式处理回调 ===")
		fmt.Printf("解析JSON数据成功: %+v\n", jsonData)

		// 如果是JSON格式，转换为WXPayNotify格式
		notifyReq := &payment.WXPayNotify{
			ReturnCode: "SUCCESS",
			ResultCode: "SUCCESS",
			OutTradeNo: jsonData.OrderNo,
			TimeEnd:    time.Now().Format("20060102150405"),
		}

		// 处理支付回调
		if err := service.Payment.HandleNotify(notifyReq); err != nil {
			fmt.Printf("处理支付回调失败: %v\n", err)
			c.JSON(http.StatusOK, gin.H{
				"code": 500,
				"msg":  err.Error(),
			})
			return
		}

		fmt.Println("支付回调处理成功")
		c.JSON(http.StatusOK, gin.H{
			"code": 200,
			"msg":  "success",
		})
		return
	}

	// 如果不是JSON格式，尝试解析为XML格式
	fmt.Println("=== 尝试使用XML格式处理回调 ===")
	var notifyReq payment.WXPayNotify
	if err := xml.Unmarshal(body, &notifyReq); err != nil {
		fmt.Printf("解析XML数据失败（这是正常的，如果已经以JSON格式处理）: %v\n", err)
		c.XML(http.StatusOK, gin.H{
			"return_code": "FAIL",
			"return_msg":  "参数格式错误",
		})
		return
	}
	fmt.Printf("解析后的通知数据: %+v\n", notifyReq)

	// 处理支付回调
	if err := service.Payment.HandleNotify(&notifyReq); err != nil {
		fmt.Printf("处理支付回调失败: %v\n", err)
		c.XML(http.StatusOK, gin.H{
			"return_code": "FAIL",
			"return_msg":  err.Error(),
		})
		return
	}

	fmt.Println("支付回调处理成功")
	c.XML(http.StatusOK, gin.H{
		"return_code": "SUCCESS",
		"return_msg":  "OK",
	})
}

// 查询支付结果
func QueryPayment(c *gin.Context) {
	orderNo := c.Param("order_no")
	if orderNo == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "订单号不能为空",
		})
		return
	}

	// 获取当前用户ID
	userID := c.GetUint("userId")

	// 查询支付结果
	result, err := service.Payment.Query(userID, orderNo)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  err.Error(),
		})
		return
	}

	fmt.Println("查询支付结果", result)
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": result,
	})
}

// 取消支付
func CancelPayment(c *gin.Context) {
	orderNo := c.Param("order_no")
	if orderNo == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "订单号不能为空",
		})
		return
	}

	// 获取当前用户ID
	userID := c.GetUint("userId")

	// 取消支付
	if err := service.Payment.CancelPayment(userID, orderNo); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "取消支付成功",
	})
}

// RefundPayment 申请退款
func RefundPayment(c *gin.Context) {
	// 只有管理员可以操作退款
	adminID, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code": 401,
			"msg":  "未登录或登录已过期",
		})
		return
	}

	// 解析请求参数
	var req RefundRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fmt.Printf("RefundPayment 参数绑定错误: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  fmt.Sprintf("参数错误: %v", err),
		})
		return
	}

	fmt.Printf("RefundPayment 请求参数: OrderNo=%s, RefundFee=%.2f, RefundReason=%s\n",
		req.OrderNo, req.RefundFee, req.RefundReason)

	// 验证退款金额
	if req.RefundFee <= 0 {
		fmt.Printf("RefundPayment 退款金额无效: %.2f\n", req.RefundFee)
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "退款金额必须大于0",
		})
		return
	}

	// 调用退款服务
	refundResult, err := service.Payment.Refund(adminID.(uint), req.OrderNo, req.RefundFee, req.RefundReason)
	if err != nil {
		fmt.Printf("RefundPayment 退款失败: %v\n", err)

		// 根据错误类型返回不同的状态码
		if strings.Contains(err.Error(), "订单不存在") ||
			strings.Contains(err.Error(), "订单状态不允许退款") {
			c.JSON(http.StatusBadRequest, gin.H{
				"code": 400,
				"msg":  err.Error(),
			})
		} else if strings.Contains(err.Error(), "加载TLS证书失败") {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":  500,
				"msg":   "系统配置错误，请联系管理员",
				"error": err.Error(),
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code": 500,
				"msg":  err.Error(),
			})
		}
		return
	}

	fmt.Printf("RefundPayment 退款成功: %+v\n", refundResult)
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": refundResult,
		"msg":  "退款申请成功",
	})
}

// RefundNotify 退款回调通知
func RefundNotify(c *gin.Context) {
	fmt.Println("=== 收到退款回调请求 ===")
	contentType := c.GetHeader("Content-Type")
	fmt.Printf("Content-Type: %s\n", contentType)

	// 读取原始请求数据
	body, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		fmt.Printf("读取请求数据失败: %v\n", err)
		c.XML(http.StatusOK, gin.H{
			"return_code": "FAIL",
			"return_msg":  "读取请求数据失败",
		})
		return
	}
	fmt.Printf("原始请求数据: %s\n", string(body))

	// 解析XML数据
	var notifyReq payment.WXRefundNotify
	if err := xml.Unmarshal(body, &notifyReq); err != nil {
		fmt.Printf("解析XML数据失败: %v\n", err)
		c.XML(http.StatusOK, gin.H{
			"return_code": "FAIL",
			"return_msg":  "参数格式错误",
		})
		return
	}
	fmt.Printf("解析后的通知数据: %+v\n", notifyReq)

	// 检查必要字段
	if notifyReq.ReturnCode == "" {
		fmt.Println("退款回调缺少必要字段: return_code")
		c.XML(http.StatusOK, gin.H{
			"return_code": "FAIL",
			"return_msg":  "缺少必要字段",
		})
		return
	}

	// 处理退款回调
	if err := service.Payment.HandleRefundNotify(&notifyReq); err != nil {
		fmt.Printf("处理退款回调失败: %v\n", err)
		c.XML(http.StatusOK, gin.H{
			"return_code": "FAIL",
			"return_msg":  err.Error(),
		})
		return
	}

	fmt.Println("退款回调处理成功")
	c.XML(http.StatusOK, gin.H{
		"return_code": "SUCCESS",
		"return_msg":  "OK",
	})
}

// QueryRefund 查询退款状态
func QueryRefund(c *gin.Context) {
	orderNo := c.Param("order_no")
	if orderNo == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "订单号不能为空",
		})
		return
	}

	// 获取当前用户ID
	adminID := c.GetUint("userId")
	fmt.Printf("QueryRefund 请求参数: adminID=%d, orderNo=%s\n", adminID, orderNo)

	// 查询退款结果
	result, err := service.Payment.QueryRefund(adminID, orderNo)
	if err != nil {
		fmt.Printf("QueryRefund 查询失败: %v\n", err)

		// 根据错误类型返回不同的状态码
		if strings.Contains(err.Error(), "订单不存在") {
			c.JSON(http.StatusBadRequest, gin.H{
				"code": 400,
				"msg":  err.Error(),
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code": 500,
				"msg":  err.Error(),
			})
		}
		return
	}

	fmt.Printf("QueryRefund 查询结果: %+v\n", result)
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": result,
	})
}

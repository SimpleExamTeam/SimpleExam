package payment

// WXPayParams 微信支付参数结构体
type WXPayParams struct {
	AppID     string `json:"appId"`
	TimeStamp string `json:"timeStamp"`
	NonceStr  string `json:"nonceStr"`
	Package   string `json:"package"`
	SignType  string `json:"signType"`
	PaySign   string `json:"paySign"`
}

// WXPayNotify 微信支付回调结构体
type WXPayNotify struct {
	ReturnCode    string `xml:"return_code"`
	ReturnMsg     string `xml:"return_msg"`
	ResultCode    string `xml:"result_code"`
	AppID         string `xml:"appid"`
	MchID         string `xml:"mch_id"`
	DeviceInfo    string `xml:"device_info"`
	NonceStr      string `xml:"nonce_str"`
	Sign          string `xml:"sign"`
	TransactionID string `xml:"transaction_id"`
	OutTradeNo    string `xml:"out_trade_no"`
	TotalFee      int    `xml:"total_fee"`
	TimeEnd       string `xml:"time_end"`
	BankType      string `xml:"bank_type"`
	CashFee       string `xml:"cash_fee"`
	FeeType       string `xml:"fee_type"`
	IsSubscribe   string `xml:"is_subscribe"`
	TradeType     string `xml:"trade_type"`
	OpenID        string `xml:"openid"`
	// 优惠券相关字段
	CouponCount string `xml:"coupon_count"`
	CouponFee   string `xml:"coupon_fee"`
	CouponFee_0 string `xml:"coupon_fee_0"`
	CouponID_0  string `xml:"coupon_id_0"`
}

// WXRefundRequest 微信退款请求参数
type WXRefundRequest struct {
	AppID         string `xml:"appid"`
	MchID         string `xml:"mch_id"`
	NonceStr      string `xml:"nonce_str"`
	Sign          string `xml:"sign"`
	SignType      string `xml:"sign_type,omitempty"`
	TransactionID string `xml:"transaction_id,omitempty"`
	OutTradeNo    string `xml:"out_trade_no"`
	OutRefundNo   string `xml:"out_refund_no"`
	TotalFee      int    `xml:"total_fee"`
	RefundFee     int    `xml:"refund_fee"`
	RefundDesc    string `xml:"refund_desc,omitempty"`
	NotifyUrl     string `xml:"notify_url,omitempty"`
}

// WXRefundResponse 微信退款响应参数
type WXRefundResponse struct {
	ReturnCode    string `xml:"return_code"`
	ReturnMsg     string `xml:"return_msg"`
	ResultCode    string `xml:"result_code"`
	ErrCode       string `xml:"err_code"`
	ErrCodeDes    string `xml:"err_code_des"`
	AppID         string `xml:"appid"`
	MchID         string `xml:"mch_id"`
	NonceStr      string `xml:"nonce_str"`
	Sign          string `xml:"sign"`
	TransactionID string `xml:"transaction_id"`
	OutTradeNo    string `xml:"out_trade_no"`
	OutRefundNo   string `xml:"out_refund_no"`
	RefundID      string `xml:"refund_id"`
	RefundFee     int    `xml:"refund_fee"`
	TotalFee      int    `xml:"total_fee"`
	CashFee       int    `xml:"cash_fee"`
}

// WXRefundNotify 微信退款回调通知参数
type WXRefundNotify struct {
	ReturnCode string `xml:"return_code"`
	ReturnMsg  string `xml:"return_msg"`
	AppID      string `xml:"appid"`
	MchID      string `xml:"mch_id"`
	NonceStr   string `xml:"nonce_str"`
	ReqInfo    string `xml:"req_info"` // 加密信息
}

// WXRefundNotifyDecrypted 解密后的微信退款回调通知参数
type WXRefundNotifyDecrypted struct {
	TransactionID       string `xml:"transaction_id"`
	OutTradeNo          string `xml:"out_trade_no"`
	RefundID            string `xml:"refund_id"`
	OutRefundNo         string `xml:"out_refund_no"`
	TotalFee            int    `xml:"total_fee"`
	RefundFee           int    `xml:"refund_fee"`
	RefundStatus        string `xml:"refund_status"`
	SuccessTime         string `xml:"success_time"`
	RefundRecvAccout    string `xml:"refund_recv_accout"`
	RefundAccount       string `xml:"refund_account"`
	RefundRequestSource string `xml:"refund_request_source"`
}

// WXRefundQueryRequest 微信退款查询请求参数
type WXRefundQueryRequest struct {
	AppID         string `xml:"appid"`
	MchID         string `xml:"mch_id"`
	NonceStr      string `xml:"nonce_str"`
	Sign          string `xml:"sign"`
	SignType      string `xml:"sign_type,omitempty"`
	TransactionID string `xml:"transaction_id,omitempty"`
	OutTradeNo    string `xml:"out_trade_no,omitempty"`
	OutRefundNo   string `xml:"out_refund_no,omitempty"`
	RefundID      string `xml:"refund_id,omitempty"`
}

// WXRefundQueryResponse 微信退款查询响应参数
type WXRefundQueryResponse struct {
	ReturnCode         string `xml:"return_code"`
	ReturnMsg          string `xml:"return_msg"`
	ResultCode         string `xml:"result_code"`
	ErrCode            string `xml:"err_code"`
	ErrCodeDes         string `xml:"err_code_des"`
	AppID              string `xml:"appid"`
	MchID              string `xml:"mch_id"`
	NonceStr           string `xml:"nonce_str"`
	Sign               string `xml:"sign"`
	TotalRefundCount   int    `xml:"total_refund_count"`
	TransactionID      string `xml:"transaction_id"`
	OutTradeNo         string `xml:"out_trade_no"`
	TotalFee           int    `xml:"total_fee"`
	RefundCount        int    `xml:"refund_count"`
	OutRefundNo0       string `xml:"out_refund_no_0"`
	RefundID0          string `xml:"refund_id_0"`
	RefundFee0         int    `xml:"refund_fee_0"`
	RefundStatus0      string `xml:"refund_status_0"`
	RefundRecvAccout0  string `xml:"refund_recv_accout_0"`
	RefundSuccessTime0 string `xml:"refund_success_time_0"`
}

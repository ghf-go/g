package g

import (
	"time"
)

const (
	PAYMENT_WX      = "wx"
	PAYMENT_ALI     = "ali"
	PAYMENT_YINLIAN = "yinlian"
)

// 创建订单参数
type OrderParam struct {
	PaymentChannel string //支付通道
	PaymentType    string //支付方式，h5,app,jsapi,native
	OrderID        string //订单ID
	OrderTitle     string //订单标题
	OrderDesc      string //订单描述
	OrderAttach    string //订单扩展信息
	OrderAmount    uint64 //订单价格，单位分
}

// 下单后返回数据
type OrderOut struct {
	PaymentChannel string //支付通道
	PaymentType    string //支付方式，h5,app,jsapi,native
	OrderID        string //订单ID
	ResultStr      string //返回的字符串信息

}

// 退款参数
type RefundParam struct {
	PaymentChannel string //支付通道
	OrderID        string //订单ID
	RefundOrderID  string //退款订单号
	RefundAmount   int64  //退款金额
	Msg            string //退款原因
}
type RefundOut struct {
	RefundID string //退款订单号
}

// 支付相关
type Payment interface {
	CreateOrder(OrderParam) *OrderOut                     //创建订单
	CloseOrder(OrderParam)                                //关闭订单
	Refund(RefundParam) *RefundOut                        //退款
	NotifyUrl(func(*NotifyData) error) GHandlerFunc       //支付回掉地址
	RefundNotifyUrl(func(*NotifyData) error) GHandlerFunc //退款回掉地址
}
type NotifyData struct {
	PaymentChannel string //支付通道
	OrderID        string //订单ID
	RefundOrderID  string //退款订单号
	IsPayment      bool   //是否是支付
	IsRefund       bool   //是否是退款
	IsSuccess      bool   //是否成功
	PlaformUID     string //第三方平台账号ID
	TransactionID  string //第三方订单ID
	Amount         int64  //金额 单位分
	Ext            string //通知信息
}

// 微信回掉通知结构
type WxNotifyBody struct {
	Mchid               string             `json:"mchid"`
	TransactionID       string             `json:"transaction_id"`
	OutTradeNo          string             `json:"out_trade_no"`
	RefundID            string             `json:"refund_id"`
	OutRefundNo         string             `json:"out_refund_no"`
	RefundStatus        string             `json:"refund_status"`
	TradeState          string             `json:"trade_state"`
	SuccessTime         time.Time          `json:"success_time"`
	UserReceivedAccount string             `json:"user_received_account"`
	Amount              WxNotifyBodyAmount `json:"amount"`
	Payer               WxNotifyBodyPayer  `json:"payer"`
	AppID               string             `json:"AppID"`
	TradeStateDesc      string             `json:"trade_state_desc"`
	TradeType           string             `json:"trade_type"`
	Attach              string             `json:"attach"`
}
type WxNotifyBodyAmount struct {
	Total       int64 `json:"total"`
	Refund      int64 `json:"refund"`
	PayerTotal  int64 `json:"payer_total"`
	PayerRefund int64 `json:"payer_refund"`
}
type WxNotifyBodyPayer struct {
	Openid string `json:"openid"`
}

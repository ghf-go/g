package g

// 创建订单参数
type OrderParam struct{}

// 退款参数
type RefundParam struct{}

// 支付相关
type Payment interface {
	CreateOrder(OrderParam)                 //创建订单
	Refund(RefundParam)                     //退款
	NotifyUrl(*GContext) GHandlerFunc       //支付回掉地址
	RefundNotifyUrl(*GContext) GHandlerFunc //退款回掉地址
}

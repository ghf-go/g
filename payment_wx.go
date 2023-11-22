package g

type confPaymentWx struct{}

// https://github.com/wechatpay-apiv3/wechatpay-go
// 微信支付回掉

// 创建订单
func (p *confPaymentWx) CreateOrder(OrderParam) {}

// 退款
func (p *confPaymentWx) Refund(RefundParam) {}

// 支付回掉地址
func (p *confPaymentWx) NotifyUrl(*GContext) GHandlerFunc {
	return nil
}

// 退款回掉地址
func (p *confPaymentWx) RefundNotifyUrl(*GContext) GHandlerFunc {
	return nil
}

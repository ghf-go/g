package g

type confPaymentZli struct{}

// 创建订单
func (p *confPaymentZli) CreateOrder(OrderParam) {}

// 退款
func (p *confPaymentZli) Refund(RefundParam) {}

// 支付回掉地址
func (p *confPaymentZli) NotifyUrl(*GContext) GHandlerFunc {
	return nil
}

// 退款回掉地址
func (p *confPaymentZli) RefundNotifyUrl(*GContext) GHandlerFunc {
	return nil
}

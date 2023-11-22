package g

type confPaymentYinLian struct{}

// https://open.sandpay.com.cn/product/detail/43310/43781/

// 创建订单
func (p *confPaymentYinLian) CreateOrder(OrderParam) {}

// 退款
func (p *confPaymentYinLian) Refund(RefundParam) {}

// 支付回掉地址
func (p *confPaymentYinLian) NotifyUrl(*GContext) GHandlerFunc {
	return nil
}

// 退款回掉地址
func (p *confPaymentYinLian) RefundNotifyUrl(*GContext) GHandlerFunc {
	return nil
}

package g

import (
	"context"
	"encoding/json"
	"time"

	"github.com/wechatpay-apiv3/wechatpay-go/core"
	"github.com/wechatpay-apiv3/wechatpay-go/core/auth/verifiers"
	"github.com/wechatpay-apiv3/wechatpay-go/core/downloader"
	"github.com/wechatpay-apiv3/wechatpay-go/core/notify"
	"github.com/wechatpay-apiv3/wechatpay-go/core/option"
	"github.com/wechatpay-apiv3/wechatpay-go/services/partnerpayments/jsapi"
	"github.com/wechatpay-apiv3/wechatpay-go/services/partnerpayments/native"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments/app"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments/h5"
	"github.com/wechatpay-apiv3/wechatpay-go/services/refunddomestic"
	"github.com/wechatpay-apiv3/wechatpay-go/utils"
)

type confPaymentWx struct {
	AppID           string `yaml:"app_id"`
	MchID           string `yaml:"mch_id"`
	MchIdNumber     string `yaml:"mch_id_num"`
	MchApiV3Key     string `yaml:"mch_api_v3_key"`
	PriviteKeyPem   string `yaml:"private_key_pem"`
	NotifyURL       string `yaml:"notify_url"`
	RefundNotifyURL string `yaml:"refund_notify_url"`
	client          *core.Client
	handle          *notify.Handler
}

// https://github.com/wechatpay-apiv3/wechatpay-go
// 微信支付回掉

// 创建订单
func (p *confPaymentWx) CreateOrder(r OrderParam) *OrderOut {
	switch r.PaymentType {
	case "h5":
		return p.createOrderH5(r)
	case "app":
		return p.createOrderApp(r)
	case "jsapi":
		return p.createOrderJsApi(r)
	case "native":
		return p.createOrderNative(r)
	}
	return nil
}

// 关闭订单
func (p *confPaymentWx) CloseOrder(r OrderParam) {
	switch r.PaymentType {
	case "h5":
		p.closeOrderH5(r)
	case "app":
		p.closeOrderApp(r)
	case "jsapi":
		p.closeOrderJsApi(r)
	case "native":
		p.closeOrderNative(r)
	}

}

// https://pay.weixin.qq.com/docs/merchant/apis/refund/refunds/create.html
// 退款
func (p *confPaymentWx) Refund(r RefundParam) *RefundOut {
	svc := refunddomestic.RefundsApiService{Client: p.getClient()}
	resp, _, err := svc.Create(context.Background(),
		refunddomestic.CreateRequest{
			SubMchid:    core.String(p.MchID),
			OutTradeNo:  core.String(r.OrderID),
			OutRefundNo: core.String(r.RefundOrderID),
			Reason:      core.String(r.Msg),
			NotifyUrl:   core.String(p.RefundNotifyURL),
			Amount: &refunddomestic.AmountReq{
				Currency: core.String("CNY"),
				Refund:   core.Int64(r.RefundAmount),
				Total:    core.Int64(r.RefundAmount),
			},
		},
	)
	if err != nil {
		return nil
	}
	return &RefundOut{
		RefundID: *resp.RefundId,
	}
}

// 支付回掉地址
func (p *confPaymentWx) NotifyUrl(call func(*NotifyData) error) GHandlerFunc {
	return func(c *GContext) {
		ret := &WxNotifyBody{}
		rd, e := p.handle.ParseNotifyRequest(c.Context(), c.Request, ret)
		if e != nil {
			return //失败
		}
		ext, _ := json.Marshal(rd)
		args := &NotifyData{
			PaymentChannel: PAYMENT_WX,
			TransactionID:  ret.TransactionID,
			IsPayment:      true,
			IsRefund:       false,
			IsSuccess:      ret.TradeState == "SUCCESS",
			PlaformUID:     ret.Payer.Openid,
			Amount:         ret.Amount.PayerTotal,
			RefundOrderID:  "",
			OrderID:        ret.OutTradeNo,
			Ext:            string(ext),
		}
		if call(args) != nil {
			return //失败
		}
		//成功处理
	}
}

// 退款回掉地址
func (p *confPaymentWx) RefundNotifyUrl(call func(*NotifyData) error) GHandlerFunc {
	return func(c *GContext) {
		ret := &WxNotifyBody{}
		rd, e := p.handle.ParseNotifyRequest(c.Context(), c.Request, ret)
		if e != nil {
			return //失败
		}
		ext, _ := json.Marshal(rd)
		args := &NotifyData{
			PaymentChannel: PAYMENT_WX,
			TransactionID:  ret.TransactionID,
			IsPayment:      false,
			IsRefund:       true,
			PlaformUID:     ret.Payer.Openid,
			Amount:         ret.Amount.Refund,
			RefundOrderID:  ret.OutRefundNo,
			IsSuccess:      ret.RefundStatus == "SUCCESS",
			OrderID:        ret.OutTradeNo,
			Ext:            string(ext),
		}
		if call(args) != nil {
			return //失败
		}
		//成功处理
	}
}

// h5下单
func (p *confPaymentWx) createOrderH5(r OrderParam) *OrderOut {
	svc := h5.H5ApiService{Client: p.getClient()}
	ctx := context.Background()
	resp, _, err := svc.Prepay(ctx,
		h5.PrepayRequest{
			Appid:         core.String(p.AppID),
			Mchid:         core.String(p.MchID),
			Description:   core.String(r.OrderDesc),
			OutTradeNo:    core.String(r.OrderID),
			TimeExpire:    core.Time(time.Now().Add(time.Minute * 30)),
			Attach:        core.String(r.OrderAttach),
			NotifyUrl:     core.String(p.NotifyURL),
			SupportFapiao: core.Bool(false),
			Amount: &h5.Amount{
				Currency: core.String("CNY"),
				Total:    core.Int64(int64(r.OrderAmount)),
			},
		},
	)
	if err != nil {
		return nil
	}
	return &OrderOut{
		OrderID:        r.OrderID,
		PaymentChannel: r.PaymentChannel,
		PaymentType:    r.PaymentType,
		ResultStr:      *resp.H5Url,
	}
}

// App下单
func (p *confPaymentWx) createOrderApp(r OrderParam) *OrderOut {
	svc := app.AppApiService{Client: p.getClient()}
	ctx := context.Background()
	resp, _, err := svc.Prepay(ctx,
		app.PrepayRequest{
			Appid:         core.String(p.AppID),
			Mchid:         core.String(p.MchID),
			Description:   core.String(r.OrderDesc),
			OutTradeNo:    core.String(r.OrderID),
			TimeExpire:    core.Time(time.Now().Add(time.Minute * 30)),
			Attach:        core.String(r.OrderAttach),
			NotifyUrl:     core.String(p.NotifyURL),
			SupportFapiao: core.Bool(false),
			Amount: &app.Amount{
				Currency: core.String("CNY"),
				Total:    core.Int64(int64(r.OrderAmount)),
			},
		},
	)
	if err != nil {
		return nil
	}
	return &OrderOut{
		OrderID:        r.OrderID,
		PaymentChannel: r.PaymentChannel,
		PaymentType:    r.PaymentType,
		ResultStr:      *resp.PrepayId,
	}
}
func (p *confPaymentWx) createOrderJsApi(r OrderParam) *OrderOut {
	svc := jsapi.JsapiApiService{Client: p.getClient()}
	ctx := context.Background()
	resp, _, err := svc.Prepay(ctx,
		jsapi.PrepayRequest{
			SpAppid:       core.String(p.AppID),
			SpMchid:       core.String(p.MchID),
			Description:   core.String(r.OrderDesc),
			OutTradeNo:    core.String(r.OrderID),
			TimeExpire:    core.Time(time.Now().Add(time.Minute * 30)),
			Attach:        core.String(r.OrderAttach),
			NotifyUrl:     core.String(p.NotifyURL),
			SupportFapiao: core.Bool(false),
			Amount: &jsapi.Amount{
				Currency: core.String("CNY"),
				Total:    core.Int64(int64(r.OrderAmount)),
			},
		},
	)
	if err != nil {
		return nil
	}
	return &OrderOut{
		OrderID:        r.OrderID,
		PaymentChannel: r.PaymentChannel,
		PaymentType:    r.PaymentType,
		ResultStr:      *resp.PrepayId,
	}
}
func (p *confPaymentWx) createOrderNative(r OrderParam) *OrderOut {
	svc := native.NativeApiService{Client: p.getClient()}
	ctx := context.Background()
	resp, _, err := svc.Prepay(ctx,
		native.PrepayRequest{
			SpAppid:       core.String(p.AppID),
			SpMchid:       core.String(p.MchID),
			Description:   core.String(r.OrderDesc),
			OutTradeNo:    core.String(r.OrderID),
			TimeExpire:    core.Time(time.Now().Add(time.Minute * 30)),
			Attach:        core.String(r.OrderAttach),
			NotifyUrl:     core.String(p.NotifyURL),
			SupportFapiao: core.Bool(false),
			Amount: &native.Amount{
				Currency: core.String("CNY"),
				Total:    core.Int64(int64(r.OrderAmount)),
			},
		},
	)
	if err != nil {
		return nil
	}
	return &OrderOut{
		OrderID:        r.OrderID,
		PaymentChannel: r.PaymentChannel,
		PaymentType:    r.PaymentType,
		ResultStr:      *resp.CodeUrl,
	}
}
func (p *confPaymentWx) closeOrderH5(r OrderParam) {
	svc := h5.H5ApiService{Client: p.getClient()}
	ctx := context.Background()
	svc.CloseOrder(ctx,
		h5.CloseOrderRequest{
			Mchid:      core.String(p.MchID),
			OutTradeNo: core.String(r.OrderID),
		},
	)
}
func (p *confPaymentWx) closeOrderApp(r OrderParam) {
	svc := app.AppApiService{Client: p.getClient()}
	ctx := context.Background()
	svc.CloseOrder(ctx,
		app.CloseOrderRequest{
			Mchid:      core.String(p.MchID),
			OutTradeNo: core.String(r.OrderID),
		},
	)
}
func (p *confPaymentWx) closeOrderJsApi(r OrderParam) {
	svc := jsapi.JsapiApiService{Client: p.getClient()}
	ctx := context.Background()
	svc.CloseOrder(ctx,
		jsapi.CloseOrderRequest{
			SpMchid:    core.String(p.MchID),
			OutTradeNo: core.String(r.OrderID),
		},
	)
}
func (p *confPaymentWx) closeOrderNative(r OrderParam) {
	svc := native.NativeApiService{Client: p.getClient()}
	ctx := context.Background()
	svc.CloseOrder(ctx,
		native.CloseOrderRequest{
			SpMchid:    core.String(p.MchID),
			OutTradeNo: core.String(r.OrderID),
		},
	)
}

// 获取连接
func (p *confPaymentWx) getClient() *core.Client {
	if p.client == nil {
		mchPrivateKey, err := utils.LoadPrivateKey(p.PriviteKeyPem)
		if err != nil {
			Error("加载微信私钥错误")
			panic("加载微信私钥错误")
		}
		ctx := context.Background()
		// 使用商户私钥等初始化 client，并使它具有自动定时获取微信支付平台证书的能力
		opts := []core.ClientOption{
			option.WithWechatPayAutoAuthCipher(p.MchID, p.MchIdNumber, mchPrivateKey, p.MchApiV3Key),
		}
		p.client, err = core.NewClient(ctx, opts...)
		if err != nil {
			Error("创建微信支付客户端失败 %s", err.Error())
			panic("创建微信支付客户端失败 " + err.Error())
		}
		downloader.MgrInstance().RegisterDownloaderWithClient(ctx, p.client, p.MchID, p.MchApiV3Key)
		certificateVisitor := downloader.MgrInstance().GetCertificateVisitor(p.MchID)
		// 3. 使用证书访问器初始化 `notify.Handler`
		p.handle = notify.NewNotifyHandler(p.MchApiV3Key, verifiers.NewSHA256WithRSAVerifier(certificateVisitor))
	}
	return p.client
}

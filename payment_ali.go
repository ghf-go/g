package g

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"
)

type confPaymentZli struct {
	AppID         string `yaml:"app_id"`
	NotifyURL     string `yaml:"notify_url"`
	AppPublicPem  string `yaml:"app_public_pem"`
	AliPublicPem  string `yaml:"ali_public_pem"`
	RootPem       string `yaml:"root_pem"`
	AliGateWay    string `yaml:"gateway"`
	AliPrivateKey string `yaml:"alipay_private_key"`

	isInit bool //是否已经初始化

	privateKey *rsa.PrivateKey
}
type aliResponseData struct {
	Sign  string `json:"sign"`
	H5Pay struct {
		Code       string `json:"code"`
		Msg        string `json:"msg"`
		OrderID    string `json:"out_trade_no"`
		AliOrderID string `json:"trade_no"`
		Amount     string `json:"total_amount"`
		SellerID   string `json:"seller_id"`
	} `json:"alipay_trade_wap_pay_response"`
	AppPay struct {
		Code       string `json:"code"`
		Msg        string `json:"msg"`
		OrderID    string `json:"out_trade_no"`
		AliOrderID string `json:"trade_no"`
		Amount     string `json:"total_amount"`
		SellerID   string `json:"seller_id"`
	} `json:"alipay_trade_app_pay_response"`
	Refund struct {
		Code       string `json:"code"`
		Msg        string `json:"msg"`
		OrderID    string `json:"out_trade_no"`
		AliOrderID string `json:"trade_no"`
		Amount     string `json:"refund_fee"`
		SellerID   string `json:"buyer_logon_id"`
	} `json:"alipay_trade_refund_response"`
}

func (p *confPaymentZli) init() {
	b, _ := pem.Decode([]byte(p.AliPrivateKey))
	if b == nil {
		Error("阿里支付秘钥错误")
		panic("阿里支付秘钥错误")
	}
	var e error
	p.privateKey, e = x509.ParsePKCS1PrivateKey(b.Bytes)
	if e != nil {
		Error("阿里支付秘钥错误 %s", e.Error())
		panic("阿里支付秘钥错误")
	}
}

// 创建订单
func (p *confPaymentZli) CreateOrder(r OrderParam) *OrderOut {
	if !p.isInit {
		p.init()
	}
	switch r.PaymentType {
	case "h5":
		return p.createOrderH5(r)
	case "app":
		return p.createOrderApp(r)
		// case "pc":
		// 	return p.createOrderPC(r)

	}
	return nil
}

//	func (p *confPaymentZli) createOrderPC(r OrderParam) *OrderOut {
//		args := p.newParam("alipay.trade.page.pay", map[string]string{
//			"out_trade_no": r.OrderID,
//			"total_amount": fmt.Sprintf("%02f", r.OrderAmount/100),
//			"subject":      r.OrderTitle + r.OrderDesc,
//			"product_code": "FAST_INSTANT_TRADE_PAY",
//		})
//		ret := &aliResponseData{}
//		e := p.exec(args, ret)
//		if e != nil {
//			return nil
//		}
//		if ret.H5Pay.Code != "10000" {
//			return nil
//		}
//		outData, _ := json.Marshal(ret)
//		return &OrderOut{
//			PaymentChannel: PAYMENT_ALI,
//			PaymentType:    "h5",
//			OrderID:        ret.H5Pay.OrderID,
//			ResultStr:      string(outData),
//		}
//	}
func (p *confPaymentZli) createOrderH5(r OrderParam) *OrderOut {
	args := p.newParam("alipay.trade.wap.pay", map[string]string{
		"out_trade_no": r.OrderID,
		"total_amount": fmt.Sprintf("%02f", r.OrderAmount/100),
		"subject":      r.OrderTitle + r.OrderDesc,
	})
	ret := &aliResponseData{}
	e := p.exec(args, ret)
	if e != nil {
		return nil
	}
	if ret.H5Pay.Code != "10000" {
		return nil
	}
	outData, _ := json.Marshal(ret)
	return &OrderOut{
		PaymentChannel: PAYMENT_ALI,
		PaymentType:    "h5",
		OrderID:        ret.H5Pay.OrderID,
		ResultStr:      string(outData),
	}
}
func (p *confPaymentZli) createOrderApp(r OrderParam) *OrderOut {
	args := p.newParam("alipay.trade.app.pay", map[string]string{
		"out_trade_no": r.OrderID,
		"total_amount": fmt.Sprintf("%02f", r.OrderAmount/100),
		"subject":      r.OrderTitle + r.OrderDesc,
	})
	ret := &aliResponseData{}
	e := p.exec(args, ret)
	if e != nil {
		return nil
	}
	if ret.AppPay.Code != "10000" {
		return nil
	}
	outData, _ := json.Marshal(ret)
	return &OrderOut{
		PaymentChannel: PAYMENT_ALI,
		PaymentType:    "app",
		OrderID:        ret.AppPay.OrderID,
		ResultStr:      string(outData),
	}
}

// 关闭订单
func (p *confPaymentZli) CloseOrder(r OrderParam) {
	if !p.isInit {
		p.init()
	}
	args := p.newParam("alipay.trade.close", map[string]string{
		"out_trade_no": r.OrderID,
	})
	ret := &aliResponseData{}
	p.exec(args, ret)

}

// 退款
func (p *confPaymentZli) Refund(r RefundParam) *RefundOut {
	args := p.newParam("alipay.trade.refund", map[string]string{
		"out_trade_no":  r.OrderID,
		"refund_amount": fmt.Sprintf("%02f", r.OrderAmount/100),
	})
	ret := &aliResponseData{}
	e := p.exec(args, ret)
	if e != nil {
		return nil
	}
	if ret.Refund.Code != "10000" {
		return nil
	}
	// outData, _ := json.Marshal(ret)
	return &RefundOut{
		RefundID: ret.Refund.AliOrderID,
	}
}

// 支付回掉地址
func (p *confPaymentZli) NotifyUrl(func(*NotifyData) error) GHandlerFunc {
	return nil
}

// 退款回掉地址
func (p *confPaymentZli) RefundNotifyUrl(func(*NotifyData) error) GHandlerFunc {
	return nil
}

// 执行请求
func (p *confPaymentZli) exec(arg *aliReqParam, ret any) error {
	arg.build()
	params := []string{}
	for k, v := range arg.data {
		params = append(params, fmt.Sprintf("%s=%s", k, v))
	}

	dd := strings.NewReader(string(strings.Join(params, "&")))
	req, e := http.NewRequest(http.MethodPost, p.AliGateWay, dd)
	if e != nil {
		return e
	}
	rep, e := http.DefaultClient.Do(req)
	if e != nil {
		return e
	}
	defer rep.Body.Close()
	rd, e := io.ReadAll(rep.Body)
	if e != nil {
		return e
	}
	return json.Unmarshal(rd, ret)
}
func (p *confPaymentZli) newParam(cmd string, obj any) *aliReqParam {
	bd, _ := json.Marshal(obj)
	return &aliReqParam{
		data: map[string]string{"app_id": p.AppID,
			"method":      cmd,
			"format":      "JSON",
			"charset":     "utf-8",
			"sign_type":   "RSA2",
			"timestamp":   time.Now().Format("2006-01-02 15:04:05"),
			"version":     "1.0",
			"notify_url":  p.NotifyURL,
			"biz_content": string(bd),
		},
		keys: []string{"app_id", "method", "format", "charset", "sign_type", "timestamp", "version", "notify_url", "biz_content"},
		conf: p,
	}
}

type aliReqParam struct {
	data map[string]string
	keys []string
	conf *confPaymentZli
}

func (p *aliReqParam) put(key, val string) {
	p.data[key] = val
	p.keys = append(p.keys, key)
}
func (p aliReqParam) Swap(i, j int) {
	p.keys[j], p.keys[i] = p.keys[i], p.keys[j]
}
func (p aliReqParam) Len() int {
	return len(p.keys)
}
func (p aliReqParam) Less(i, j int) bool {
	return p.keys[i] > p.keys[j]
}

func (p aliReqParam) build() {
	sort.Sort(p)
	retstr := ""
	isFirst := true
	for _, k := range p.keys {
		if isFirst {
			isFirst = false
			retstr += fmt.Sprintf("%s=%s", k, p.data[k])
		} else {
			retstr += fmt.Sprintf("&%s=%s", k, p.data[k])
		}
	}
	hasNew := sha256.New()
	hasNew.Write([]byte(retstr))

	rd, e := rsa.SignPKCS1v15(rand.Reader, p.conf.privateKey, crypto.SHA256, hasNew.Sum(nil))
	if e != nil {
		Error("阿里支付秘钥错误")
		panic("阿里支付秘钥错误")
	}
	p.data["sign"] = base64.StdEncoding.EncodeToString(rd)
}

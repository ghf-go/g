package g

import (
	"encoding/json"
	"log"

	"github.com/sideshow/apns2"
	"github.com/sideshow/apns2/certificate"
	"github.com/sideshow/apns2/token"
)

type appPushConf struct {
	IosConf *iosApnsConf `yaml:"ios"`
}

// https://github.com/sideshow/apns2
// https://github.com/firebase/firebase-admin-go
// IOS推送配置
type iosApnsConf struct {
	P12File   []byte        `yaml:"p12file"`
	P12Passwd string        `yaml:"passwd"`
	JwtFile   []byte        `yaml:"jwt"`
	KeyId     string        `yaml:"key_id"`
	TeamId    string        `yaml:"team_id"`
	Env       string        `yaml:"env"`
	clinet    *apns2.Client //链接
}

// 获取配置信息
func (c *iosApnsConf) getClient() *apns2.Client {
	if c.clinet == nil {
		if len(c.P12File) > 0 {
			cert, err := certificate.FromP12Bytes(c.P12File, c.P12Passwd)
			if err != nil {
				log.Fatal("Cert Error:", err)
			}
			if c.Env == "dev" {
				c.clinet = apns2.NewClient(cert).Development()
			} else {
				c.clinet = apns2.NewClient(cert).Production()
			}
		} else {
			authKey, err := token.AuthKeyFromBytes(c.JwtFile)
			if err != nil {
				log.Fatal("token error:", err)
			}
			token := &token.Token{
				AuthKey: authKey,
				KeyID:   c.KeyId,
				TeamID:  c.TeamId,
			}
			if c.Env == "dev" {
				c.clinet = apns2.NewTokenClient(token).Development()
			} else {
				c.clinet = apns2.NewTokenClient(token).Production()
			}
		}
	}
	return c.clinet
}

// 发送IOS的推送通知
func SendIosPush(e *GEngine, deviceToken, topic string, data any) {
	msg, _ := json.Marshal(data)
	notification := &apns2.Notification{
		DeviceToken: deviceToken,
		Topic:       topic,
		Payload:     msg,
	}
	e.conf.AppPushConf.IosConf.getClient().Push(notification)
}

// 发送Andorid的推送通知
func SendAndroidPush(e *GEngine, deviceToken, topic string, data any) {

}

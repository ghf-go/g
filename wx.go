package g

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/url"
	"time"
)

// 微信公众号配置
type wxWeb struct {
	AppId       string
	AppSecret   string
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
}
type Wx struct {
	web wxWeb
}

// 创建微信配置
func NewWx() *Wx {
	return &Wx{}
}

// 接收到微信推送的消息
type wxMsg struct {
	c            *GContext `xml:"-"`
	ToUserName   string    `xml:"ToUserName"`
	FromUserName string    `xml:"FromUserName"`
	CreateTime   uint64    `xml:"CreateTime"`
	MsgType      string    `xml:"MsgType"`
	Content      string    `xml:"Content"`
	MsgId        int64     `xml:"MsgId"`
	MsgDataId    string    `xml:"MsgDataId"`
	Idx          string    `xml:"Idx"`

	PicUrl       string `xml:"PicUrl"`
	MediaId      string `xml:"MediaId"`
	Format       string `xml:"Format"`
	Recognition  string `xml:"Recognition"`
	ThumbMediaId string `xml:"ThumbMediaId"`

	Label      string  `xml:"Label"`
	Location_Y float64 `xml:"Location_Y"`
	Location_X float64 `xml:"Location_X"`
	Scale      int     `xml:"Scale"`

	Title       string `xml:"Title"`
	Description string `xml:"Description"`
	Url         string `xml:"Url"`

	Event     string  `xml:"Event"`
	EventKey  string  `xml:"EventKey"`
	Ticket    string  `xml:"Ticket"`
	Latitude  float64 `xml:"Latitude"`
	Longitude float64 `xml:"Longitude"`
	Precision float64 `xml:"Precision"`
}

// 创建微信接收消息方法
func WxHandle(call func(c *GContext, msg *wxMsg)) GHandlerFunc {
	return func(c *GContext) {
		msg := &wxMsg{}
		data, e := io.ReadAll(c.Request.Body)
		if e != nil {
			c.WebJsonFail(-1, e.Error())
			return
		}

		defer c.Request.Body.Close()
		e = xml.Unmarshal(data, msg)
		if e != nil {
			c.WebJsonFail(-1, e.Error())
			return
		}
		msg.c = c
		call(c, msg)
	}
}

// 微信服务器验证接口使用的时候注册即可
func WxDomainCheckAction(c *GContext) {
	c.Writer.Write([]byte(c.Request.URL.Query().Get("echostr")))
}

// 服务器端获取的Token信息
type wxservertoken struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
}

type WxWebToken struct {
	AccessToken    string `json:"access_token"`
	ExpiresIn      int    `json:"expires_in"`
	RefreshToken   string `json:"refresh_token"`
	OpenId         string `json:"openid"`
	Scope          string `json:"scope"`
	IsSnapshotuser int    `json:"is_snapshotuser"`
	UnionId        string `json:"unionid"`
	ErrCode        int    `json:"errcode"`
	ErrMsg         string `json:"errmsg"`
}

// 微信用户信息
type WxUserInfo struct {
	OpenId     string   `json:"openid"`
	NickName   string   `json:"nickname"`
	Sex        int      `json:"sex"`
	Province   string   `json:"province"`
	City       string   `json:"city"`
	Country    string   `json:"country"`
	HeadImgUrl string   `json:"headimgurl"`
	UnionId    string   `json:"unionid"`
	Privilege  []string `json:"privilege"`

	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}

// 服务器端获取Token
func (w *Wx) ServerTokenFirst() {
	url := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/token?grant_type=client_credential&appid=%s&secret=%s", w.web.AppId, w.web.AppSecret)
	st := &wxservertoken{}
	if GetJSON(url, st) != nil {
		w.web.AccessToken = st.AccessToken
		w.web.ExpiresIn = st.ExpiresIn
	}
}

// 获取公众号后台的Token
func (w *Wx) ServerToken(force_refresh ...bool) {
	rd := Map{
		"grant_type":    "client_credential",
		"appid":         w.web.AppId,
		"secret":        w.web.AppSecret,
		"force_refresh": false,
	}
	if len(force_refresh) > 0 {
		rd["force_refresh"] = force_refresh[0]
	}
	st := &wxservertoken{}
	if PostJSON("https://api.weixin.qq.com/cgi-bin/stable_token", rd, st) == nil {
		w.web.AccessToken = st.AccessToken
		w.web.ExpiresIn = st.ExpiresIn
	}
}

// 获取公众号后台配置的菜单
func (w *Wx) ServerGetMenu() Map {
	ret := Map{}
	GetJSON(fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/get_current_selfmenu_info?access_token=%s", w.web.AccessToken), ret)
	return ret
}

// 删除公众号的菜单
func (w *Wx) ServerDelMenu() bool {
	ret := Map{}
	GetJSON(fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/menu/delete?access_token=%s", w.web.AccessToken), ret)
	return ret.GetInt("errcode") == 0
}

// 更新公众号菜单
func (w *Wx) ServerCreateMenu(data any) bool {
	ret := Map{}
	PostJSON(fmt.Sprintf(" https://api.weixin.qq.com/cgi-bin/menu/create?access_token=%s", w.web.AccessToken), data, ret)
	return ret.GetInt("errcode") == 0
}

// 网页授权基本功能
func (w *Wx) WebUrlBase(returnurl string) string {
	return fmt.Sprintf("https://open.weixin.qq.com/connect/oauth2/authorize?appid=%s&redirect_uri=%s&response_type=code&scope=snsapi_base&state=123#wechat_redirec", w.web.AppId, url.QueryEscape(returnurl))
}

// 网页授权用户信息
func (w *Wx) WebUrlUserInfo(returnurl string) string {
	return fmt.Sprintf("https://open.weixin.qq.com/connect/oauth2/authorize?appid=%s&redirect_uri=%s&response_type=code&scope=snsapi_userinfo&state=123#wechat_redirec", w.web.AppId, url.QueryEscape(returnurl))
}

// 获取web的TOKEN
func (w *Wx) WebGetToken(code string, ret *WxWebToken) error {
	return GetJSON(fmt.Sprintf("https://api.weixin.qq.com/sns/oauth2/access_token?appid=%s&secret=%s&code=%s&grant_type=authorization_code", w.web.AppId, w.web.AppSecret, code), ret)
}

// 刷新WEB TOKEN
func (w *Wx) WebRefreshToken(refresh_token string, ret *WxWebToken) error {
	return GetJSON(fmt.Sprintf("https://api.weixin.qq.com/sns/oauth2/refresh_token?appid=%s&grant_type=refresh_token&refresh_token=%s", w.web.AppId, refresh_token), ret)
}

// 获取用户信息
func (w *Wx) GetUserInfo(token, openid string, ret *WxUserInfo) error {
	return GetJSON(fmt.Sprintf("https://api.weixin.qq.com/sns/userinfo?access_token=%s&openid=%s&lang=zh_CN", token, openid), ret)
}

// 检查Token是否有效
func (w *Wx) WebCheckToken(token, openid string) bool {
	ret := &WxUserInfo{}
	if GetJSON(fmt.Sprintf("https://api.weixin.qq.com/sns/auth?access_token=%s&openid=%s", token, openid), ret) == nil {
		return ret.ErrCode == 0
	}
	return false
}

// 上传临时素材
func (w *Wx) UploadTmpMedia(mediaType, fileName string, data []byte) string {
	ret := Map{}
	if PostFileByteJSON(fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/media/upload?access_token=%s&type=%s", w.web.AccessToken, mediaType), "media", fileName, data, ret) != nil {
		return ""
	}
	return ret.GetString("media_id", "")
}

// 获取临时素材信息
func (w *Wx) GetTmpMedia(mid string) string {
	ret := Map{}
	if GetJSON(fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/media/get?access_token=%s&media_id=%s", w.web.AccessToken, mid), ret) != nil {
		return ""
	}
	return ret.GetString("video_url", "")
}

// 上传素材
func (w *Wx) UploadMedia(mediaType, fileName string, data []byte) string {
	ret := Map{}
	if PostFileByteJSON(fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/material/add_material?access_token=%s&type=%s", w.web.AccessToken, mediaType), "media", fileName, data, ret) != nil {
		return ""
	}
	return ret.GetString("media_id", "")
}

// 上传图文消息的图片，图片要小于1M
func (w *Wx) UploadImgMsg(mediaType, fileName string, data []byte) string {
	ret := Map{}
	if PostFileByteJSON(fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/media/uploadimg?access_token=%s", w.web.AccessToken), "media", fileName, data, ret) != nil {
		return ""
	}
	return ret.GetString("url", "")
}

// ////消息相关

// 回复文本消息
func (m *wxMsg) SendText(msg string) {
	m.c.Writer.Write([]byte(fmt.Sprintf("<xml><ToUserName><![CDATA[%s]]></ToUserName><FromUserName><![CDATA[%s]]></FromUserName><CreateTime>%d</CreateTime><MsgType><![CDATA[text]]></MsgType><Content><![CDATA[%s]]></Content></xml>", m.FromUserName, m.ToUserName, time.Now().Unix(), msg)))
}

// 回复图片消息
func (m *wxMsg) SendImg(MediaId string) {
	m.c.Writer.Write([]byte(fmt.Sprintf("<xml><ToUserName><![CDATA[%s]]></ToUserName><FromUserName><![CDATA[%s]]></FromUserName><CreateTime>%d</CreateTime><MsgType><![CDATA[image]]></MsgType><Image><MediaId><![CDATA[%s]]></MediaId></Image></xml>", m.FromUserName, m.ToUserName, time.Now().Unix(), MediaId)))

}

// 回复声音
func (m *wxMsg) SendVoice(MediaId string) {
	m.c.Writer.Write([]byte(fmt.Sprintf("<xml><ToUserName><![CDATA[%s]]></ToUserName><FromUserName><![CDATA[%s]]></FromUserName><CreateTime>%d</CreateTime><MsgType><![CDATA[voice]]></MsgType><Voice><MediaId><![CDATA[%s]]></MediaId></Voice></xml>", m.FromUserName, m.ToUserName, time.Now().Unix(), MediaId)))

}

// 回复视频
func (m *wxMsg) SendVideo(MediaId, title, desc string) {
	m.c.Writer.Write([]byte(fmt.Sprintf("<xml><ToUserName><![CDATA[%s]]></ToUserName><FromUserName><![CDATA[%s]]></FromUserName><CreateTime>%d</CreateTime><MsgType><![CDATA[video]]></MsgType><Video><MediaId><![CDATA[%s]]></MediaId><Title><![CDATA[%s]]></Title><Description><![CDATA[%s]]></Description></Video></xml>", m.FromUserName, m.ToUserName, time.Now().Unix(), MediaId, title, desc)))

}

// 回复音乐
func (m *wxMsg) SendMusic(title, desc, thumid, murl, hqurl string) {
	m.c.Writer.Write([]byte(fmt.Sprintf("<xml><ToUserName><![CDATA[%s]]></ToUserName><FromUserName><![CDATA[%s]]></FromUserName><CreateTime>%d</CreateTime><MsgType><![CDATA[music]]></MsgType><Music><ThumbMediaId><![CDATA[%s]]></ThumbMediaId><Title><![CDATA[%s]]></Title><Description><![CDATA[%s]]></Description><MusicUrl><![CDATA[%]]></MusicUrl><HQMusicUrl><![CDATA[%s]]></HQMusicUrl></Music></xml>", m.FromUserName, m.ToUserName, time.Now().Unix(), thumid, title, desc, murl, hqurl)))

}

// 回复图文 map {"title":xxx,"desc":"","picurl":"","url":""}
func (m *wxMsg) SendNews(data []Map) {
	ret := fmt.Sprintf("<xml><ToUserName><![CDATA[%s]]></ToUserName><FromUserName><![CDATA[%s]]></FromUserName><CreateTime>%d</CreateTime><MsgType><![CDATA[news]]></MsgType>", m.FromUserName, m.ToUserName, time.Now().Unix())
	ret += fmt.Sprintf("<ArticleCount>%d</ArticleCount> <Articles>", len(data))
	for _, item := range data {
		ret += fmt.Sprintf("<item><Title><![CDATA[%s]]></Title><Description><![CDATA[%s]]></Description><PicUrl><![CDATA[%s]]></PicUrl><Url><![CDATA[%s]]></Url> </item>",
			item.GetString("title"), item.GetString("desc"), item.GetString("picurl"), item.GetString("url"))
	}
	ret += "</Articles></xml>"
	m.c.Writer.Write([]byte(ret))

}

package g

import (
	"fmt"
	"net/url"
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

package g

import (
	"encoding/json"
	"net"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

const (
	CT_HTTP = 0
	CT_TCP  = 1
	CT_UPD  = 2
)

type GContext struct {
	engine            *GEngine
	isCancel          bool
	webHf             []GHandlerFunc
	webHfCurrentIndex int
	clientType        int //客户端类型
	clientIP          string
	conn              net.Conn
	_httpWriter       http.ResponseWriter
	Request           *http.Request
	Writer            *GResponseWrite
	wscon             *websocket.Conn
	session           map[string]any //Session存储的数据
}
type GHandlerFunc func(*GContext)

func (c *GContext) flush() {
	if c.Writer.statusCode > 0 {
		c._httpWriter.WriteHeader(c.Writer.statusCode)
	}
	c._httpWriter.Write(c.Writer.data.Bytes())
}

// 设置用户ID
func (c *GContext) SetUserId(uid uint64) {
	c.SessionSet("uid", uid)
}

// 回去用户ID
func (c *GContext) GetUserId() uint64 {
	return c.Session("uid").(uint64)
}

// 检查账号是否登录
func (c *GContext) CheckoutUserLogin() bool {
	return c.GetUserId() > 0
}

// 检查管理员是否登录
func (c *GContext) CheckoutAdminLogin() bool {
	return c.GetAdminId() > 0
}

// 获取管理员ID
func (c *GContext) SetAdminId(uid uint64) {
	c.SessionSet("admin_uid", uid)
}

// 获取管理员id
func (c *GContext) GetAdminId() uint64 {
	r := c.Session("admin_uid")
	if r == nil {
		return 0
	}
	return r.(uint64)
}

// 设置session字符串
func (c *GContext) SessionSetString(key, val string) {
	c.SessionSet(key, val)
}

// 获取session字符串
func (c *GContext) SessionGetString(key string) string {
	r := c.Session(key)
	if r != nil {
		return r.(string)
	}
	return ""
}

// 设置Session
func (c *GContext) SessionSet(key string, val any) {
	c.session[key] = val
}

// 删除session
func (c *GContext) SessionDel(key ...string) {
	if len(key) == 0 {
		c.SessionDestory()
	} else {
		for _, k := range key {
			delete(c.session, k)
		}
	}
}

// 获取session信息
func (c *GContext) Session(key ...string) any {
	if len(key) == 1 {
		if r, ok := c.session[key[0]]; ok {
			return r
		}
		return nil
	}
	ret := map[string]any{}
	for _, k := range key {
		if r, ok := c.session[k]; ok {
			ret[k] = r
		} else {
			ret[k] = nil
		}
	}
	return ret
}

// 清空session
func (c *GContext) SessionDestory() {
	c.session = map[string]any{}
}

// 下一个方法
func (c *GContext) Next() {
	if c.webHfCurrentIndex >= len(c.webHf) {
		return
	}
	if c.clientType == CT_HTTP {
		c.webHfCurrentIndex += 1
		c.webHf[c.webHfCurrentIndex-1](c)

	} else if c.clientType == CT_TCP {
		c.webHf[c.webHfCurrentIndex](c)
		c.webHfCurrentIndex += 1
	} else if c.clientType == CT_UPD {
		c.webHf[c.webHfCurrentIndex](c)
		c.webHfCurrentIndex += 1
	}
}

// 获取客户端IP
func (c *GContext) GetClientIP() string {
	if c.clientIP == "" {
		if c.clientType == CT_HTTP {
			c.clientIP = GetRequestIP(c.Request)
		} else if c.clientType == CT_TCP {
			c.clientIP = c.conn.RemoteAddr().String()
		} else if c.clientType == CT_UPD {
			c.clientIP = c.conn.RemoteAddr().String()
		} else {
			c.clientIP = "unknow"
		}
	}
	return c.clientIP
}

func (c *GContext) webJson(obj any) {
	data, e := json.Marshal(obj)
	if e != nil {
		sysDebug("Web 返回json 编码失败 %s -> %v", e.Error(), obj)
	} else {
		c.Writer.Write(data)
	}
}

// web json失败
func (c *GContext) WebJsonFail(code int, msg string) {
	c.webJson(map[string]any{
		"code": code,
		"msg":  msg,
		"data": map[string]any{},
	})
}

// web json成功
func (c *GContext) WebJsonSuccess(obj any) {
	c.webJson(map[string]any{
		"code": 0,
		"msg":  "",
		"data": obj,
	})
}

// 显示模版
func (c *GContext) WebView(obj any, tpl string) {
	c.engine.template.Lookup("_templates").ExecuteTemplate(c.Writer, tpl, obj)
}

// 使用JSONP
func (c *GContext) WebJsonP(call string, data any) {
	c.Writer.Write([]byte(call + "("))
	c.webJson(data)
	c.Writer.Write([]byte(");"))
}

// 获取数据库
func (c *GContext) GetDB() *gorm.DB {
	return c.engine.GetDB()
}

// 获取Redis
func (c *GContext) GetRedis() *redis.Client {
	return c.engine.GetRedis()
}

// 获取Reids
func (c *GContext) GetRedisCluster() *redis.ClusterClient {
	return c.engine.GetRedisCluster()
}
func (c *GContext) Cancel() {
	c.isCancel = true
}
func (c *GContext) IsCancel() bool {
	return c.isCancel
}
func (c *GContext) WsReadMsg() (messageType int, p []byte, err error) {
	return c.wscon.ReadMessage()
}
func (c *GContext) WsReadJSON(obj any) error {
	return c.wscon.ReadJSON(obj)
}
func (c *GContext) WsWriteMessage(messageType int, data []byte) error {
	return c.wscon.WriteMessage(messageType, data)
}
func (c *GContext) WsWriteJSON(obj any) error {
	return c.wscon.WriteJSON(obj)
}

// 绑定POST提交的JSON数据
func (c *GContext) BindJSON(obj any) error {
	r := bindRequestBodyJson(c.Request, obj)
	if r != nil {
		c.WebJsonFail(1, "系统错误")
	}
	return r
}

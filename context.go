package g

import (
	"encoding/json"
	"fmt"
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
}
type GHandlerFunc func(*GContext)

func (c *GContext) flush() {
	if c.Writer.statusCode > 0 {
		c._httpWriter.WriteHeader(c.Writer.statusCode)
	}
	c._httpWriter.Write(c.Writer.data.Bytes())
}

// 数据绑定
func (c *GContext) Bind(obj any) error {
	return nil
}

// 绑定JSON
func (c *GContext) BindJSON(obj any) error {
	return nil
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
		fmt.Println(obj, e)
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
func (c *GContext) WebView(obj any, tpl ...string) {

}

// 使用JSONP
func (c *GContext) WebJsonP(call string, data any) {

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

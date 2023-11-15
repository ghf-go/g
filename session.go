package g

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// jwt session
func jwt_session(g *GContext) {
	cf := g.engine.conf.Session
	srcdata := g.Request.Header.Get(cf.Name)
	if srcdata == "" {
		ck, e := g.Request.Cookie(cf.Name)
		if e == nil {
			srcdata = ck.Value
		}
	}
	if srcdata != "" {
		ddata := AesDecode(cf.JwtKey, srcdata)
		json.Unmarshal([]byte(ddata), g.session)
	}
	g.Next()
	if len(g.session) > 0 {
		data, e := json.Marshal(g.session)
		if e == nil {
			sdata := AesEncode(cf.JwtKey, string(data))
			g.Writer.Header().Add(cf.Name, sdata)
			http.SetCookie(g.Writer, &http.Cookie{
				Name:    cf.Name,
				Value:   sdata,
				Path:    "/",
				Expires: time.Now().Add(time.Second * time.Duration(cf.Expire)),
			})
		}
	}
}

// redis session
func redis_session(c *GContext) {
	cf := c.engine.conf.Session
	sid := c.Request.Header.Get(cf.Name)
	if sid == "" || sid == "null" {
		ck, e := c.Request.Cookie(cf.Name)
		if e != nil {
			sid = ck.Value
		}
	}
	if sid == "" || sid == "null" {
		sid = fmt.Sprintf("%s_%d", cf.RedisKey, time.Now().UnixNano())
	}
	if sid != "" {
		data, e := c.GetRedis().Get(c.Request.Context(), sid).Result()
		if e != nil {
			sysDebug("获取Redis失败 -> %s", e.Error())
		} else {
			json.Unmarshal([]byte(data), &c.session)
			// fmt.Println(c.session, data, sid, "asdsdf")
		}
	}
	c.Next()
	if len(c.session) > 0 {

		rdata, e := json.Marshal(c.session)
		if e == nil {
			c.GetRedis().Set(c.Request.Context(), sid, string(rdata), time.Duration(cf.Expire)*time.Second)
			c.Writer.Header().Add(cf.Name, sid)
			http.SetCookie(c.Writer, &http.Cookie{
				Name:    cf.Name,
				Value:   sid,
				Path:    "/",
				Expires: time.Now().Add(time.Duration(cf.Expire) * time.Second),
			})
		}

	}

}

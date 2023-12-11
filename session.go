package g

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
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
		us := json.NewDecoder(strings.NewReader(ddata))
		us.UseNumber()
		us.Decode(&g.session)
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
	isApp := c.Request.Header.Get("Appid") != ""
	if !isApp {
		if sid == "" || sid == "null" {
			ck, e := c.Request.Cookie(cf.Name)
			if e == nil {
				sid = ck.Value
			}
		}
		if sid == "" || sid == "null" {
			sid = c.Request.URL.Query().Get(cf.Name)
			sysDebug("redis getToken %s", sid)
		}
	}

	if sid == "" || sid == "null" {
		sid = fmt.Sprintf("%s_%d", cf.RedisKey, time.Now().UnixNano())
	}
	if sid != "" {
		data, e := c.GetRedis().Get(c.Request.Context(), sid).Result()
		if e != nil {
			sysDebug("获取Redis失败 -> %s : %s", sid, e.Error())
		} else {
			us := json.NewDecoder(strings.NewReader(data))
			us.UseNumber()
			us.Decode(&c.session)
		}
	}
	c.sid = sid
	c.Next()
	if len(c.session) > 0 {
		rdata, e := json.Marshal(c.session)
		if e == nil {
			c.GetRedis().Set(c.Request.Context(), sid, string(rdata), time.Duration(cf.Expire)*time.Second)
		}
	}
	if isApp {
		c.Writer.Header().Add(cf.Name, sid)
	} else {
		http.SetCookie(c.Writer, &http.Cookie{
			Name:    cf.Name,
			Value:   sid,
			Path:    "/",
			Expires: time.Now().Add(time.Duration(cf.Expire) * time.Second),
		})
	}

}

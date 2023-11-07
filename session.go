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
				Expires: time.Now().Add(cf.Expire * time.Second),
			})
		}
	}
}

// redis session
func redis_session(g *GContext) {
	cf := g.engine.conf.Session
	sid := g.Request.Header.Get(cf.Name)
	if sid != "" {
		data, e := g.GetRedis().Get(g.Request.Context(), sid).Result()
		if e != nil {
			fmt.Println(e.Error())
		} else {
			json.Unmarshal([]byte(data), g.session)
		}
	}
	g.Next()
	if len(g.session) > 0 {
		if sid == "" {
			sid = fmt.Sprintf("%s_%d", cf.RedisKey, time.Now().UnixNano())
		}
		rdata, e := json.Marshal(g.session)
		if e == nil {
			g.GetRedis().Set(g.Request.Context(), sid, string(rdata), cf.Expire*time.Second)
			g.Writer.Header().Add(cf.Name, sid)
			http.SetCookie(g.Writer, &http.Cookie{
				Name:    cf.Name,
				Value:   sid,
				Path:    "/",
				Expires: time.Now().Add(cf.Expire * time.Second),
			})
		}

	}

}

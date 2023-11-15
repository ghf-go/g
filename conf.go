package g

import (
	"fmt"
	"net/smtp"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/plugin/dbresolver"
)

type _dbconf struct {
	Host     string `yaml:"host"`
	UserName string `yaml:"username"`
	Password string `yaml:"passwd"`
}
type dbConf struct {
	DbName          string    `yaml:"dbname"`
	Charset         string    `yaml:"charset"`
	MaxIdleConns    int       `yaml:"max_idle_cons"`
	MaxOpenConns    int       `yaml:"max_open_cons"`
	ConnMaxIdleTime int       `yaml:"con_max_idle_time"`
	ConnMaxLifetime int       `yaml:"con_max_life_time"`
	Write           _dbconf   `yaml:"write"`
	Read            []_dbconf `yaml:"read"`
}

// 发送邮件的配置
type stmpConf struct {
	Host         string `yaml:"host"`
	UserName     string `yaml:"username"`
	Passwd       string `yaml:"passwd"`
	AuthType     string `yaml:"auth_type"`
	TemplatePrex string `yaml:"template_pre"`
}
type redisConf struct {
	Addr            string `yaml:"addr"`
	UserName        string `yaml:"username"`
	Password        string `yaml:"password"`
	MinIdleConns    int    `yaml:"min_idle_cons"`
	MaxIdleConns    int    `yaml:"max_idle_cons"`
	MaxActiveConns  int    `yaml:"max_active_cons"`
	ConnMaxIdleTime int    `yaml:"con_max_idle_time"`
	ConnMaxLifetime int    `yaml:"con_max_life_time"`
}
type appConf struct {
	WebPort int `yaml:"web_port"` //Web 端口
	TcpPort int `yaml:"tcp_port"` //Sock 端口
	UdpPort int `yaml:"udp_port"` //Sock 端口
}

type sessionConf struct {
	Driver   string `yaml:"driver"`
	Name     string `yaml:"session_name"`
	Expire   int    `yaml:"session_expire"`
	JwtKey   string `yaml:"jwtKey"`
	RedisKey string `yaml:"redis_key"`
}
type AppConf struct {
	App     appConf     `yaml:"app"`
	Db      dbConf      `yaml:"db"`
	Redis   redisConf   `yaml:"redis"`
	Session sessionConf `yaml:"session"`
	Stmp    stmpConf    `yaml:"stmp"` //邮件服务器配置
}

// 发送邮件
func (c AppConf) SendMail(to, subject string, isHtml bool, msg []byte) error {
	var auth smtp.Auth
	switch c.Stmp.AuthType {
	case "CRAMMD5":
		auth = smtp.CRAMMD5Auth(c.Stmp.UserName, c.Stmp.Passwd)
	case "HOTMAIL":
		auth = NewHotmailStmpAuth(c.Stmp.UserName, c.Stmp.Passwd)
	default:
		auth = smtp.PlainAuth("", c.Stmp.UserName, c.Stmp.Passwd, c.Stmp.Host)
	}
	content_type := ""
	if isHtml {
		content_type = "Content-Type: text/html; charset=UTF-8"
	} else {
		content_type = "Content-Type: text/plain" + "; charset=UTF-8"
	}
	msg = []byte("To: " + to + "\r\nFrom: " + c.Stmp.UserName + "\r\nSubject: " + subject + "\r\n" + content_type + "\r\n\r\n" + string(msg))

	return smtp.SendMail(c.Stmp.Host, auth, c.Stmp.UserName, []string{to}, msg)
}

// 获取数据连接
func (c AppConf) getMysql() *gorm.DB {

	db, e := gorm.Open(mysql.Open(fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=%s&parseTime=True&loc=Local", c.Db.Write.UserName,
		c.Db.Write.Password, c.Db.Write.Host, c.Db.DbName, c.Db.Charset)), &gorm.Config{})
	if e != nil {
		panic(e.Error())
	}
	rs := []gorm.Dialector{}
	for _, rc := range c.Db.Read {
		rs = append(rs, mysql.Open(fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=%s&parseTime=True&loc=Local", rc.UserName,
			rc.Password, rc.Host, c.Db.DbName, c.Db.Charset)))
	}
	db.Use(dbresolver.Register(dbresolver.Config{
		Sources: []gorm.Dialector{mysql.Open(fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=%s&parseTime=True&loc=Local", c.Db.Write.UserName,
			c.Db.Write.Password, c.Db.Write.Host, c.Db.DbName, c.Db.Charset))},
		Replicas: rs,
	}).SetMaxIdleConns(c.Db.MaxIdleConns).SetMaxOpenConns(c.Db.MaxOpenConns).SetConnMaxIdleTime(time.Minute * time.Duration(c.Db.ConnMaxIdleTime)).SetConnMaxLifetime(time.Minute * time.Duration(c.Db.ConnMaxLifetime)))
	return db
}

// 获取Redis连接
func (c AppConf) getRedis() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:            c.Redis.Addr,
		Username:        c.Redis.UserName,
		Password:        c.Redis.Password,
		MinIdleConns:    c.Redis.MinIdleConns,
		MaxIdleConns:    c.Redis.MaxIdleConns,
		MaxActiveConns:  c.Redis.MaxActiveConns,
		ConnMaxIdleTime: time.Minute * time.Duration(c.Redis.ConnMaxIdleTime),
		ConnMaxLifetime: time.Minute * time.Duration(c.Redis.ConnMaxLifetime),
	})
}

// 获取Redis连接
func (c AppConf) getClusterClient() *redis.ClusterClient {
	return redis.NewClusterClient(&redis.ClusterOptions{
		Addrs:           strings.Split(c.Redis.Addr, ","),
		Username:        c.Redis.UserName,
		Password:        c.Redis.Password,
		MinIdleConns:    c.Redis.MinIdleConns,
		MaxIdleConns:    c.Redis.MaxIdleConns,
		MaxActiveConns:  c.Redis.MaxActiveConns,
		ConnMaxIdleTime: time.Minute * time.Duration(c.Redis.ConnMaxIdleTime),
		ConnMaxLifetime: time.Minute * time.Duration(c.Redis.ConnMaxLifetime),
	})
}

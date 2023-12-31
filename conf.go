package g

import (
	"fmt"
	"net/smtp"
	"strings"
	"time"

	"github.com/qiniu/go-sdk/v7/auth"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/plugin/dbresolver"
)

// 存储配置
type _storeConf struct {
	Driver    string `yaml:"driver"`
	AccessKey string `yaml:"access_key"`
	SecretKey string `yaml:"secret_key"`
	Bucket    string `yaml:"bucket"`
	ZoneHost  string `yaml:"zone"`
	CdnDomain string `yaml:"cdnhost"`
	qini      *auth.Credentials
}

// 数据库配置
type _dbconf struct {
	Host     string `yaml:"host"`
	UserName string `yaml:"username"`
	Password string `yaml:"passwd"`
}

// 支付配置
type paymentConf struct {
	AliConf     *confPaymentZli     `yaml:"ali"`
	WxConf      *confPaymentWx      `yaml:"wx"`
	YinLianCOnf *confPaymentYinLian `yaml:"yinlian"`
}

// 数据库配置
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

// redis 配置
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

// 端口配置
type appConf struct {
	WebPort     int    `yaml:"web_port"`     //Web 端口
	TcpPort     int    `yaml:"tcp_port"`     //Sock 端口
	UdpPort     int    `yaml:"udp_port"`     //Sock 端口
	TemplateDir string `yaml:"template_dir"` //模板路径
}

// 微信服务号配置
type wxWeb struct {
	AppId       string `yaml:"app_id"`
	AppSecret   string `yaml:"app_secret"`
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
}
type wxConf struct {
	web *wxWeb `yaml:"app"`
}

// session配置
type sessionConf struct {
	Driver   string `yaml:"driver"`
	Name     string `yaml:"session_name"`
	Expire   int    `yaml:"session_expire"`
	JwtKey   string `yaml:"jwtKey"`
	RedisKey string `yaml:"redis_key"`
}

// 应用配置
type AppConf struct {
	App         appConf      `yaml:"app"`
	Db          dbConf       `yaml:"db"`
	Redis       redisConf    `yaml:"redis"`
	Session     sessionConf  `yaml:"session"`
	Stmp        stmpConf     `yaml:"stmp"`    //邮件服务器配置
	WxConf      *wxConf      `yaml:"wechat"`  //微信配置
	PaymentConf *paymentConf `yaml:"payment"` //支付配置
	StoreConf   *_storeConf  `yaml:"store"`   //存储配置
	AppPushConf *appPushConf `yaml:"app_push_notify"`
}

// 发送邮件
func (c AppConf) SendMail(to, subject string, isHtml bool, msg []byte) error {
	// fmt.Println(c.Stmp)
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

// 获取微信配置
func (c AppConf) GetWxConf() *wxConf {
	return c.WxConf
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

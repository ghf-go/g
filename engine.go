package g

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"html/template"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
	"gopkg.in/yaml.v3"
	"gorm.io/gorm"
)

var wsupgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type _webServer struct {
	webRouter *router_web_node
	webServer *http.Server
	wshandles map[string]GHandlerFunc
}
type _sockServer struct {
	serverFd   net.Listener
	isStop     bool
	beforeFunc func(net.Conn)
}

type GEngine struct {
	Ctx          context.Context
	conf         AppConf //配置
	db           *gorm.DB
	redis        *redis.Client
	redisCluster *redis.ClusterClient
	webServer    *_webServer
	tcpServer    *_sockServer
	udpServer    *_sockServer
	template     *template.Template
	mq           gmqServer
	jobs         gJobServer
}

// 新建引擎
func NewGEngine() *GEngine {
	return &GEngine{
		Ctx: context.Background(),
		webServer: &_webServer{
			webRouter: &router_web_node{
				mid:   []GHandlerFunc{},
				nodes: map[string]*router_web_node{},
				hf:    map[string]map[string]GHandlerFunc{},
			},
		},
		tcpServer: &_sockServer{},
		udpServer: &_sockServer{},
		template:  template.New("_templates"),
		mq:        gmqServer{},
		jobs:      gJobServer{},
	}
}

// 注册模版方法
func (ge *GEngine) SetTemplateFuncMap(tf template.FuncMap) {
	ge.template.Funcs(tf)
}

// 注册模版
func (ge *GEngine) SetTemplate(groupname, path string, ff embed.FS) {
	ffs, e := ff.ReadDir(path)
	if e != nil {
		panic(e.Error())
	}
	for _, item := range ffs {
		if item.IsDir() {
			name := item.Name()
			finame := strings.Replace(item.Name(), path, "", 0)
			ge.SetTemplate(groupname+finame+"_", name, ff)
		} else {
			dd, e := ff.ReadFile(item.Name())
			if e != nil {
				panic(e.Error())
			}
			_, e = ge.template.New(groupname + strings.Replace(item.Name(), path, "", 0)).Parse(string(dd))
			if e != nil {
				panic(e.Error())
			}
		}
	}
}

// 服务运行
func (ge *GEngine) Start(confString []byte) {
	sc := AppConf{}
	e := yaml.Unmarshal(confString, &sc)
	if e != nil {
		panic(e)
	}
	ge.conf = sc
	switch sc.Session.Driver {
	case "jwt":
		ge.webServer.webRouter.mid = append([]GHandlerFunc{jwt_session}, ge.webServer.webRouter.mid...)
	case "redis":
		ge.webServer.webRouter.mid = append([]GHandlerFunc{redis_session}, ge.webServer.webRouter.mid...)
	}
	ge.redis = sc.getRedis()
	ge.db = sc.getMysql()
	// fmt.Println(sc)
	ge.redisCluster = sc.getClusterClient()
	if ge.conf.App.WebPort > 0 {
		ge.webServerStart()
	}
	if len(ge.jobs) > 0 { //启动job
		ge.jobs.start(ge)
	}
	if len(ge.mq) > 0 { //启动队列消费
		ge.mq.start(ge)
	}
	// if ge.conf.App.TcpPort > 0 {
	// 	ge.tcpServerStart()
	// }
	// if ge.conf.App.UdpPort > 0 {
	// 	ge.udpServerStart()
	// }
	//关闭功能
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, syscall.SIGINT, syscall.SIGKILL, syscall.SIGTERM)
	<-sigc
	ct, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	if ge.tcpServer.serverFd != nil {
		ge.tcpServer.isStop = true
	}
	if ge.udpServer.serverFd != nil {
		ge.udpServer.isStop = true
	}

	if ge.webServer.webServer != nil {
		ge.webServer.webServer.Shutdown(ct) //关闭web
	}
	if len(ge.jobs) > 0 { //启动job
		ge.jobs.stop()
	}
	if len(ge.mq) > 0 { //启动队列消费
		ge.mq.stop()
	}
}

// 启动TCP服务
func (ge *GEngine) tcpServerStart() {
	sf, e := net.Listen("tcp", fmt.Sprintf(":%d", ge.conf.App.TcpPort))
	if e != nil {
		panic("tcp服务监听失败" + e.Error())
	}
	ge.tcpServer.serverFd = sf
	go func() {
		for !ge.tcpServer.isStop {
			c, e := ge.tcpServer.serverFd.Accept()
			if e != nil {
				continue
			} else {
				if ge.tcpServer.beforeFunc != nil {
					ge.tcpServer.beforeFunc(c)
				}
			}
		}
	}()
}

// 启动udp服务
func (ge *GEngine) udpServerStart() {
	sf, e := net.Listen("udp", fmt.Sprintf(":%d", ge.conf.App.UdpPort))
	if e != nil {
		panic("udp服务监听失败" + e.Error())
	}
	ge.udpServer.serverFd = sf
	go func() {
		for !ge.udpServer.isStop {
			c, e := ge.udpServer.serverFd.Accept()
			if e != nil {
				continue
			} else {
				if ge.udpServer.beforeFunc != nil {
					ge.udpServer.beforeFunc(c)
				}
			}
		}
	}()
}

// web服务开启
func (ge *GEngine) webServerStart() {
	ge.webServer.webServer = &http.Server{
		Addr:    fmt.Sprintf(":%d", ge.conf.App.WebPort),
		Handler: ge,
	}

	go func() {
		fmt.Println("qidong web", ge)
		if e := ge.webServer.webServer.ListenAndServe(); e != nil {
			panic("开启WEB服务失败" + e.Error())
		}
		fmt.Println("stop web", ge)
	}()
}

// 注册websock
func (ge *GEngine) WebSock() {}

// 网页路由
func (ge *GEngine) WebAny(name string, fen GHandlerFunc) {
	ge.webServer.webRouter.add(name, http.MethodGet, fen)
	ge.webServer.webRouter.add(name, http.MethodPost, fen)
	ge.webServer.webRouter.add(name, http.MethodPut, fen)
	ge.webServer.webRouter.add(name, http.MethodPatch, fen)
	ge.webServer.webRouter.add(name, http.MethodDelete, fen)
	ge.webServer.webRouter.add(name, http.MethodHead, fen)
	ge.webServer.webRouter.add(name, http.MethodTrace, fen)
	ge.webServer.webRouter.add(name, http.MethodOptions, fen)
}
func (ge *GEngine) WebPost(name string, fen GHandlerFunc) {
	ge.webServer.webRouter.add(name, http.MethodPost, fen)
}
func (ge *GEngine) WebGet(name string, fen GHandlerFunc) {
	ge.webServer.webRouter.add(name, http.MethodGet, fen)
}
func (ge *GEngine) WebDelete(name string, fen GHandlerFunc) {
	ge.webServer.webRouter.add(name, http.MethodDelete, fen)
}
func (ge *GEngine) WebPut(name string, fen GHandlerFunc) {
	ge.webServer.webRouter.add(name, http.MethodPut, fen)
}
func (ge *GEngine) WebOptions(name string, fen GHandlerFunc) {
	ge.webServer.webRouter.add(name, http.MethodOptions, fen)
}
func (ge *GEngine) WebTrace(name string, fen GHandlerFunc) {
	ge.webServer.webRouter.add(name, http.MethodTrace, fen)
}
func (ge *GEngine) WebHead(name string, fen GHandlerFunc) {
	ge.webServer.webRouter.add(name, http.MethodHead, fen)
}
func (ge *GEngine) WebPatch(name string, fen GHandlerFunc) {
	ge.webServer.webRouter.add(name, http.MethodPatch, fen)
}
func (ge *GEngine) WebGroup(name string, fen ...GHandlerFunc) *router_web_node {
	return ge.webServer.webRouter.addGroup(name, fen...)
}

// Vue路径
func (ge *GEngine) WebVue() {}

// Vue路径 静态路径
func (ge *GEngine) WebVueHistory(path, dirPath string, fs embed.FS) {
	if !strings.HasSuffix(path, "/") {
		path += "/"
	}
	if !strings.HasSuffix(dirPath, "/") {
		dirPath += "/"
	}
	ge.WebGroup(path, func(g *GContext) {
		const defaultIndex = "index.html"
		fielname := g.Request.URL.Path[len(path):]
		if fielname == "" {
			fielname = defaultIndex
		}
		f, e := fs.Open(dirPath + fielname)
		if e != nil {
			fielname = defaultIndex
		}
		f, _ = fs.Open(dirPath + fielname)
		defer f.Close()
		st, _ := f.Stat()
		modtime := st.ModTime()
		if g.Request.Method != "GET" || g.Request.Method != "HEAD" { //返回304功能
			ims := g.Request.Header.Get("If-Modified-Since")

			if ims != "" && IsTimeZero(modtime) {
				t, err := http.ParseTime(ims)
				if err == nil {
					modtime = modtime.Truncate(time.Second)
					if ret := modtime.Compare(t); ret <= 0 {
						h := g.Writer.Header()
						delete(h, "Content-Type")
						delete(h, "Content-Length")
						delete(h, "Content-Encoding")
						if h.Get("Etag") != "" {
							delete(h, "Last-Modified")
						}
						g.Writer.WriteHeader(http.StatusNotModified)
						return
					}
				}
			}
		}

		data, e := fs.ReadFile(dirPath + fielname)
		if e != nil {
			fielname = defaultIndex
			data, _ = fs.ReadFile(dirPath + defaultIndex)
		}
		g.Writer.WriteHeader(http.StatusOK)
		g.Writer.Header().Set("Last-Modified", modtime.UTC().Format(http.TimeFormat))
		g.Writer.Header().Set("Content-Type", http.DetectContentType(data))
		g.Writer.Write(data)
	})
}

// Socket
func (ge *GEngine) SockAction() {}

// httpHandle
func (ge *GEngine) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// fmt.Println(r.URL, "------")
	path := r.URL.Path
	c := &GContext{
		engine:      ge,
		_httpWriter: w,
		Request:     r,
		clientType:  CT_HTTP,
		Writer: &GResponseWrite{
			header:     w.Header(),
			statusCode: 0,
			data:       bytes.NewBuffer([]byte("")),
		},
		session: map[string]any{},
	}
	// fmt.Println(c)
	if h, ok := ge.webServer.wshandles[path]; ok {
		conn, err := wsupgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()
		c.wscon = conn
		go h(c)
		return
	}
	hs, e := ge.webServer.webRouter.getHandle(path, r.Method, []GHandlerFunc{})
	c.webHf = hs
	if e != nil {
		c.webHf = append(c.webHf, func(g *GContext) {
			g.WebJsonFail(-1, e.Error())
		})
	}
	c.Next()
	c.flush()
}

// 获取数据库
func (ge *GEngine) GetDB() *gorm.DB {
	return ge.db
}

// 获取Redis
func (ge *GEngine) GetRedis() *redis.Client {
	return ge.redis
}

// 获取Reids
func (ge *GEngine) GetRedisCluster() *redis.ClusterClient {
	return ge.redisCluster
}

// 注册任务列表
func (ge *GEngine) AddJob(j ...GJob) {
	ge.jobs = append(ge.jobs, j...)
}

// 注册消息队列
func (ge *GEngine) AddMq(q ...GMQ) {
	ge.mq = append(ge.mq, q...)
}

// 注册Redis队列
func (ge *GEngine) AddMqRedis(redisKey string, msgcall func(msg string)) {
	ge.mq = append(ge.mq, NewMqRedis(redisKey, msgcall))
}

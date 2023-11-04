package g

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
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
	// ge.redis = sc.getRedis()
	// ge.db = sc.getMysql()
	// fmt.Println(sc)
	ge.redisCluster = sc.getClusterClient()
	if ge.conf.App.WebPort > 0 {
		ge.webServerStart()
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

// 静态文件路径
func (ge *GEngine) WebStatic() {}

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
	}
	// fmt.Println(c)
	if h, ok := ge.webServer.wshandles[path]; ok {
		conn, err := wsupgrader.Upgrade(w, r, nil)
		if err == nil {
			defer conn.Close()
			return
		}
		go func() {
			for {
				h(c)
			}
		}()
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

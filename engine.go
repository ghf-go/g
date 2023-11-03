package g

import (
	"context"
	"net/http"

	"github.com/redis/go-redis/v9"
	"gopkg.in/yaml.v3"
	"gorm.io/gorm"
)

type GEngine struct {
	Ctx          context.Context
	conf         AppConf //配置
	db           *gorm.DB
	redis        *redis.Client
	redisCluster *redis.ClusterClient
	webRouter    *router_web_node
}

// 新建引擎
func NewGEngine() *GEngine {
	return &GEngine{
		Ctx:       context.Background(),
		webRouter: &router_web_node{},
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
	ge.redis = sc.getRedis()
	ge.db = sc.getMysql()
	ge.redisCluster = sc.getClusterClient()
}

// 注册websock
func (ge *GEngine) WebSock() {}

// 网页路由
func (ge *GEngine) WebAny(name string, fen WebHandlerFunc) {
	ge.webRouter.add(name, http.MethodGet, fen)
	ge.webRouter.add(name, http.MethodPost, fen)
	ge.webRouter.add(name, http.MethodPut, fen)
	ge.webRouter.add(name, http.MethodPatch, fen)
	ge.webRouter.add(name, http.MethodDelete, fen)
	ge.webRouter.add(name, http.MethodHead, fen)
	ge.webRouter.add(name, http.MethodTrace, fen)
	ge.webRouter.add(name, http.MethodOptions, fen)
}
func (ge *GEngine) WebPost(name string, fen WebHandlerFunc) {
	ge.webRouter.add(name, http.MethodPost, fen)
}
func (ge *GEngine) WebGet(name string, fen WebHandlerFunc) {
	ge.webRouter.add(name, http.MethodGet, fen)
}
func (ge *GEngine) WebDelete(name string, fen WebHandlerFunc) {
	ge.webRouter.add(name, http.MethodDelete, fen)
}
func (ge *GEngine) WebPut(name string, fen WebHandlerFunc) {
	ge.webRouter.add(name, http.MethodPut, fen)
}
func (ge *GEngine) WebOptions(name string, fen WebHandlerFunc) {
	ge.webRouter.add(name, http.MethodOptions, fen)
}
func (ge *GEngine) WebTrace(name string, fen WebHandlerFunc) {
	ge.webRouter.add(name, http.MethodTrace, fen)
}
func (ge *GEngine) WebHead(name string, fen WebHandlerFunc) {
	ge.webRouter.add(name, http.MethodHead, fen)
}
func (ge *GEngine) WebPatch(name string, fen WebHandlerFunc) {
	ge.webRouter.add(name, http.MethodPatch, fen)
}
func (ge *GEngine) WebGroup(name string, fen ...WebHandlerFunc) *router_web_node {
	return ge.webRouter.addGroup(name, fen...)
}

// Vue路径
func (ge *GEngine) WebVue() {}

// 静态文件路径
func (ge *GEngine) WebStatic() {}

// Socket
func (ge *GEngine) SockAction() {}

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

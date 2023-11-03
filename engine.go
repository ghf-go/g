package g

import (
	"context"

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
}

// 新建引擎
func NewGEngine() *GEngine {
	return &GEngine{}
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

// 网页
func (ge *GEngine) WebAny()    {}
func (ge *GEngine) WebPost()   {}
func (ge *GEngine) WebGet()    {}
func (ge *GEngine) WebDelete() {}
func (ge *GEngine) WebPut()    {}
func (ge *GEngine) WebOption() {}
func (ge *GEngine) WebVue()    {}
func (ge *GEngine) WebStatic() {}
func (ge *GEngine) WebGroup()  {}

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

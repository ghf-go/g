package g

import (
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type GContext struct {
	engine   *GEngine
	isCancel bool
}

type SockHandlerFunc func(*GContext)

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

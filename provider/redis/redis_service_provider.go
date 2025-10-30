package redis

import (
	_ "embed"

	"bit-labs.cn/owl"
	"bit-labs.cn/owl/contract/foundation"
	"bit-labs.cn/owl/provider/conf"
	"github.com/redis/go-redis/v9"
)

var _ foundation.ServiceProvider = (*RedisServiceProvider)(nil)

type RedisServiceProvider struct {
	app foundation.Application
}

func (r *RedisServiceProvider) Register() {
	r.app.Register(func(c *conf.Configure) redis.UniversalClient {
		var opt Options
		err := c.GetConfig("redis", &opt)
		owl.PanicIf(err)

		return InitRedis(&opt)
	})
}

func (r *RedisServiceProvider) Boot() {
	// Redis 服务启动时的初始化逻辑
}

//go:embed redis.yaml
var redisYaml string

func (r *RedisServiceProvider) GenerateConf() map[string]string {
	return map[string]string{
		"redis.yaml": redisYaml,
	}
}

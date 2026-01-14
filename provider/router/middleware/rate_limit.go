package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ulule/limiter/v3"
	mgin "github.com/ulule/limiter/v3/drivers/middleware/gin"
	"github.com/ulule/limiter/v3/drivers/store/memory"
)

func RateLimiter(period time.Duration, limit int64) gin.HandlerFunc {
	// 创建速率配置
	rate := limiter.Rate{
		Period: period,
		Limit:  limit,
	}
	// 将数据存入内存
	store := memory.NewStore()

	// 创建速率实例, 必须是真实的请求
	instance := limiter.New(store, rate, limiter.WithTrustForwardHeader(true))

	// 生成gin中间件
	return mgin.NewMiddleware(instance)
}

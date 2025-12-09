package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// Options Redis 配置选项
type Options struct {
	Mode       string           `yaml:"mode"`       // single 或 cluster
	Single     SingleConfig     `yaml:"single"`     // 单机配置
	Cluster    ClusterConfig    `yaml:"cluster"`    // 集群配置
	Pool       PoolConfig       `yaml:"pool"`       // 连接池配置
	Timeout    TimeoutConfig    `yaml:"timeout"`    // 超时配置
	Connection ConnectionConfig `yaml:"connection"` // 连接管理配置
	Reconnect  ReconnectConfig  `yaml:"reconnect"`  // 自动重连配置
}

// SingleConfig 单机模式配置
type SingleConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Password string `yaml:"password"`
	Database int    `yaml:"database"`
}

// ClusterConfig 集群模式配置
type ClusterConfig struct {
	Addrs          []string `yaml:"addrs"`
	Password       string   `yaml:"password"`
	MaxRedirects   int      `yaml:"max-redirects"`
	ReadOnly       bool     `yaml:"read-only"`
	RouteByLatency bool     `yaml:"route-by-latency"`
	RouteRandomly  bool     `yaml:"route-randomly"`
}

// PoolConfig 连接池配置
type PoolConfig struct {
	PoolSize        int `yaml:"pool-size"`
	MinIdleConns    int `yaml:"min-idle-conns"`
	MaxRetries      int `yaml:"max-retries"`
	MinRetryBackoff int `yaml:"min-retry-backoff"`
	MaxRetryBackoff int `yaml:"max-retry-backoff"`
}

// TimeoutConfig 超时配置
type TimeoutConfig struct {
	DialTimeout  int `yaml:"dial-timeout"`
	ReadTimeout  int `yaml:"read-timeout"`
	WriteTimeout int `yaml:"write-timeout"`
	PoolTimeout  int `yaml:"pool-timeout"`
}

// ConnectionConfig 连接管理配置
type ConnectionConfig struct {
	IdleCheckFrequency int `yaml:"idle-check-frequency"`
	IdleTimeout        int `yaml:"idle-timeout"`
	MaxConnAge         int `yaml:"max-conn-age"`
}

// ReconnectConfig 自动重连配置
type ReconnectConfig struct {
	Enabled           bool    `yaml:"enabled"`
	Interval          int     `yaml:"interval"`
	MaxAttempts       int     `yaml:"max-attempts"`
	BackoffMultiplier float64 `yaml:"backoff-multiplier"`
	MaxInterval       int     `yaml:"max-interval"`
}

// InitRedis 初始化 Redis 连接
func InitRedis(opt *Options) redis.UniversalClient {
	var client redis.UniversalClient

	switch opt.Mode {
	case "cluster":
		client = createClusterClient(opt)
	case "single":
		fallthrough
	default:
		client = createSingleClient(opt)
	}

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := client.Ping(ctx).Result()
	if err != nil {
		panic(fmt.Sprintf("failed to connect to redis: %v", err))
	}

	// 启用自动重连
	if opt.Reconnect.Enabled {
		go startReconnectMonitor(client, opt)
	}

	return client
}

// createSingleClient 创建单机 Redis 客户端
func createSingleClient(opt *Options) redis.UniversalClient {
	return redis.NewClient(&redis.Options{
		Addr:            fmt.Sprintf("%s:%d", opt.Single.Host, opt.Single.Port),
		Password:        opt.Single.Password,
		DB:              opt.Single.Database,
		PoolSize:        opt.Pool.PoolSize,
		MinIdleConns:    opt.Pool.MinIdleConns,
		MaxRetries:      opt.Pool.MaxRetries,
		MinRetryBackoff: time.Duration(opt.Pool.MinRetryBackoff) * time.Millisecond,
		MaxRetryBackoff: time.Duration(opt.Pool.MaxRetryBackoff) * time.Millisecond,
		DialTimeout:     time.Duration(opt.Timeout.DialTimeout) * time.Second,
		ReadTimeout:     time.Duration(opt.Timeout.ReadTimeout) * time.Second,
		WriteTimeout:    time.Duration(opt.Timeout.WriteTimeout) * time.Second,
		PoolTimeout:     time.Duration(opt.Timeout.PoolTimeout) * time.Second,
	})
}

// createClusterClient 创建集群 Redis 客户端
func createClusterClient(opt *Options) redis.UniversalClient {
	return redis.NewClusterClient(&redis.ClusterOptions{
		Addrs:           opt.Cluster.Addrs,
		Password:        opt.Cluster.Password,
		MaxRedirects:    opt.Cluster.MaxRedirects,
		ReadOnly:        opt.Cluster.ReadOnly,
		RouteByLatency:  opt.Cluster.RouteByLatency,
		RouteRandomly:   opt.Cluster.RouteRandomly,
		PoolSize:        opt.Pool.PoolSize,
		MinIdleConns:    opt.Pool.MinIdleConns,
		MaxRetries:      opt.Pool.MaxRetries,
		MinRetryBackoff: time.Duration(opt.Pool.MinRetryBackoff) * time.Millisecond,
		MaxRetryBackoff: time.Duration(opt.Pool.MaxRetryBackoff) * time.Millisecond,
		DialTimeout:     time.Duration(opt.Timeout.DialTimeout) * time.Second,
		ReadTimeout:     time.Duration(opt.Timeout.ReadTimeout) * time.Second,
		WriteTimeout:    time.Duration(opt.Timeout.WriteTimeout) * time.Second,
		PoolTimeout:     time.Duration(opt.Timeout.PoolTimeout) * time.Second,
	})
}

// startReconnectMonitor 启动自动重连监控
func startReconnectMonitor(client redis.UniversalClient, opt *Options) {
	interval := time.Duration(opt.Reconnect.Interval) * time.Second
	maxInterval := time.Duration(opt.Reconnect.MaxInterval) * time.Second
	attempts := 0
	currentInterval := interval

	for {
		time.Sleep(currentInterval)

		// 检查连接状态
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		_, err := client.Ping(ctx).Result()
		cancel()

		if err != nil {
			attempts++

			// 检查是否超过最大重连次数
			if opt.Reconnect.MaxAttempts > 0 && attempts >= opt.Reconnect.MaxAttempts {
				fmt.Printf("Redis reconnect failed after %d attempts, giving up\n", attempts)
				return
			}

			fmt.Printf("Redis connection lost, attempting reconnect (attempt %d)...\n", attempts)

			// 指数退避
			currentInterval = time.Duration(float64(currentInterval) * opt.Reconnect.BackoffMultiplier)
			if currentInterval > maxInterval {
				currentInterval = maxInterval
			}
		} else {
			// 连接正常，重置计数器和间隔
			if attempts > 0 {
				fmt.Printf("Redis connection restored after %d attempts\n", attempts)
				attempts = 0
				currentInterval = interval
			}
		}
	}
}

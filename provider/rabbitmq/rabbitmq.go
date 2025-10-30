package rabbitmq

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/streadway/amqp"
)

// Options RabbitMQ 配置选项
type Options struct {
	URL        string           `yaml:"url"`
	Connection ConnectionConfig `yaml:"connection"`
	TLS        TLSConfig        `yaml:"tls"`
	Pool       PoolConfig       `yaml:"pool"`
	Exchange   ExchangeConfig   `yaml:"exchange"`
	Queue      QueueConfig      `yaml:"queue"`
	Consumer   ConsumerConfig   `yaml:"consumer"`
	Producer   ProducerConfig   `yaml:"producer"`
	Message    MessageConfig    `yaml:"message"`
	Retry      RetryConfig      `yaml:"retry"`
	Logging    LoggingConfig    `yaml:"logging"`
	Monitoring MonitoringConfig `yaml:"monitoring"`
}

// ConnectionConfig 连接配置
type ConnectionConfig struct {
	Host                 string `yaml:"host"`
	Port                 int    `yaml:"port"`
	Username             string `yaml:"username"`
	Password             string `yaml:"password"`
	VHost                string `yaml:"vhost"`
	Timeout              int    `yaml:"timeout"`
	Heartbeat            int    `yaml:"heartbeat"`
	TLS                  bool   `yaml:"tls"`
	MaxReconnectAttempts int    `yaml:"max-reconnect-attempts"`
	ReconnectInterval    int    `yaml:"reconnect-interval"`
}

// TLSConfig TLS/SSL 配置
type TLSConfig struct {
	Enabled            bool   `yaml:"enabled"`
	CertFile           string `yaml:"cert-file"`
	KeyFile            string `yaml:"key-file"`
	CAFile             string `yaml:"ca-file"`
	InsecureSkipVerify bool   `yaml:"insecure-skip-verify"`
}

// PoolConfig 连接池配置
type PoolConfig struct {
	MaxConnections     int `yaml:"max-connections"`
	MinIdleConnections int `yaml:"min-idle-connections"`
	IdleTimeout        int `yaml:"idle-timeout"`
	MaxLifetime        int `yaml:"max-lifetime"`
}

// ExchangeConfig 交换机配置
type ExchangeConfig struct {
	DefaultName string `yaml:"default-name"`
	DefaultType string `yaml:"default-type"`
	Durable     bool   `yaml:"durable"`
	AutoDelete  bool   `yaml:"auto-delete"`
	Internal    bool   `yaml:"internal"`
	NoWait      bool   `yaml:"no-wait"`
}

// QueueConfig 队列配置
type QueueConfig struct {
	NamePrefix string                 `yaml:"name-prefix"`
	Durable    bool                   `yaml:"durable"`
	AutoDelete bool                   `yaml:"auto-delete"`
	Exclusive  bool                   `yaml:"exclusive"`
	NoWait     bool                   `yaml:"no-wait"`
	Args       map[string]interface{} `yaml:"args"`
}

// ConsumerConfig 消费者配置
type ConsumerConfig struct {
	Tag            string `yaml:"tag"`
	AutoAck        bool   `yaml:"auto-ack"`
	Exclusive      bool   `yaml:"exclusive"`
	NoWait         bool   `yaml:"no-wait"`
	PrefetchCount  int    `yaml:"prefetch-count"`
	PrefetchSize   int    `yaml:"prefetch-size"`
	GlobalPrefetch bool   `yaml:"global-prefetch"`
}

// ProducerConfig 生产者配置
type ProducerConfig struct {
	Mandatory   bool `yaml:"mandatory"`
	Immediate   bool `yaml:"immediate"`
	ConfirmMode bool `yaml:"confirm-mode"`
	Timeout     int  `yaml:"timeout"`
}

// MessageConfig 消息配置
type MessageConfig struct {
	ContentType     string `yaml:"content-type"`
	ContentEncoding string `yaml:"content-encoding"`
	DeliveryMode    uint8  `yaml:"delivery-mode"`
	Priority        uint8  `yaml:"priority"`
	Expiration      string `yaml:"expiration"`
}

// RetryConfig 重试配置
type RetryConfig struct {
	Enabled           bool    `yaml:"enabled"`
	MaxAttempts       int     `yaml:"max-attempts"`
	Interval          int     `yaml:"interval"`
	BackoffMultiplier float64 `yaml:"backoff-multiplier"`
	MaxInterval       int     `yaml:"max-interval"`
}

// LoggingConfig 日志配置
type LoggingConfig struct {
	Debug               bool `yaml:"debug"`
	LogConnectionEvents bool `yaml:"log-connection-events"`
	LogMessageEvents    bool `yaml:"log-message-events"`
	LogErrorDetails     bool `yaml:"log-error-details"`
}

// MonitoringConfig 监控配置
type MonitoringConfig struct {
	Enabled            bool `yaml:"enabled"`
	Interval           int  `yaml:"interval"`
	MonitorQueues      bool `yaml:"monitor-queues"`
	MonitorExchanges   bool `yaml:"monitor-exchanges"`
	MonitorConnections bool `yaml:"monitor-connections"`
}

// RabbitMQClient RabbitMQ 客户端包装器
type RabbitMQClient struct {
	conn     *amqp.Connection
	channels map[string]*amqp.Channel
	options  *Options
	ctx      context.Context
	cancel   context.CancelFunc
	mutex    sync.RWMutex
}

// InitRabbitMQ 初始化 RabbitMQ 客户端
func InitRabbitMQ(opt *Options) *RabbitMQClient {
	// 设置默认值
	setDefaults(opt)

	// 创建上下文
	ctx, cancel := context.WithCancel(context.Background())

	client := &RabbitMQClient{
		channels: make(map[string]*amqp.Channel),
		options:  opt,
		ctx:      ctx,
		cancel:   cancel,
	}

	return client
}

// Connect 连接到 RabbitMQ 服务器
func (r *RabbitMQClient) Connect() error {
	var err error

	// 构建连接 URL
	url := r.options.URL
	if url == "" {
		url = r.buildConnectionURL()
	}

	// 配置 TLS
	var config *amqp.Config
	if r.options.TLS.Enabled {
		tlsConfig := &tls.Config{
			InsecureSkipVerify: r.options.TLS.InsecureSkipVerify,
		}

		if r.options.TLS.CertFile != "" && r.options.TLS.KeyFile != "" {
			cert, err := tls.LoadX509KeyPair(r.options.TLS.CertFile, r.options.TLS.KeyFile)
			if err != nil {
				return fmt.Errorf("failed to load TLS certificate: %w", err)
			}
			tlsConfig.Certificates = []tls.Certificate{cert}
		}

		config = &amqp.Config{
			TLSClientConfig: tlsConfig,
			Heartbeat:       time.Duration(r.options.Connection.Heartbeat) * time.Second,
		}
	} else {
		config = &amqp.Config{
			Heartbeat: time.Duration(r.options.Connection.Heartbeat) * time.Second,
		}
	}

	// 建立连接
	r.conn, err = amqp.DialConfig(url, *config)
	if err != nil {
		return fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	if r.options.Logging.LogConnectionEvents {
		log.Println("RabbitMQ client connected")
	}

	// 监听连接关闭事件
	go r.handleConnectionClose()

	return nil
}

// Disconnect 断开连接
func (r *RabbitMQClient) Disconnect() error {
	r.cancel()

	r.mutex.Lock()
	defer r.mutex.Unlock()

	// 关闭所有通道
	for name, ch := range r.channels {
		if err := ch.Close(); err != nil {
			log.Printf("Failed to close channel %s: %v", name, err)
		}
	}
	r.channels = make(map[string]*amqp.Channel)

	// 关闭连接
	if r.conn != nil && !r.conn.IsClosed() {
		return r.conn.Close()
	}

	return nil
}

// GetChannel 获取或创建通道
func (r *RabbitMQClient) GetChannel(name string) (*amqp.Channel, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if ch, exists := r.channels[name]; exists && r.IsConnected() {
		return ch, nil
	}

	ch, err := r.conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to create channel: %w", err)
	}

	// 设置 QoS
	if err := ch.Qos(
		r.options.Consumer.PrefetchCount,
		r.options.Consumer.PrefetchSize,
		r.options.Consumer.GlobalPrefetch,
	); err != nil {
		ch.Close()
		return nil, fmt.Errorf("failed to set QoS: %w", err)
	}

	r.channels[name] = ch
	return ch, nil
}

// DeclareExchange 声明交换机
func (r *RabbitMQClient) DeclareExchange(name, kind string) error {
	ch, err := r.GetChannel("exchange")
	if err != nil {
		return err
	}

	return ch.ExchangeDeclare(
		name,
		kind,
		r.options.Exchange.Durable,
		r.options.Exchange.AutoDelete,
		r.options.Exchange.Internal,
		r.options.Exchange.NoWait,
		nil,
	)
}

// DeclareQueue 声明队列
func (r *RabbitMQClient) DeclareQueue(name string) (amqp.Queue, error) {
	ch, err := r.GetChannel("queue")
	if err != nil {
		return amqp.Queue{}, err
	}

	return ch.QueueDeclare(
		name,
		r.options.Queue.Durable,
		r.options.Queue.AutoDelete,
		r.options.Queue.Exclusive,
		r.options.Queue.NoWait,
		amqp.Table(r.options.Queue.Args),
	)
}

// BindQueue 绑定队列到交换机
func (r *RabbitMQClient) BindQueue(queueName, routingKey, exchangeName string) error {
	ch, err := r.GetChannel("bind")
	if err != nil {
		return err
	}

	return ch.QueueBind(
		queueName,
		routingKey,
		exchangeName,
		r.options.Queue.NoWait,
		nil,
	)
}

// Publish 发布消息
func (r *RabbitMQClient) Publish(exchange, routingKey string, body []byte) error {
	ch, err := r.GetChannel("publish")
	if err != nil {
		return err
	}

	if r.options.Producer.ConfirmMode {
		if err := ch.Confirm(false); err != nil {
			return fmt.Errorf("failed to put channel in confirm mode: %w", err)
		}
	}

	publishing := amqp.Publishing{
		ContentType:     r.options.Message.ContentType,
		ContentEncoding: r.options.Message.ContentEncoding,
		DeliveryMode:    r.options.Message.DeliveryMode,
		Priority:        r.options.Message.Priority,
		Timestamp:       time.Now(),
		Body:            body,
	}

	if r.options.Message.Expiration != "" {
		publishing.Expiration = r.options.Message.Expiration
	}

	err = ch.Publish(
		exchange,
		routingKey,
		r.options.Producer.Mandatory,
		r.options.Producer.Immediate,
		publishing,
	)

	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	if r.options.Producer.ConfirmMode {
		if confirmed := ch.NotifyPublish(make(chan amqp.Confirmation, 1)); confirmed != nil {
			select {
			case confirm := <-confirmed:
				if !confirm.Ack {
					return fmt.Errorf("message was not confirmed by server")
				}
			case <-time.After(time.Duration(r.options.Producer.Timeout) * time.Second):
				return fmt.Errorf("publish confirmation timeout")
			}
		}
	}

	if r.options.Logging.LogMessageEvents {
		log.Printf("Published message to exchange %s with routing key %s", exchange, routingKey)
	}

	return nil
}

// Consume 消费消息
func (r *RabbitMQClient) Consume(queueName string, handler func(amqp.Delivery)) error {
	ch, err := r.GetChannel("consume")
	if err != nil {
		return err
	}

	msgs, err := ch.Consume(
		queueName,
		r.options.Consumer.Tag,
		r.options.Consumer.AutoAck,
		r.options.Consumer.Exclusive,
		false, // no-local
		r.options.Consumer.NoWait,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to register consumer: %w", err)
	}

	go func() {
		for {
			select {
			case msg, ok := <-msgs:
				if !ok {
					return
				}

				if r.options.Logging.LogMessageEvents {
					log.Printf("Received message from queue %s", queueName)
				}

				handler(msg)
			case <-r.ctx.Done():
				return
			}
		}
	}()

	return nil
}

// IsConnected 检查连接状态
func (r *RabbitMQClient) IsConnected() bool {
	return r.conn != nil && !r.conn.IsClosed()
}

// GetConnection 获取原始连接
func (r *RabbitMQClient) GetConnection() *amqp.Connection {
	return r.conn
}

// buildConnectionURL 构建连接 URL
func (r *RabbitMQClient) buildConnectionURL() string {
	scheme := "amqp"
	if r.options.Connection.TLS {
		scheme = "amqps"
	}

	return fmt.Sprintf("%s://%s:%s@%s:%d%s",
		scheme,
		r.options.Connection.Username,
		r.options.Connection.Password,
		r.options.Connection.Host,
		r.options.Connection.Port,
		r.options.Connection.VHost,
	)
}

// handleConnectionClose 处理连接关闭事件
func (r *RabbitMQClient) handleConnectionClose() {
	closeCh := make(chan *amqp.Error)
	r.conn.NotifyClose(closeCh)

	select {
	case err := <-closeCh:
		if err != nil {
			if r.options.Logging.LogConnectionEvents {
				log.Printf("RabbitMQ connection closed: %v", err)
			}

			// 尝试重连
			if r.options.Connection.MaxReconnectAttempts > 0 {
				r.reconnect()
			}
		}
	case <-r.ctx.Done():
		return
	}
}

// reconnect 重连逻辑
func (r *RabbitMQClient) reconnect() {
	for attempt := 1; attempt <= r.options.Connection.MaxReconnectAttempts; attempt++ {
		if r.options.Logging.LogConnectionEvents {
			log.Printf("Attempting to reconnect to RabbitMQ (attempt %d/%d)", attempt, r.options.Connection.MaxReconnectAttempts)
		}

		time.Sleep(time.Duration(r.options.Connection.ReconnectInterval) * time.Second)

		if err := r.Connect(); err != nil {
			if r.options.Logging.LogErrorDetails {
				log.Printf("Reconnection attempt %d failed: %v", attempt, err)
			}
			continue
		}

		if r.options.Logging.LogConnectionEvents {
			log.Println("Successfully reconnected to RabbitMQ")
		}
		return
	}

	if r.options.Logging.LogErrorDetails {
		log.Printf("Failed to reconnect to RabbitMQ after %d attempts", r.options.Connection.MaxReconnectAttempts)
	}
}

// setDefaults 设置默认值
func setDefaults(opt *Options) {
	if opt.Connection.Host == "" {
		opt.Connection.Host = "localhost"
	}
	if opt.Connection.Port == 0 {
		opt.Connection.Port = 5672
	}
	if opt.Connection.Username == "" {
		opt.Connection.Username = "guest"
	}
	if opt.Connection.Password == "" {
		opt.Connection.Password = "guest"
	}
	if opt.Connection.VHost == "" {
		opt.Connection.VHost = "/"
	}
	if opt.Connection.Timeout == 0 {
		opt.Connection.Timeout = 30
	}
	if opt.Connection.Heartbeat == 0 {
		opt.Connection.Heartbeat = 60
	}
	if opt.Connection.MaxReconnectAttempts == 0 {
		opt.Connection.MaxReconnectAttempts = 5
	}
	if opt.Connection.ReconnectInterval == 0 {
		opt.Connection.ReconnectInterval = 5
	}
	if opt.Pool.MaxConnections == 0 {
		opt.Pool.MaxConnections = 10
	}
	if opt.Pool.MinIdleConnections == 0 {
		opt.Pool.MinIdleConnections = 2
	}
	if opt.Pool.IdleTimeout == 0 {
		opt.Pool.IdleTimeout = 300
	}
	if opt.Pool.MaxLifetime == 0 {
		opt.Pool.MaxLifetime = 3600
	}
	if opt.Exchange.DefaultType == "" {
		opt.Exchange.DefaultType = "direct"
	}
	if opt.Consumer.PrefetchCount == 0 {
		opt.Consumer.PrefetchCount = 1
	}
	if opt.Producer.Timeout == 0 {
		opt.Producer.Timeout = 30
	}
	if opt.Message.ContentType == "" {
		opt.Message.ContentType = "application/json"
	}
	if opt.Message.DeliveryMode == 0 {
		opt.Message.DeliveryMode = 2 // 持久化
	}
	if opt.Retry.MaxAttempts == 0 {
		opt.Retry.MaxAttempts = 3
	}
	if opt.Retry.Interval == 0 {
		opt.Retry.Interval = 1
	}
	if opt.Retry.BackoffMultiplier == 0 {
		opt.Retry.BackoffMultiplier = 2.0
	}
	if opt.Retry.MaxInterval == 0 {
		opt.Retry.MaxInterval = 60
	}
	if opt.Monitoring.Interval == 0 {
		opt.Monitoring.Interval = 30
	}
}

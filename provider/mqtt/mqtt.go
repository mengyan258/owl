package mqtt

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// Options MQTT 配置选项
type Options struct {
	Broker       string             `json:"broker"`
	Client       ClientConfig       `json:"client"`
	TLS          TLSConfig          `json:"tls"`
	Message      MessageConfig      `json:"message"`
	Subscription SubscriptionConfig `json:"subscription"`
	Publish      PublishConfig      `json:"publish"`
	Logging      LoggingConfig      `json:"logging"`
	Buffer       BufferConfig       `json:"buffer"`
	Performance  PerformanceConfig  `json:"performance"`
}

// ClientConfig 客户端配置
type ClientConfig struct {
	ID                   string `json:"id"`
	Username             string `json:"username"`
	Password             string `json:"password"`
	CleanSession         bool   `json:"clean-session"`
	KeepAlive            int    `json:"keep-alive"`
	ConnectTimeout       int    `json:"connect-timeout"`
	AutoReconnect        bool   `json:"auto-reconnect"`
	ReconnectInterval    int    `json:"reconnect-interval"`
	MaxReconnectAttempts int    `json:"max-reconnect-attempts"`
}

// TLSConfig TLS/SSL 配置
type TLSConfig struct {
	Enabled            bool   `json:"enabled"`
	CertFile           string `json:"cert-file"`
	KeyFile            string `json:"key-file"`
	CAFile             string `json:"ca-file"`
	InsecureSkipVerify bool   `json:"insecure-skip-verify"`
}

// MessageConfig 消息配置
type MessageConfig struct {
	DefaultQoS byte `json:"default-qos"`
	Retain     bool `json:"retain"`
	Timeout    int  `json:"timeout"`
}

// SubscriptionConfig 订阅配置
type SubscriptionConfig struct {
	DefaultQoS byte `json:"default-qos"`
	Timeout    int  `json:"timeout"`
}

// PublishConfig 发布配置
type PublishConfig struct {
	DefaultQoS byte `json:"default-qos"`
	Timeout    int  `json:"timeout"`
	WaitForAck bool `json:"wait-for-ack"`
}

// LoggingConfig 日志配置
type LoggingConfig struct {
	Debug               bool `json:"debug"`
	LogConnectionEvents bool `json:"log-connection-events"`
	LogMessageEvents    bool `json:"log-message-events"`
}

// BufferConfig 缓冲区配置
type BufferConfig struct {
	SendBufferSize    int `json:"send-buffer-size"`
	ReceiveBufferSize int `json:"receive-buffer-size"`
	MessageQueueSize  int `json:"message-queue-size"`
}

// PerformanceConfig 性能配置
type PerformanceConfig struct {
	MaxConcurrentConnections int `json:"max-concurrent-connections"`
	MessageHandlerCount      int `json:"message-handler-count"`
	BatchSize                int `json:"batch-size"`
}

// MQTTClient MQTT 客户端包装器
type MQTTClient struct {
	client  mqtt.Client
	options *Options
	ctx     context.Context
	cancel  context.CancelFunc
}

// InitMQTT 初始化 MQTT 客户端
func InitMQTT(opt *Options) *MQTTClient {
	// 设置默认值
	setDefaults(opt)

	// 创建客户端选项
	opts := mqtt.NewClientOptions()
	opts.AddBroker(opt.Broker)

	// 设置客户端 ID
	if opt.Client.ID != "" {
		opts.SetClientID(opt.Client.ID)
	} else {
		opts.SetClientID(fmt.Sprintf("mqtt-client-%d", time.Now().UnixNano()))
	}

	// 设置认证
	if opt.Client.Username != "" {
		opts.SetUsername(opt.Client.Username)
	}
	if opt.Client.Password != "" {
		opts.SetPassword(opt.Client.Password)
	}

	// 设置连接参数
	opts.SetCleanSession(opt.Client.CleanSession)
	opts.SetKeepAlive(time.Duration(opt.Client.KeepAlive) * time.Second)
	opts.SetConnectTimeout(time.Duration(opt.Client.ConnectTimeout) * time.Second)
	opts.SetAutoReconnect(opt.Client.AutoReconnect)

	if opt.Client.AutoReconnect {
		opts.SetMaxReconnectInterval(time.Duration(opt.Client.ReconnectInterval) * time.Second)
	}

	// 设置 TLS
	if opt.TLS.Enabled {
		tlsConfig := &tls.Config{
			InsecureSkipVerify: opt.TLS.InsecureSkipVerify,
		}

		if opt.TLS.CertFile != "" && opt.TLS.KeyFile != "" {
			cert, err := tls.LoadX509KeyPair(opt.TLS.CertFile, opt.TLS.KeyFile)
			if err != nil {
				log.Printf("Failed to load TLS certificate: %v", err)
			} else {
				tlsConfig.Certificates = []tls.Certificate{cert}
			}
		}

		opts.SetTLSConfig(tlsConfig)
	}

	// 设置日志
	if opt.Logging.Debug {
		mqtt.DEBUG = log.New(log.Writer(), "[MQTT-DEBUG] ", log.LstdFlags)
		mqtt.WARN = log.New(log.Writer(), "[MQTT-WARN] ", log.LstdFlags)
		mqtt.CRITICAL = log.New(log.Writer(), "[MQTT-CRITICAL] ", log.LstdFlags)
		mqtt.ERROR = log.New(log.Writer(), "[MQTT-ERROR] ", log.LstdFlags)
	}

	// 设置连接事件处理器
	if opt.Logging.LogConnectionEvents {
		opts.SetOnConnectHandler(func(client mqtt.Client) {
			log.Println("MQTT client connected")
		})

		opts.SetConnectionLostHandler(func(client mqtt.Client, err error) {
			log.Printf("MQTT connection lost: %v", err)
		})

		opts.SetReconnectingHandler(func(client mqtt.Client, opts *mqtt.ClientOptions) {
			log.Println("MQTT client reconnecting...")
		})
	}

	// 创建客户端
	client := mqtt.NewClient(opts)

	// 创建上下文
	ctx, cancel := context.WithCancel(context.Background())

	return &MQTTClient{
		client:  client,
		options: opt,
		ctx:     ctx,
		cancel:  cancel,
	}
}

// Connect 连接到 MQTT 服务器
func (m *MQTTClient) Connect() error {
	token := m.client.Connect()
	if token.Wait() && token.Error() != nil {
		return fmt.Errorf("failed to connect to MQTT broker: %w", token.Error())
	}
	return nil
}

// Disconnect 断开连接
func (m *MQTTClient) Disconnect() {
	m.cancel()
	m.client.Disconnect(250)
}

// Publish 发布消息
func (m *MQTTClient) Publish(topic string, payload interface{}) error {
	return m.PublishWithQoS(topic, payload, m.options.Publish.DefaultQoS)
}

// PublishWithQoS 使用指定 QoS 发布消息
func (m *MQTTClient) PublishWithQoS(topic string, payload interface{}, qos byte) error {
	token := m.client.Publish(topic, qos, m.options.Message.Retain, payload)

	if m.options.Publish.WaitForAck {
		if token.WaitTimeout(time.Duration(m.options.Publish.Timeout)*time.Second) && token.Error() != nil {
			return fmt.Errorf("failed to publish message: %w", token.Error())
		}
	}

	return nil
}

// Subscribe 订阅主题
func (m *MQTTClient) Subscribe(topic string, callback mqtt.MessageHandler) error {
	return m.SubscribeWithQoS(topic, m.options.Subscription.DefaultQoS, callback)
}

// SubscribeWithQoS 使用指定 QoS 订阅主题
func (m *MQTTClient) SubscribeWithQoS(topic string, qos byte, callback mqtt.MessageHandler) error {
	// 如果启用了消息事件日志，包装回调函数
	if m.options.Logging.LogMessageEvents {
		originalCallback := callback
		callback = func(client mqtt.Client, msg mqtt.Message) {
			log.Printf("Received message on topic %s: %s", msg.Topic(), string(msg.Payload()))
			originalCallback(client, msg)
		}
	}

	token := m.client.Subscribe(topic, qos, callback)
	if token.WaitTimeout(time.Duration(m.options.Subscription.Timeout)*time.Second) && token.Error() != nil {
		return fmt.Errorf("failed to subscribe to topic %s: %w", topic, token.Error())
	}

	return nil
}

// Unsubscribe 取消订阅
func (m *MQTTClient) Unsubscribe(topics ...string) error {
	token := m.client.Unsubscribe(topics...)
	if token.WaitTimeout(time.Duration(m.options.Subscription.Timeout)*time.Second) && token.Error() != nil {
		return fmt.Errorf("failed to unsubscribe from topics: %w", token.Error())
	}

	return nil
}

// IsConnected 检查连接状态
func (m *MQTTClient) IsConnected() bool {
	return m.client.IsConnected()
}

// GetClient 获取原始客户端
func (m *MQTTClient) GetClient() mqtt.Client {
	return m.client
}

// setDefaults 设置默认值
func setDefaults(opt *Options) {
	if opt.Client.KeepAlive == 0 {
		opt.Client.KeepAlive = 60
	}
	if opt.Client.ConnectTimeout == 0 {
		opt.Client.ConnectTimeout = 30
	}
	if opt.Client.ReconnectInterval == 0 {
		opt.Client.ReconnectInterval = 5
	}
	if opt.Message.Timeout == 0 {
		opt.Message.Timeout = 30
	}
	if opt.Subscription.Timeout == 0 {
		opt.Subscription.Timeout = 10
	}
	if opt.Publish.Timeout == 0 {
		opt.Publish.Timeout = 10
	}
	if opt.Buffer.SendBufferSize == 0 {
		opt.Buffer.SendBufferSize = 1024
	}
	if opt.Buffer.ReceiveBufferSize == 0 {
		opt.Buffer.ReceiveBufferSize = 1024
	}
	if opt.Buffer.MessageQueueSize == 0 {
		opt.Buffer.MessageQueueSize = 100
	}
	if opt.Performance.MaxConcurrentConnections == 0 {
		opt.Performance.MaxConcurrentConnections = 100
	}
	if opt.Performance.MessageHandlerCount == 0 {
		opt.Performance.MessageHandlerCount = 10
	}
	if opt.Performance.BatchSize == 0 {
		opt.Performance.BatchSize = 50
	}
}

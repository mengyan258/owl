package foundation

type ServiceProvider interface {
	Register()                       // 注册服务
	Boot()                           // 启动服务，视图，路由等等都可以在这个方法中初始化
	GenerateConf() map[string]string // 生成配置
}

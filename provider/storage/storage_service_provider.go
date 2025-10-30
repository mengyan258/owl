package storage

import (
	"bit-labs.cn/owl"
	"bit-labs.cn/owl/bootstrap/conf"
	"bit-labs.cn/owl/contract/foundation"
	"bit-labs.cn/owl/provider/storage/impl"
	_ "embed"
)

var _ foundation.ServiceProvider = (*StorageServiceProvider)(nil)

// StorageServiceProvider 存储服务提供者
type StorageServiceProvider struct {
	app foundation.Application
}

func NewStorageServiceProvider(app foundation.Application) *StorageServiceProvider {
	return &StorageServiceProvider{
		app: app,
	}
}

// Register 注册服务
func (s *StorageServiceProvider) Register() {
	s.app.Register(func(c *conf.Configure) *StorageManager {
		var opt Options
		err := c.GetConfig("storage", &opt)
		owl.PanicIf(err)

		// 设置默认值
		setDefaults(&opt)

		// 初始化存储管理器
		manager := NewStorageManager()

		// 初始化本地存储
		if opt.Local.Root != "" {
			localStorage := impl.NewLocalStorage(&opt.Local)
			manager.AddDriver("local", localStorage)
		}

		// 初始化 S3 存储
		if opt.S3.AccessKeyID != "" {
			if s3Storage, err := impl.NewS3Storage(&opt.S3); err == nil {
				manager.AddDriver("s3", s3Storage)
			}
		}

		// 初始化 MinIO 存储
		if opt.MinIO.AccessKeyID != "" {
			if minioStorage, err := impl.NewMinIOStorage(&opt.MinIO); err == nil {
				manager.AddDriver("minio", minioStorage)
			}
		}

		// 初始化腾讯云 COS 存储
		if opt.COS.SecretID != "" {
			if cosStorage, err := impl.NewCOSStorage(&opt.COS); err == nil {
				manager.AddDriver("cos", cosStorage)
			}
		}

		// 初始化七牛云存储
		if opt.Qiniu.AccessKey != "" {
			if qiniuStorage, err := impl.NewQiniuStorage(&opt.Qiniu); err == nil {
				manager.AddDriver("qiniu", qiniuStorage)
			}
		}

		// 设置默认驱动
		if err := manager.SetDefaultDriver(opt.Default); err != nil {
			owl.PanicIf(err)
		}

		return manager
	})
}

// Boot 启动服务
func (s *StorageServiceProvider) Boot() {

}

//go:embed storage.yaml
var storageYaml string

// GenerateConf 生成配置文件
func (s *StorageServiceProvider) GenerateConf() map[string]string {
	return map[string]string{
		"storage.yaml": storageYaml,
	}
}

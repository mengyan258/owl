package ocr

import (
	_ "embed"

	"bit-labs.cn/owl/contract/foundation"
	"bit-labs.cn/owl/provider/conf"
)

var _ foundation.ServiceProvider = (*OcrServiceProvider)(nil)

type OcrServiceProvider struct{ app foundation.Application }

func (i *OcrServiceProvider) Description() string {
	return "OCR 服务聚合与多厂商客户端，提供图像识别功能"
}

type Options struct {
	Provider string      `json:"provider"`
	Baidu    BaiduConf   `json:"baidu"`
	Aliyun   AliyunConf  `json:"aliyun"`
	Tencent  TencentConf `json:"tencent"`
}

func (i *OcrServiceProvider) Register() {
	i.app.Register(func(c *conf.Configure) *Service {
		var opt Options
		if err := c.GetConfig("ocr", &opt); err != nil {
			panic(err)
		}
		var client Client
		switch opt.Provider {
		case "baidu":
			client = NewBaiduClient(opt.Baidu)
		case "aliyun":
			client = NewAliyunClient(opt.Aliyun)
		case "tencent":
			client = NewTencentClient(opt.Tencent)
		default:
			client = NewBaiduClient(opt.Baidu)
		}
		return NewService(client)
	})
}

func (i *OcrServiceProvider) Boot() {}

//go:embed ocr.yaml
var ocrYaml string

func (i *OcrServiceProvider) Conf() map[string]string {
	return map[string]string{"ocr.yaml": ocrYaml}
}

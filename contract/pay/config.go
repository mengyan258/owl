package pay

type AlipayConfig struct {
	AppID               string `json:"app_id"`
	PrivateKeyPath      string `json:"private_key_path"`
	AlipayPublicKeyPath string `json:"alipay_public_key_path"`
	NotifyURL           string `json:"notify_url"`
	ReturnURL           string `json:"return_url"`
	Sandbox             bool   `json:"sandbox"`
}

type WechatConfig struct {
	MchID            string `json:"mch_id"`
	AppID            string `json:"app_id"`
	ApiV3Key         string `json:"api_v3_key"`
	SerialNo         string `json:"serial_no"`
	PrivateKeyPath   string `json:"private_key_path"`
	PlatformCertPath string `json:"platform_cert_path"`
	NotifyURL        string `json:"notify_url"`
}

type CardConfig struct {
	Gateway        string `json:"gateway"`
	MerchantID     string `json:"merchant_id"`
	PrivateKeyPath string `json:"private_key_path"`
	PublicKeyPath  string `json:"public_key_path"`
	NotifyURL      string `json:"notify_url"`
	ThreeDSEnabled bool   `json:"threeds_enabled"`
}

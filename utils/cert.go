package utils

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"github.com/denisbrodbeck/machineid"
	jsoniter "github.com/json-iterator/go"
	"os"
	"time"
)

type LicensePayload struct {
	Version   string    `json:"version"`
	LicenseID string    `json:"licenseID"`
	Customer  string    `json:"customer"`
	Product   string    `json:"product"`
	NotBefore time.Time `json:"notBefore"`
	NotAfter  time.Time `json:"notAfter"`
	Features  []string  `json:"features"`
}

type SignedLicense struct {
	Payload   LicensePayload `json:"payload"`
	Signature string         `json:"signature"` // base64
}

func GenLicense(privateKeyStr string, payload LicensePayload) ([]byte, error) {
	payload.LicenseID, _ = machineid.ID()
	// 序列化 payload
	payloadBytes, _ := json.Marshal(payload)

	// 解析私钥
	block, _ := pem.Decode([]byte(privateKeyStr))
	if block == nil || block.Type != "PRIVATE KEY" {
		return nil, errors.New("私钥解析失败")
	}

	privateKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	// 使用 RSA 私钥签名
	hash := sha256.Sum256(payloadBytes)
	signature, _ := rsa.SignPKCS1v15(rand.Reader, privateKey.(*rsa.PrivateKey), crypto.SHA256, hash[:])

	signed := SignedLicense{
		Payload:   payload,
		Signature: base64.StdEncoding.EncodeToString(signature),
	}

	lisense, err := jsoniter.Marshal(signed)
	//os.WriteFile("license.lic", lisense, 0644)

	return lisense, err

}

func VerifyLicense(publicKeyStr string, licenseFile string) (bool, *LicensePayload) {
	// 加载公钥
	block, _ := pem.Decode([]byte(publicKeyStr))
	pubKey, _ := x509.ParsePKIXPublicKey(block.Bytes)
	rsaPub := pubKey.(*rsa.PublicKey)

	// 读取授权文件
	data, err := os.ReadFile(licenseFile)
	if err != nil {
		return false, nil
	}
	var lic SignedLicense
	json.Unmarshal(data, &lic)

	// 验证签名
	payloadBytes, _ := json.Marshal(lic.Payload)
	sigBytes, _ := base64.StdEncoding.DecodeString(lic.Signature)
	hash := sha256.Sum256(payloadBytes)
	if err := rsa.VerifyPKCS1v15(rsaPub, crypto.SHA256, hash[:], sigBytes); err != nil {
		fmt.Println("授权文件签名无效")
		return false, nil
	}

	// 检查有效期
	now := time.Now()
	if now.Before(lic.Payload.NotBefore) || now.After(lic.Payload.NotAfter) {

		fmt.Println("授权已过期或尚未生效")
		return false, nil
	}
	fmt.Println("✅ 授权通过！欢迎使用", lic.Payload.Product)
	return true, &lic.Payload
}

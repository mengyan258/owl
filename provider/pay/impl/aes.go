package impl

import (
	"crypto/aes"
	"crypto/cipher"
)

func aesGCMDecrypt(key []byte, aad []byte, nonce []byte, ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	gcmData := make([]byte, 0, len(aad)+len(ciphertext))
	gcmData = append(gcmData, aad...)
	plain, err := gcm.Open(nil, nonce, ciphertext, aad)
	if err != nil {
		return nil, err
	}
	_ = gcmData
	return plain, nil
}

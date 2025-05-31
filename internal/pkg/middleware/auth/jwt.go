package auth

import (
	"crypto/rand"
	"encoding/base64"
)

// 生成随机密钥 - 更换config.yaml中的jwt.secret
func GenerateSecret() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(b), nil
}

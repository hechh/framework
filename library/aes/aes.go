package aes

import (
	"crypto/aes"
	"encoding/base64"
	"errors"
	"fmt"

	"github.com/hechh/framework/library/safe"
)

// AesEncrypto AES-ECB 加密 + PKCS7 填充，输出 Base64 字符串。
// secretKey 长度须为 16/24/32 字节（AES-128/192/256）。
func AesEncrypto(body, secretKey []byte) (string, error) {
	block, err := aes.NewCipher(secretKey)
	if err != nil {
		return "", fmt.Errorf("aes: 创建加密块失败: %w", err)
	}

	padded := pkcs7Pad(body, aes.BlockSize)
	ciphertext := make([]byte, len(padded))

	for i := 0; i < len(padded); i += aes.BlockSize {
		block.Encrypt(ciphertext[i:i+aes.BlockSize], padded[i:i+aes.BlockSize])
	}

	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// AesDecrypto 解密 AES-ECB + PKCS7 密文。
// 自动识别输入：Base64 字符串 或 原始密文字节。
func AesDecrypto(body []byte, secretKey []byte) ([]byte, error) {
	data := body
	if decoded, err := base64.StdEncoding.DecodeString(safe.BytesToString(body)); err == nil {
		data = decoded
	} else if len(body)%aes.BlockSize != 0 {
		return nil, fmt.Errorf("aes: 密文格式无效: %w", err)
	}

	if len(data)%aes.BlockSize != 0 {
		return nil, errors.New("密文长度不是块大小的整数倍")
	}

	block, err := aes.NewCipher(secretKey)
	if err != nil {
		return nil, fmt.Errorf("aes: 创建解密块失败: %w", err)
	}

	plaintext := make([]byte, len(data))
	for i := 0; i < len(data); i += aes.BlockSize {
		block.Decrypt(plaintext[i:i+aes.BlockSize], data[i:i+aes.BlockSize])
	}

	return pkcs7Unpad(plaintext)
}

// ====================== PKCS7 填充 ======================

func pkcs7Pad(data []byte, blockSize int) []byte {
	padding := blockSize - (len(data) % blockSize)
	padtext := make([]byte, padding)
	for i := range padtext {
		padtext[i] = byte(padding)
	}
	return append(data, padtext...)
}

func pkcs7Unpad(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return nil, errors.New("pkcs7: 空数据")
	}
	padding := int(data[len(data)-1])
	if padding == 0 || padding > aes.BlockSize || padding > len(data) {
		return nil, fmt.Errorf("pkcs7: 无效填充字节 %d", padding)
	}
	for i := len(data) - padding; i < len(data); i++ {
		if data[i] != byte(padding) {
			return nil, errors.New("pkcs7: 填充字节不一致")
		}
	}
	return data[:len(data)-padding], nil
}

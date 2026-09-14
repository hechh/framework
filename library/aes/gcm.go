package aes

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"

	"github.com/hechh/framework/library/safe"
)

// GCM_NONCE_SIZE GCM nonce 长度（12 字节，NIST SP 800-38D 对各场景的推荐值）
const GCM_NONCE_SIZE = 12

// AesGcmEncrypto AES-GCM（AEAD）加密，返回 Base64(nonce || 密文 || tag)。
//
// 与 AesEncrypto（AES-ECB）的区别：
//   - 每次调用随机生成 nonce，相同明文产生不同密文（ECB 是确定性的，可做模式分析）；
//   - 密文自带 tag，解密时可以校验完整性：被篡改/拼接/重放的密文会解密失败，
//     而 ECB 无法发现这类修改。
//
// secretKey 长度须为 16/24/32 字节（AES-128/192/256）。
func AesGcmEncrypto(body, secretKey []byte) (string, error) {
	gcm, err := newGcm(secretKey)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, GCM_NONCE_SIZE)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("aes: 生成 GCM nonce 失败: %w", err)
	}

	// Seal 输出 = 密文 || tag；dst 传 nonce 使其前置，得到 nonce || 密文 || tag
	sealed := gcm.Seal(nonce, nonce, body, nil)
	return base64.StdEncoding.EncodeToString(sealed), nil
}

// AesGcmDecrypto 解密 AesGcmEncrypto 的密文并校验完整性。
// 自动识别输入：Base64 字符串 或 原始密文字节（与 AesDecrypto 保持一致）。
// tag 校验失败（密文被篡改、密钥不匹配）一律返回错误。
func AesGcmDecrypto(body []byte, secretKey []byte) ([]byte, error) {
	data := body
	if decoded, err := base64.StdEncoding.DecodeString(safe.BytesToString(body)); err == nil {
		data = decoded
	}

	gcm, err := newGcm(secretKey)
	if err != nil {
		return nil, err
	}
	if len(data) < GCM_NONCE_SIZE+gcm.Overhead() {
		return nil, errors.New("aes: GCM 密文长度不足（至少需要 nonce + tag）")
	}

	plain, err := gcm.Open(nil, data[:GCM_NONCE_SIZE], data[GCM_NONCE_SIZE:], nil)
	if err != nil {
		return nil, fmt.Errorf("aes: GCM 解密失败（密文被篡改或密钥不匹配）: %w", err)
	}
	return plain, nil
}

// newGcm 按密钥长度创建 GCM 实例
func newGcm(secretKey []byte) (cipher.AEAD, error) {
	block, err := aes.NewCipher(secretKey)
	if err != nil {
		return nil, fmt.Errorf("aes: 创建加密块失败: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("aes: 创建 GCM 失败: %w", err)
	}
	return gcm, nil
}

package aes

// 加解密协议版本协商（双栈迁移用）。
//
// 迁移期约定：
//   - 客户端在请求头 CRYPTO_VERSION_HEADER 中声明 CRYPTO_VERSION_GCM 时，
//     请求体按 AES-GCM 解析、应答体也按 AES-GCM 加密；
//   - 不带该头（或带其它值）时走旧的 AES-ECB 协议，旧客户端/旧前端无需改动即可继续工作；
//   - 应答版本必须跟随请求版本，否则旧对端解不开应答。
//
// 客户端覆盖率到位后，删掉 EncryptVersioned/DecryptVersioned 里的 ECB 分支即可下线旧协议。
const (
	// CRYPTO_VERSION_HEADER 加解密协议版本协商请求头
	CRYPTO_VERSION_HEADER = "X-Crypto"
	// CRYPTO_VERSION_GCM 支持 AES-GCM 的客户端在 CRYPTO_VERSION_HEADER 中声明的版本值
	CRYPTO_VERSION_GCM = "gcm1"
)

// EncryptVersioned 按协议版本加密：gcm 为 true 用 AES-GCM（AEAD，带完整性校验），否则用旧 AES-ECB。
func EncryptVersioned(plain, secretKey []byte, gcm bool) (string, error) {
	if gcm {
		return AesGcmEncrypto(plain, secretKey)
	}
	return AesEncrypto(plain, secretKey)
}

// DecryptVersioned 按协议版本解密：gcm 为 true 用 AES-GCM（校验 tag，被篡改即报错），否则用旧 AES-ECB。
func DecryptVersioned(body, secretKey []byte, gcm bool) ([]byte, error) {
	if gcm {
		return AesGcmDecrypto(body, secretKey)
	}
	return AesDecrypto(body, secretKey)
}

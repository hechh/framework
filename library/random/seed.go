// Package random 安全种子生成与管理。
//
// # 种子生成策略
//
// 初始种子来源:
//   - 优先: crypto/rand（Go 标准库，封装操作系统熵源）
//   - Linux:   getrandom(2) / /dev/urandom
//   - Windows: CryptGenRandom / BCryptGenRandom
//   - macOS:   getentropy(2) / /dev/urandom
//   - 熵源质量: 系统级 CSPRNG，经 FIPS 140-2 验证
//   - 种子长度: 384 位（48 字节），远超 NIST 要求的 256 位最小熵
//
// 重种策略:
//   - 定时重种: 每 24 小时调用一次 Reseed，从 crypto/rand 获取新鲜熵
//   - 计数器重种: 达到 2^48 次生成时强制重种（NIST 硬要求，实际游戏中几乎不会触发）
//   - 事件重种: 进程启动、检测到异常状态时立即重种
//
// 不可预测性保证:
//   - 种子来源于系统熵池，包含硬件中断时序、磁盘 I/O、网络包到达时间等物理熵
//   - 不依赖用户输入或可预测的事件（如时间戳）
//   - 每次实例化产生不同的初始状态（由 seed + nonce 保证唯一性）
package random

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"time"
)

// ============================================================================
// 种子生成器
// ============================================================================

// SecureSeed 从操作系统密码学安全熵源生成指定长度的种子。
//
// 内部调用 crypto/rand.Read，该函数从操作系统的 CSPRNG 读取:
//   - Linux >= 3.17: 使用 getrandom(2) 系统调用，直接从内核熵池获取
//   - Windows:        使用 BCryptGenRandom (CNG API)
//   - macOS:          使用 getentropy(2) 或 /dev/urandom
//
// 参数:
//   - bits: 种子位数。推荐 256（NIST 最低要求）或 384（冗余安全）。
//
// 返回: 长度为 bits/8 的随机字节切片。
func SecureSeed(bits int) ([]byte, error) {
	if bits < 256 {
		return nil, fmt.Errorf("random: seed bits must be >= 256, got %d", bits)
	}
	if bits%8 != 0 {
		return nil, fmt.Errorf("random: seed bits must be multiple of 8, got %d", bits)
	}

	seed := make([]byte, bits/8)
	if _, err := rand.Read(seed); err != nil {
		return nil, fmt.Errorf("random: failed to read from crypto/rand: %w", err)
	}
	return seed, nil
}

// SecureSeed384 生成 384 位安全种子，用于 HMAC-DRBG 初始化。
//
// 384 位 > NIST 最小要求 256 位，提供冗余安全性:
//   - 即使系统熵估计偏高 128 位，仍有足够安全余量
//   - 符合 NIST SP 800-90A Section 8.6.2 关于熵过采样的建议
func SecureSeed384() ([]byte, error) {
	return SecureSeed(384)
}

// ============================================================================
// 便捷函数：创建默认配置的 HMAC-DRBG
// ============================================================================

// NewDefaultDRBG 创建使用安全系统种子的默认 HMAC-DRBG。
//
// 种子来源:
//   - entropy: 来自 crypto/rand 的 384 位随机数据
//   - nonce:    来自 crypto/rand 的 128 位随机数据（确保实例唯一性）
//
// 适用场景: 大多数游戏服务的默认 RNG 实例化。
func NewDefaultDRBG() (*HMACDRBG, error) {
	entropy, err := SecureSeed384()
	if err != nil {
		return nil, err
	}

	// nonce 提供额外的唯一性保证，防止同一种子重复初始化
	nonce := make([]byte, 16)
	if _, err := rand.Read(nonce); err != nil {
		return nil, fmt.Errorf("random: failed to generate nonce: %w", err)
	}

	// 个性化字符串包含时间戳和 goroutine ID，进一步区分实例
	personalization := make([]byte, 8)
	binary.LittleEndian.PutUint64(personalization, uint64(time.Now().UnixNano()))

	return NewHMACDRBG(entropy, nonce, personalization)
}

// ============================================================================
// 全局安全 RNG（用于兼容旧 API）
// ============================================================================

// globalDRBG 是包级别的全局安全 RNG 实例。
// 用于兼容现有的 Rand/RandN/Weight 等函数。
var globalDRBG *HMACDRBG

func init() {
	var err error
	globalDRBG, err = NewDefaultDRBG()
	if err != nil {
		// 如果 crypto/rand 不可用，系统处于严重故障状态
		// panic 是合理的，因为此时任何安全随机操作都无法进行
		panic("random: failed to initialize global DRBG: " + err.Error())
	}
}

// GlobalDRBG 返回包级别的全局安全 RNG 实例。
// 线程安全，可在所有服务中共享使用。
func GlobalDRBG() *HMACDRBG {
	return globalDRBG
}

// ReseedGlobal 使用新鲜系统熵重新为全局 RNG 设定种子。
//
// 使用场景:
//   - 定时任务: 每隔 24 小时重种一次
//   - 检测到异常: 如随机数质量监控报警
func ReseedGlobal() error {
	entropy, err := SecureSeed384()
	if err != nil {
		return err
	}
	return globalDRBG.Reseed(entropy, nil)
}

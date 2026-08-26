// Package random 随机数到游戏结果的映射、洗牌与抽样。
//
// # 映射算法概述
//
// 从随机数到游戏结果的转换必须保证无偏差（unbiased），
// 即每个可能结果出现的概率完全相等。
//
// 本包提供三种核心映射:
//   - 整数范围映射：使用拒绝采样消除取模偏差
//   - 洗牌：Fisher-Yates 算法，每种排列等概率
//   - 不放回抽样：等价于 Bingo"抽球"逻辑
//
// # 拒绝采样证明（以 Intn 为例）
//
// 问题：将 [0, 2^64-1] 均匀分布的随机数映射到 [0, n-1]。
//
// 取模法 `r % n` 存在偏差：因为 2^64 通常不能被 n 整除，
// 较大的余数对应的结果会略微更频繁。
//
// 拒绝采样法:
//  1. 令 limit = 2^64 - (2^64 mod n) = 最大不产生偏差的随机值
//  2. 生成 r ∈ [0, 2^64-1]
//  3. 如果 r >= limit，拒绝并重新生成
//  4. 返回 r mod n
//
// 此时对于每个 k ∈ [0, n-1]，
// P(result=k) = (floor(2^64/n) / 2^64) = 1/n
// 因此完全无偏差。
package random

import (
	"errors"
	"math"
)

// ============================================================================
// 无偏差整数映射
// ============================================================================

var (
	// ErrInvalidRange 表示范围参数无效（min > max）
	ErrInvalidRange = errors.New("random: invalid range, min must be <= max")
	// ErrEmptySlice 表示空切片操作
	ErrEmptySlice = errors.New("random: cannot sample from empty slice")
)

// Intn 返回 [0, n) 范围内的无偏差安全随机整数。
//
// 算法: 拒绝采样（Rejection Sampling）
//   - 避免取模偏差（modulo bias）
//   - 每个值出现概率严格 = 1/n
//
// 参数: n 为上界（不含），必须 > 0。
func Intn(n uint64) (uint64, error) {
	return IntnWithDRBG(globalDRBG, n)
}

// IntnWithDRBG 使用指定 DRBG 返回 [0, n) 范围内的无偏差安全随机整数。
func IntnWithDRBG(d *HMACDRBG, n uint64) (uint64, error) {
	if n == 0 {
		return 0, ErrInvalidRange
	}

	// 2^64 mod n 产生的偏差区间，拒绝该区间的值
	limit := math.MaxUint64 - (math.MaxUint64 % n)

	for {
		v, err := d.NextUint64()
		if err != nil {
			return 0, err
		}
		if v < limit {
			return v % n, nil
		}
		// v >= limit → 在偏差区间内，重新生成
	}
}

// IntRange 返回 [min, max] 范围内的无偏差安全随机整数。
func IntRange(min, max uint64) (uint64, error) {
	return IntRangeWithDRBG(globalDRBG, min, max)
}

// IntRangeWithDRBG 使用指定 DRBG 返回 [min, max] 范围内的无偏差安全随机整数。
func IntRangeWithDRBG(d *HMACDRBG, min, max uint64) (uint64, error) {
	if min > max {
		return 0, ErrInvalidRange
	}
	// 映射到 [0, max-min]，然后平移
	v, err := IntnWithDRBG(d, max-min+1)
	if err != nil {
		return 0, err
	}
	return min + v, nil
}

// Float64 返回 [0.0, 1.0) 范围内的安全随机浮点数。
//
// 实现: 取 53 位随机位（float64 的尾数位数），除以 2^53。
// 该方法保证在 [0,1) 上均匀分布，精度达到 2^-53。
func Float64() (float64, error) {
	return Float64WithDRBG(globalDRBG)
}

// Float64WithDRBG 使用指定 DRBG 返回 [0.0, 1.0) 范围内的安全随机浮点数。
func Float64WithDRBG(d *HMACDRBG) (float64, error) {
	// 使用 53 位（float64 尾数精度），确保所有可能值等概率
	v, err := IntnWithDRBG(d, 1<<53)
	if err != nil {
		return 0, err
	}
	return float64(v) / float64(1<<53), nil
}

// Float32 返回 [0.0, 1.0) 范围内的安全随机 float32。
func Float32() (float32, error) {
	v, err := Intn(1 << 24) // float32 尾数为 24 位（含隐含位）
	if err != nil {
		return 0, err
	}
	return float32(v) / float32(1<<24), nil
}

// ============================================================================
// Fisher-Yates 洗牌算法
// ============================================================================

// Shuffle 使用 Fisher-Yates 算法随机排列 n 个元素。
//
// 算法: Fisher-Yates Shuffle（又称 Knuth Shuffle）
//
//	for i := n - 1; i > 0; i-- {
//	    j := random(0, i)
//	    swap(i, j)
//	}
//
// 均匀性证明:
//   - 第 i 步时，位置 i 被赋值为剩余 (i+1) 个元素中任一个的概率 = 1/(i+1)
//   - 因此任意排列出现的概率 = 1/n × 1/(n-1) × ... × 1/1 = 1/n!
//   - 每种排列严格等概率
//
// 参数:
//   - n: 元素数量
//   - swap: 交换函数，接收两个索引 i 和 j
//
// 典型 Bingo 用法: Shuffle(75, func(i, j int) { balls[i], balls[j] = balls[j], balls[i] })
func Shuffle(n int, swap func(i, j int)) error {
	return ShuffleWithDRBG(globalDRBG, n, swap)
}

// ShuffleWithDRBG 使用指定 DRBG 执行 Fisher-Yates 洗牌。
func ShuffleWithDRBG(d *HMACDRBG, n int, swap func(i, j int)) error {
	if n < 1 {
		return ErrInvalidRange
	}

	for i := n - 1; i > 0; i-- {
		j, err := IntnWithDRBG(d, uint64(i+1))
		if err != nil {
			return err
		}
		swap(i, int(j))
	}
	return nil
}

// ============================================================================
// 不放回抽样（Bingo 抽球逻辑）
// ============================================================================

// Sample 从 [0, n) 范围中不放回地随机抽取 k 个不重复元素。
//
// 算法: 部分 Fisher-Yates 洗牌 + 拒绝采样
//  1. 预分配结果切片
//  2. 对于每次抽取: 生成 [0, n) 范围的随机数
//  3. 用 map 判重（拒绝采样变种）
//
// 对于 k << n 的情况（如 Bingo 每局抽球数远小于总球数），
// 期望拒绝次数极小，性能优良。
//
// Bingo 场景: Sample(75, 30) → 从 75 个球中抽取 30 个，顺序即为开奖顺序。
func Sample(n int, k int) ([]int, error) {
	return SampleWithDRBG(globalDRBG, n, k)
}

// SampleWithDRBG 使用指定 DRBG 执行不放回抽样。
func SampleWithDRBG(d *HMACDRBG, n int, k int) ([]int, error) {
	if n <= 0 || k < 0 {
		return nil, ErrInvalidRange
	}
	if k > n {
		k = n // 最多抽取 n 个
	}

	// 小 n 时使用 Fisher-Yates 部分洗牌（更高效）
	if k <= n/2 {
		return sampleRejection(d, n, k)
	}
	return samplePartialShuffle(d, n, k)
}

// sampleRejection 使用拒绝采样实现不放回抽样。
// 适合 k << n 的场景。
func sampleRejection(d *HMACDRBG, n int, k int) ([]int, error) {
	result := make([]int, k)
	seen := make(map[int]struct{}, k)

	for i := 0; i < k; i++ {
		for {
			v, err := IntnWithDRBG(d, uint64(n))
			if err != nil {
				return nil, err
			}
			idx := int(v)
			if _, exists := seen[idx]; !exists {
				result[i] = idx
				seen[idx] = struct{}{}
				break
			}
		}
	}
	return result, nil
}

// samplePartialShuffle 使用部分 Fisher-Yates 洗牌实现不放回抽样。
// 适合 k 接近 n 的场景。
func samplePartialShuffle(d *HMACDRBG, n int, k int) ([]int, error) {
	// 创建索引数组
	indices := make([]int, n)
	for i := range indices {
		indices[i] = i
	}

	// 部分洗牌：只需前 k 个位置
	for i := 0; i < k; i++ {
		j, err := IntRangeWithDRBG(d, uint64(i), uint64(n-1))
		if err != nil {
			return nil, err
		}
		indices[i], indices[int(j)] = indices[int(j)], indices[i]
	}

	return indices[:k], nil
}

// ============================================================================
// 批量样本生成（用于认证提交）
// ============================================================================

// SampleRawValues 生成指定数量的原始随机数样本，用于认证提交。
//
// 认证机构（如 iTech Labs）通常要求提供 50 万以上的原始随机数样本，
// 以验证随机分布的均匀性和独立性。
//
// 返回: count 个 uint64 原始随机数。
func SampleRawValues(count int) ([]uint64, error) {
	result := make([]uint64, count)
	for i := 0; i < count; i++ {
		v, err := globalDRBG.NextUint64()
		if err != nil {
			return nil, err
		}
		result[i] = v
	}
	return result, nil
}

// SampleMappedValues 生成指定数量的映射后样本（如模拟 Bingo 抽球）。
//
// 参数:
//   - count: 样本数量
//   - n: 映射范围 [0, n)
//
// 返回: count 个 [0, n) 范围内的整数。
func SampleMappedValues(count int, n uint64) ([]uint64, error) {
	result := make([]uint64, count)
	for i := 0; i < count; i++ {
		v, err := Intn(n)
		if err != nil {
			return nil, err
		}
		result[i] = v
	}
	return result, nil
}

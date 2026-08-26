// Package random 提供符合游戏认证标准的安全随机数生成系统。
//
// 本文件为向后兼容层，将原有 API 映射到 HMAC-DRBG 安全 RNG。
//
// 调用频率说明:
//   - Rand: 每次需要单个随机数时调用（如伤害波动、暴击判定），频率可达每秒千次以上
//   - RandN: 需要多个不重复随机值时调用（如随机奖励池选择）
//   - Weight: 加权随机选择时调用（如 PVE 排名匹配、掉落概率判定），频率中高
//   - WeightN: 多次不放回加权随机选择（如多阶段奖励分配）
package random

type Number interface {
	int16 | int32 | int64 | uint16 | uint32 | uint64 | int | uint | float32 | float64
}

// Rand 生成一个 [min, max] 范围内的均匀分布安全随机数。
//
// 底层实现:
//   - float32/float64: 使用 HMAC-DRBG 生成 53 位随机尾数，映射到 [0,1) 后线性缩放
//   - 整数类型: 使用 Intn 拒绝采样，确保每个值严格等概率
//
// 调用频率: 高频（每局游戏数百至数千次）
// 偏差保证: 无（拒绝采样消除取模偏差）
func Rand[T Number](min, max T) T {
	var zero T
	switch any(zero).(type) {
	case float64:
		f, _ := Float64()
		return min + T(f)*(max-min)
	case float32:
		f, _ := Float32()
		return min + T(f)*(max-min)
	default:
		val, _ := IntRange(uint64(min), uint64(max))
		return T(val)
	}
}

// RandN 生成 num 个 [min, max] 范围内不重复的安全随机数。
//
// 使用 map 判重 + 拒绝采样。如果 num 接近 (max-min+1) 且多次冲突，
// 第 11 次起放弃判重直接返回（防止死循环）。
func RandN[T Number](min, max T, num int) (rets []T) {
	tmps := map[T]struct{}{}
	for i := 0; i < num; i++ {
		rets = append(rets, randN(min, max, tmps, 10))
	}
	return
}

func randN[T Number](min, max T, tmps map[T]struct{}, times int) T {
	val := Rand(min, max)
	if _, ok := tmps[val]; ok {
		if times <= 0 {
			return val
		}
		return randN(min, max, tmps, times-1)
	}
	tmps[val] = struct{}{}
	return val
}

// Weight 根据权重函数，使用安全随机数选择一个索引。
//
// 算法:
//  1. 计算所有权重的总和 total
//  2. 使用安全 RNG 在 [0, total) 范围内生成随机数 rd
//  3. 线性扫描累加权重，首个累加值 >= rd 的索引即为结果
//
// 调用频率: 中高频（PVE 排名匹配、奖励分配等）
// 偏差保证: 无（权重累加等价于轮盘赌选择，每段概率 = w_i / Σw）
func Weight[T Number](size int, f func(int) T) int {
	var total T
	weights := make([]T, 0, size)
	for i := 0; i < size; i++ {
		w := f(i)
		total += w
		weights = append(weights, w)
	}

	// 生成 [0, total) 范围随机数，使用 Intn 确保无偏差
	rd := T(0)
	switch any(rd).(type) {
	case float64:
		f, _ := Float64()
		rd = T(f * float64(total))
	case float32:
		f, _ := Float32()
		rd = T(f * float32(total))
	default:
		v, _ := Intn(uint64(total))
		rd = T(v)
	}
	var temp T
	for i := 0; i < size; i++ {
		temp += weights[i]
		if temp > rd {
			return i
		}
	}
	// rd ∈ [0, total)，循环内必然返回。此处为安全回退。
	return size - 1
}

// WeightN 根据权重函数，不放回地选择 count 个不重复的索引。
//
// 算法:
//   - 每次选出一个索引后，从总和中减去其权重，确保剩余索引的相对概率不变
//   - 这等价于"按概率不放回抽样"，每次抽取严格遵循剩余候选集的权重分布
//
// 调用频率: 中频（多阶段奖励、组合抽奖）
func WeightN[T Number](size int, count int, f func(int) T) (rets []int) {
	var sum T
	weights := make([]T, 0, size)
	for i := 0; i < size; i++ {
		w := f(i)
		sum += w
		weights = append(weights, w)
	}
	tmps := map[int]struct{}{}
	for i := 0; i < count; i++ {
		// 生成 [0, sum) 范围随机数
		rd := T(0)
		switch any(rd).(type) {
		case float64:
			f, _ := Float64()
			rd = T(f * float64(sum))
		case float32:
			f, _ := Float32()
			rd = T(f * float32(sum))
		default:
			v, _ := Intn(uint64(sum))
			rd = T(v)
		}
		var temp T
		for j := 0; j < len(weights); j++ {
			if _, ok := tmps[j]; ok {
				continue
			}
			temp += weights[j]
			if temp > rd {
				rets = append(rets, j)
				tmps[j] = struct{}{}
				sum -= weights[j]
				break
			}
		}
	}
	return
}

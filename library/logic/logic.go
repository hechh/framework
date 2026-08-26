package logic

// Integer 约束所有整型
type Integer interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr
}

// Or 逻辑或，将多个标志位合并
func Or[T Integer](flags ...T) T {
	var ret T
	for _, flag := range flags {
		ret |= flag
	}
	return ret
}

// And 逻辑与，对多个标志位取与
func And[T Integer](flags ...T) T {
	if len(flags) == 0 {
		return 0
	}
	ret := flags[0]
	for _, flag := range flags[1:] {
		ret &= flag
	}
	return ret
}

// Has 检查 mask 中是否包含 flag 的所有位
func Has[T Integer](mask, flag T) bool {
	return mask&flag == flag
}

// SetBit 将 mask 中第 bit 位设置为 1 或 0（bit 从 0 开始）
func SetBit[T Integer](mask T, bit uint, on bool) T {
	if on {
		return mask | (1 << bit)
	}
	return mask & ^(1 << bit)
}

// GetBit 获取 mask 中第 bit 位的值（bit 从 0 开始）
func GetBit[T Integer](mask T, bit uint) bool {
	return mask&(1<<bit) != 0
}

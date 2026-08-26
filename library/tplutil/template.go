package tplutil

import (
	"math"
	"slices"
)

type Number interface {
	int16 | int32 | int64 | uint16 | uint32 | uint64 | int | uint | float32 | float64
}

func Max[T Number](a, b T) T {
	if a > b {
		return a
	}
	return b
}

func Min[T Number](a, b T) T {
	if a < b {
		return a
	}
	return b
}

func Floor[T Number](a, b T) T {
	if b == 0 {
		return 0
	}
	return T(math.Floor(float64(a) / float64(b)))
}

func Ceil[T Number](a, b T) T {
	if b == 0 {
		return 0
	}
	return T(math.Ceil(float64(a) / float64(b)))
}

func Abs[T Number](a T) T {
	if a < 0 {
		return -a
	}
	return a
}

func Or[T any](flag bool, a, b T) T {
	if flag {
		return a
	}
	return b
}

func Map2Keys[K comparable, V any](vals map[K]V) (rets []K) {
	rets = make([]K, 0, len(vals))
	for k := range vals {
		rets = append(rets, k)
	}
	return
}

func Map2Values[K comparable, V any](vals map[K]V) (rets []V) {
	rets = make([]V, 0, len(vals))
	for _, v := range vals {
		rets = append(rets, v)
	}
	return
}

// ------------------ 数组相关 ------------------
// Index 获取数组指定位置的元素，如果位置超出范围则返回默认值
func Index[T any](arr []T, pos int, def T) T {
	ll := len(arr)
	if pos < 0 {
		pos += ll
	}
	if pos < 0 || pos >= ll {
		return def
	}
	return arr[pos]
}

// Truncate 截断数组，返回一个新的切片（与原数组共享底层内存）
func Truncate[T any](arr []T, start, end int) []T {
	ll := len(arr)
	if ll == 0 {
		return nil
	}
	if start < 0 {
		start += ll
	}
	if start < 0 {
		start = 0
	}
	if end == 0 {
		end = ll
	} else if end < 0 {
		end += ll
	}
	if end > ll {
		end = ll
	}
	if start > end {
		return nil
	}
	return arr[start:end]
}

// Contains 检查数组是否包含指定元素
func Contains[T comparable](src []T, dst T) bool {
	return slices.Contains(src, dst)
}

// Filter 过滤数组，返回一个新的数组
func Filter[T comparable](arr []T, filters ...T) []T {
	if len(arr) == 0 {
		return arr
	}
	tmps := map[T]struct{}{}
	for _, v := range filters {
		tmps[v] = struct{}{}
	}
	result := make([]T, 0, len(arr))
	for _, elem := range arr {
		if _, ok := tmps[elem]; !ok {
			result = append(result, elem)
		}
	}
	return result
}

// Merge 合并多个数组，返回一个新的数组
func Merge[T any](as ...[]T) []T {
	total := 0
	for _, arr := range as {
		total += len(arr)
	}
	result := make([]T, 0, total)
	for _, arr := range as {
		result = append(result, arr...)
	}
	return result
}

// Copy 复制数组，返回一个新的数组
func Copy[T any](arr []T) []T {
	result := make([]T, len(arr))
	copy(result, arr)
	return result
}

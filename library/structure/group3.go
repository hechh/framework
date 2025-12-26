package structure

type Group3[T1 comparable, T2 comparable, T3 comparable, V any] map[Triple[T1, T2, T3]][]V

// 大小
func (d Group3[T1, T2, T3, V]) Size() int {
	return len(d)
}

// 设置
func (d Group3[T1, T2, T3, V]) Put(t1 T1, t2 T2, t3 T3, value V) {
	key := Triple[T1, T2, T3]{t1, t2, t3}
	d[key] = append(d[key], value)
}

// 读取
func (d Group3[T1, T2, T3, V]) Get(t1 T1, t2 T2, t3 T3) ([]V, bool) {
	key := Triple[T1, T2, T3]{t1, t2, t3}
	value, ok := d[key]
	return value, ok
}

// 删除
func (d Group3[T1, T2, T3, V]) Del(t1 T1, t2 T2, t3 T3) ([]V, bool) {
	key := Triple[T1, T2, T3]{t1, t2, t3}
	value, ok := d[key]
	if ok {
		delete(d, key)
	}
	return value, ok
}

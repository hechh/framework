package mapstruct

type Index3[T1 comparable, T2 comparable, T3 comparable] struct {
	f1 T1
	f2 T2
	f3 T3
}

type Map3[T1 comparable, T2 comparable, T3 comparable, V any] map[Index3[T1, T2, T3]]V

// 设置
func (d Map3[T1, T2, T3, V]) Set(t1 T1, t2 T2, t3 T3, value V) {
	d[Index3[T1, T2, T3]{t1, t2, t3}] = value
}

// 读取
func (d Map3[T1, T2, T3, V]) Get(t1 T1, t2 T2, t3 T3) (V, bool) {
	value, ok := d[Index3[T1, T2, T3]{t1, t2, t3}]
	return value, ok
}

// 删除
func (d Map3[T1, T2, T3, V]) Del(t1 T1, t2 T2, t3 T3) (V, bool) {
	value, ok := d[Index3[T1, T2, T3]{t1, t2, t3}]
	if ok {
		delete(d, Index3[T1, T2, T3]{t1, t2, t3})
	}
	return value, ok
}

// 大小
func (d Map3[T1, T2, T3, V]) Size() int {
	return len(d)
}

type Group3[T1 comparable, T2 comparable, T3 comparable, V any] map[Index3[T1, T2, T3]][]V

// 设置
func (d Group3[T1, T2, T3, V]) Put(t1 T1, t2 T2, t3 T3, value V) {
	d[Index3[T1, T2, T3]{t1, t2, t3}] = append(d[Index3[T1, T2, T3]{t1, t2, t3}], value)
}

// 读取
func (d Group3[T1, T2, T3, V]) Get(t1 T1, t2 T2, t3 T3) ([]V, bool) {
	value, ok := d[Index3[T1, T2, T3]{t1, t2, t3}]
	return value, ok
}

// 删除
func (d Group3[T1, T2, T3, V]) Del(t1 T1, t2 T2, t3 T3) ([]V, bool) {
	value, ok := d[Index3[T1, T2, T3]{t1, t2, t3}]
	if ok {
		delete(d, Index3[T1, T2, T3]{t1, t2, t3})
	}
	return value, ok
}

// 大小
func (d Group3[T1, T2, T3, V]) Size() int {
	return len(d)
}

package mapstruct

type Index4[T1 comparable, T2 comparable, T3 comparable, T4 comparable] struct {
	f1 T1
	f2 T2
	f3 T3
	f4 T4
}

type Map4[T1 comparable, T2 comparable, T3 comparable, T4 comparable, V any] map[Index4[T1, T2, T3, T4]]V

// 设置
func (d Map4[T1, T2, T3, T4, V]) Set(t1 T1, t2 T2, t3 T3, t4 T4, value V) {
	d[Index4[T1, T2, T3, T4]{t1, t2, t3, t4}] = value
}

// 读取
func (d Map4[T1, T2, T3, T4, V]) Get(t1 T1, t2 T2, t3 T3, t4 T4) (V, bool) {
	value, ok := d[Index4[T1, T2, T3, T4]{t1, t2, t3, t4}]
	return value, ok
}

// 删除
func (d Map4[T1, T2, T3, T4, V]) Del(t1 T1, t2 T2, t3 T3, t4 T4) (V, bool) {
	value, ok := d[Index4[T1, T2, T3, T4]{t1, t2, t3, t4}]
	if ok {
		delete(d, Index4[T1, T2, T3, T4]{t1, t2, t3, t4})
	}
	return value, ok
}

// 大小
func (d Map4[T1, T2, T3, T4, V]) Size() int {
	return len(d)
}

type Group4[T1 comparable, T2 comparable, T3 comparable, T4 comparable, V any] map[Index4[T1, T2, T3, T4]][]V

// 设置
func (d Group4[T1, T2, T3, T4, V]) Put(t1 T1, t2 T2, t3 T3, t4 T4, value V) {
	d[Index4[T1, T2, T3, T4]{t1, t2, t3, t4}] = append(d[Index4[T1, T2, T3, T4]{t1, t2, t3, t4}], value)
}

// 读取
func (d Group4[T1, T2, T3, T4, V]) Get(t1 T1, t2 T2, t3 T3, t4 T4) ([]V, bool) {
	value, ok := d[Index4[T1, T2, T3, T4]{t1, t2, t3, t4}]
	return value, ok
}

// 删除
func (d Group4[T1, T2, T3, T4, V]) Del(t1 T1, t2 T2, t3 T3, t4 T4) ([]V, bool) {
	value, ok := d[Index4[T1, T2, T3, T4]{t1, t2, t3, t4}]
	if ok {
		delete(d, Index4[T1, T2, T3, T4]{t1, t2, t3, t4})
	}
	return value, ok
}

// 大小
func (d Group4[T1, T2, T3, T4, V]) Size() int {
	return len(d)
}

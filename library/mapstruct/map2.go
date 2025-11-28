package mapstruct

type Index2[T1 comparable, T2 comparable] struct {
	f1 T1
	f2 T2
}

type Map2[T1 comparable, T2 comparable, V any] map[Index2[T1, T2]]V

// 设置
func (d Map2[T1, T2, V]) Set(t1 T1, t2 T2, value V) {
	d[Index2[T1, T2]{t1, t2}] = value
}

// 读取
func (d Map2[T1, T2, V]) Get(t1 T1, t2 T2) (V, bool) {
	value, ok := d[Index2[T1, T2]{t1, t2}]
	return value, ok
}

// 删除
func (d Map2[T1, T2, V]) Del(t1 T1, t2 T2) (V, bool) {
	value, ok := d[Index2[T1, T2]{t1, t2}]
	if ok {
		delete(d, Index2[T1, T2]{t1, t2})
	}
	return value, ok
}

// 大小
func (d Map2[T1, T2, V]) Size() int {
	return len(d)
}

type Group2[T1 comparable, T2 comparable, V any] map[Index2[T1, T2]][]V

// 设置
func (d Group2[T1, T2, V]) Put(t1 T1, t2 T2, value V) {
	d[Index2[T1, T2]{t1, t2}] = append(d[Index2[T1, T2]{t1, t2}], value)
}

// 读取
func (d Group2[T1, T2, V]) Get(t1 T1, t2 T2) ([]V, bool) {
	value, ok := d[Index2[T1, T2]{t1, t2}]
	return value, ok
}

// 删除
func (d Group2[T1, T2, V]) Del(t1 T1, t2 T2) ([]V, bool) {
	value, ok := d[Index2[T1, T2]{t1, t2}]
	if ok {
		delete(d, Index2[T1, T2]{t1, t2})
	}
	return value, ok
}

// 大小
func (d Group2[T1, T2, V]) Size() int {
	return len(d)
}

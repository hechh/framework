package structure

type Group1[T1 comparable, V any] map[T1][]V

func NewGroup1[T1 comparable, V any]() Group1[T1, V] {
	return make(map[T1][]V)
}

// 大小
func (d Group1[T1, V]) Size() int {
	return len(d)
}

// 设置
func (d Group1[T1, V]) Put(t1 T1, value V) {
	d[t1] = append(d[t1], value)
}

// 读取
func (d Group1[T1, V]) Get(t1 T1) ([]V, bool) {
	value, ok := d[t1]
	return value, ok
}

func (d Group1[T1, V]) Copy(t1 T1, f func(V) V) (rets []V, ok bool) {
	if values, ok := d[t1]; ok {
		rets = make([]V, len(values))
		for i, item := range values {
			rets[i] = f(item)
		}
	}
	return
}

// 删除
func (d Group1[T1, V]) Del(t1 T1) ([]V, bool) {
	value, ok := d[t1]
	if ok {
		delete(d, t1)
	}
	return value, ok
}

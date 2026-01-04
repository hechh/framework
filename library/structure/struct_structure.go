package structure

type Map1[T1 comparable, V any] map[T1]V
type Map2[T1 comparable, T2 comparable, V any] map[Pair[T1, T2]]V
type Map3[T1 comparable, T2 comparable, T3 comparable, V any] map[Triple[T1, T2, T3]]V
type Map4[T1 comparable, T2 comparable, T3 comparable, T4 comparable, V any] map[Quad[T1, T2, T3, T4]]V

// ----------map1----------------
func (d Map1[T1, V]) Size() int {
	return len(d)
}

func (d Map1[T1, V]) Put(t1 T1, value V) {
	d[t1] = value
}

func (d Map1[T1, V]) Get(t1 T1) (V, bool) {
	value, ok := d[t1]
	return value, ok
}

func (d Map1[T1, V]) Copy(t1 T1, f func(V) V) (V, bool) {
	value, ok := d[t1]
	if ok {
		value = f(value)
	}
	return value, ok
}

func (d Map1[T1, V]) Del(t1 T1) (V, bool) {
	value, ok := d[t1]
	delete(d, t1)
	return value, ok
}

// ----------map2----------------
func (d Map2[T1, T2, V]) Size() int {
	return len(d)
}

func (d Map2[T1, T2, V]) Put(t1 T1, t2 T2, value V) {
	d[Pair[T1, T2]{t1, t2}] = value
}

func (d Map2[T1, T2, V]) Get(t1 T1, t2 T2) (V, bool) {
	value, ok := d[Pair[T1, T2]{t1, t2}]
	return value, ok
}

func (d Map2[T1, T2, V]) Copy(t1 T1, t2 T2, f func(V) V) (V, bool) {
	value, ok := d[Pair[T1, T2]{t1, t2}]
	if ok {
		value = f(value)
	}
	return value, ok
}

func (d Map2[T1, T2, V]) Del(t1 T1, t2 T2) (V, bool) {
	key := Pair[T1, T2]{t1, t2}
	value, ok := d[key]
	if ok {
		delete(d, key)
	}
	return value, ok
}

// ----------map3----------------
func (d Map3[T1, T2, T3, V]) Size() int {
	return len(d)
}

func (d Map3[T1, T2, T3, V]) Put(t1 T1, t2 T2, t3 T3, value V) {
	d[Triple[T1, T2, T3]{t1, t2, t3}] = value
}

func (d Map3[T1, T2, T3, V]) Get(t1 T1, t2 T2, t3 T3) (V, bool) {
	value, ok := d[Triple[T1, T2, T3]{t1, t2, t3}]
	return value, ok
}

func (d Map3[T1, T2, T3, V]) Copy(t1 T1, t2 T2, t3 T3, f func(V) V) (V, bool) {
	value, ok := d[Triple[T1, T2, T3]{t1, t2, t3}]
	if ok {
		value = f(value)
	}
	return value, ok
}

func (d Map3[T1, T2, T3, V]) Del(t1 T1, t2 T2, t3 T3) (V, bool) {
	key := Triple[T1, T2, T3]{t1, t2, t3}
	value, ok := d[key]
	if ok {
		delete(d, key)
	}
	return value, ok
}

// ----------map4----------------
func (d Map4[T1, T2, T3, T4, V]) Size() int {
	return len(d)
}

func (d Map4[T1, T2, T3, T4, V]) Put(t1 T1, t2 T2, t3 T3, t4 T4, value V) {
	d[Quad[T1, T2, T3, T4]{t1, t2, t3, t4}] = value
}

func (d Map4[T1, T2, T3, T4, V]) Get(t1 T1, t2 T2, t3 T3, t4 T4) (V, bool) {
	value, ok := d[Quad[T1, T2, T3, T4]{t1, t2, t3, t4}]
	return value, ok
}

func (d Map4[T1, T2, T3, T4, V]) Copy(t1 T1, t2 T2, t3 T3, t4 T4, f func(V) V) (V, bool) {
	value, ok := d[Quad[T1, T2, T3, T4]{t1, t2, t3, t4}]
	if ok {
		value = f(value)
	}
	return value, ok
}

func (d Map4[T1, T2, T3, T4, V]) Del(t1 T1, t2 T2, t3 T3, t4 T4) (V, bool) {
	key := Quad[T1, T2, T3, T4]{t1, t2, t3, t4}
	value, ok := d[key]
	if ok {
		delete(d, key)
	}
	return value, ok
}

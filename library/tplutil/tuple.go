package tplutil

// 二元组（用作 map key 性能极高）
type Tuple2[T1, T2 comparable] struct {
	V1 T1
	V2 T2
}

// 创建二元组 key
func T2[T1, T2 comparable](v1 T1, v2 T2) Tuple2[T1, T2] {
	return Tuple2[T1, T2]{
		V1: v1,
		V2: v2,
	}
}

// 三元组
type Tuple3[T1, T2, T3 comparable] struct {
	V1 T1
	V2 T2
	V3 T3
}

// 创建三元组 key
func T3[T1, T2, T3 comparable](v1 T1, v2 T2, v3 T3) Tuple3[T1, T2, T3] {
	return Tuple3[T1, T2, T3]{
		V1: v1,
		V2: v2,
		V3: v3,
	}
}

// 四元组
type Tuple4[T1, T2, T3, T4 comparable] struct {
	V1 T1
	V2 T2
	V3 T3
	V4 T4
}

// 创建四元组 key
func T4[T1, T2, T3, T4 comparable](v1 T1, v2 T2, v3 T3, v4 T4) Tuple4[T1, T2, T3, T4] {
	return Tuple4[T1, T2, T3, T4]{
		V1: v1,
		V2: v2,
		V3: v3,
		V4: v4,
	}
}

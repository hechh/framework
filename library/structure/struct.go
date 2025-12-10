package structure

type Pair[T any, U any] struct {
	First  T
	Second U
}

type Triple[A any, B any, C any] struct {
	First  A
	Second B
	Three  C
}

type Quad[A any, B any, C any, D any] struct {
	First  A
	Second B
	Three  C
	Four   D
}

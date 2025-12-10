package test

import "testing"

func Read(val []byte) int {
	str := []byte("123413")
	copy(val, str)
	return len(str)
}

func TestRead(t *testing.T) {
	items := make([]byte, 1000)
	t.Log(len(items), cap(items))
	ll := Read(items)
	t.Log(len(items), ll, string(items[:ll+1]))
}

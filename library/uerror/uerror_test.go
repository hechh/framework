package uerror

import "testing"

func TestUError(t *testing.T) {
	err := Err(0, "this is a testing")
	t.Log(err.Error())
}

package util

import (
	"os"
	"os/signal"
	"reflect"
	"syscall"
	"time"
	"unsafe"
)

func StringToBytes(str string) []byte {
	if len(str) == 0 {
		return nil
	}
	s := *(*reflect.StringHeader)(unsafe.Pointer(&str))
	b := &reflect.SliceHeader{Data: s.Data, Len: s.Len, Cap: s.Len}
	return *(*[]byte)(unsafe.Pointer(b))
}

func BytesToString(bts []byte) string {
	if len(bts) == 0 {
		return ""
	}
	b := *(*reflect.SliceHeader)(unsafe.Pointer(&bts))
	s := &reflect.StringHeader{Data: b.Data, Len: b.Len}
	return *(*string)(unsafe.Pointer(s))
}

func Retry(attempts int, sleep time.Duration, f func() error) (err error) {
	for i := 0; i < attempts; i++ {
		if err = f(); err == nil {
			return
		}
		time.Sleep(sleep)
		sleep *= 2
	}
	return err
}

func Signal(ff func(), sigs ...os.Signal) {
	defaults := []os.Signal{syscall.SIGABRT, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT}
	if len(sigs) > 0 {
		defaults = append(defaults, sigs...)
	}
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, defaults...)

	<-sig
	ff()
}

func Or[T any](flag bool, a, b T) T {
	if flag {
		return a
	}
	return b
}

func Index[T any](arr []T, pos int, def T) T {
	ll := len(arr)
	if pos < 0 || pos >= ll {
		return def
	}
	return arr[pos]
}

func Prefix[T any](arr []T, pos int) []T {
	ll := len(arr)
	if pos < 0 || pos >= ll {
		return arr
	}
	return arr[:pos]
}

func Suffix[T any](arr []T, pos int) []T {
	ll := len(arr)
	if pos < 0 || pos >= ll {
		return arr
	}
	return arr[pos:]
}

func Must(err error) {
	if err != nil {
		panic(err)
	}
}

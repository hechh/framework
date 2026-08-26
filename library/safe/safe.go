package safe

import (
	"runtime/debug"
	"time"
	"unsafe"
)

// Recover 安全恢复函数,捕获panic并记录
func Recover(except func(string, ...any), f func()) {
	defer func() {
		if err := recover(); err != nil && except != nil {
			except("PANIC: %v\nStack Trace:\n%s", err, string(debug.Stack()))
		}
	}()
	f()
}

// SafeGo 安全启动goroutine,自动处理panic
func SafeGo(except func(string, ...any), f func()) {
	go Recover(except, f)
}

// StringToBytes 将字符串转换为字节数组
func StringToBytes(str string) []byte {
	if len(str) == 0 {
		return nil
	}
	return unsafe.Slice(unsafe.StringData(str), len(str))
}

// BytesToString 将字节数组转换为字符串
func BytesToString(bts []byte) string {
	if len(bts) == 0 {
		return ""
	}
	return unsafe.String(unsafe.SliceData(bts), len(bts))
}

// Retry 带重试机制的函数执行
func Retry(attempts int, sleep time.Duration, f func() error) (err error) {
	for range attempts {
		if err = f(); err == nil {
			return
		}
		time.Sleep(sleep)
	}
	return
}

// Must 执行函数并检查错误,如果有错误则panic
func Must(f func() error) {
	if err := f(); err != nil {
		panic(err)
	}
}

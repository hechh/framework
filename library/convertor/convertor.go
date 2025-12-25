package convertor

import (
	"time"

	"github.com/spf13/cast"
)

var (
	data = make(map[string]*Convertor)
)

type Convertor struct {
	origin string
	target string
	conv   func(string) any
}

func Wrapper[T any](f func(any) T) func(string) any {
	return func(val string) any {
		return f(val)
	}
}

func Register(f func(string) any, target string, origins ...string) {
	for _, origin := range origins {
		data[origin] = &Convertor{
			origin: origin,
			target: target,
			conv:   f,
		}
	}
}

func Target(origin string) string {
	if item, ok := data[origin]; ok {
		return item.target
	}
	return origin
}

func Convert(origin string, val string) any {
	if item, ok := data[origin]; ok {
		return item.conv(val)
	}
	return val
}

func init() {
	Register(Wrapper(cast.ToUint32), "uint32", "uint32", "uint8", "uint16")
	Register(Wrapper(cast.ToInt32), "int32", "int32", "int8", "int16")
	Register(Wrapper(cast.ToUint64), "uint64", "uint64")
	Register(Wrapper(cast.ToInt64), "int64", "int64")
	Register(Wrapper(cast.ToFloat32), "float32", "float")
	Register(Wrapper(cast.ToFloat64), "float64", "double")
	Register(Wrapper(cast.ToBool), "bool", "bool")

	// 特殊类型转换
	Register(timestampToInt64, "int64", "timestamp")
}

func timestampToInt64(str string) any {
	tt, err := time.ParseInLocation("2006-01-02 15:04:05", str, time.UTC)
	if err != nil {
		panic(err)
	}
	return tt.Unix()
}

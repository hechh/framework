package logger

import (
	"strings"
	"sync"
	"time"
)

const (
	LOG_TRACE = 1
	LOG_DEBUG = 2
	LOG_WARN  = 3
	LOG_INFO  = 4
	LOG_ERROR = 5
	LOG_FATAL = 6
)

type Meta struct {
	FileName string
	Line     int
	FuncName string
	Level    int32
	Msg      string
}

// 日志数据接口
type IData interface {
	Now() time.Time // 获取时间
	Add(int32)      // 并发次数
	Done() int32    // 完成次数
	Write(Meta)     // 写入数据
	Read() []byte   // 读取数据
}

// 日志写入接口
type IWriter interface {
	Push(IData) // 推送日志
	Close()     // 关闭
}

var (
	dataPool = sync.Pool{
		New: func() any {
			return NewData()
		},
	}
)

func get(times int) IData {
	obj := dataPool.Get().(IData)
	obj.Add(int32(times))
	return obj
}

func put(obj IData) {
	if obj.Done() == 0 {
		dataPool.Put(obj)
	}
}

func LevelToString(level int32) string {
	switch level {
	case LOG_TRACE:
		return "TRACE"
	case LOG_DEBUG:
		return "DEBUG"
	case LOG_WARN:
		return "WARN"
	case LOG_INFO:
		return "INFO"
	case LOG_ERROR:
		return "ERROR"
	case LOG_FATAL:
		return "FATAL"
	}
	return ""
}

func StringToLevel(str string) int32 {
	switch strings.ToUpper(str) {
	case "TRACE":
		return LOG_TRACE
	case "DEBUG":
		return LOG_DEBUG
	case "WARN":
		return LOG_WARN
	case "INFO":
		return LOG_INFO
	case "ERROR":
		return LOG_ERROR
	case "FATAL":
		return LOG_FATAL
	}
	return LOG_WARN
}

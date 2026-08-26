package mlog

import (
	"strings"
	"sync/atomic"
)

const (
	LOG_TRACE   = 1
	LOG_DEBUG   = 2
	LOG_WARN    = 3
	LOG_INFO    = 4
	LOG_ERROR   = 5
	LOG_FATAL   = 6
	FORMAT_LINE = 1
	FORMAT_JSON = 2
)

var (
	object atomic.Pointer[Logger]
)

func init() {
	object.Store(NewLogger())
}

func Level2Name(level int32) string {
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
	default:
		return "INFO"
	}
}

func Name2Level(name string) int32 {
	switch strings.ToUpper(name) {
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
	default:
		return LOG_INFO
	}
}

func Name2Format(f string) int32 {
	switch strings.ToUpper(f) {
	case "JSON":
		return FORMAT_JSON
	default:
		return FORMAT_LINE
	}
}

func SetObject(oj *Logger) {
	object.Store(oj)
}

func SetLevel(level int32) {
	if obj := object.Load(); obj != nil {
		obj.SetLevel(level)
	}
}

func SetFormat(f int32) {
	if obj := object.Load(); obj != nil {
		obj.SetFormat(f)
	}
}

func SetCaller(val bool) {
	if obj := object.Load(); obj != nil {
		obj.SetCaller(val)
	}
}

func Tracef(format string, args ...any) {
	if obj := object.Load(); obj != nil {
		obj.Outputf(2, LOG_TRACE, "", format, args...)
	}
}

func Debugf(format string, args ...any) {
	if obj := object.Load(); obj != nil {
		obj.Outputf(2, LOG_DEBUG, "", format, args...)
	}
}

// Warnf 记录WARN级别格式化日志
func Warnf(format string, args ...any) {
	if obj := object.Load(); obj != nil {
		obj.Outputf(2, LOG_WARN, "", format, args...)
	}
}

// Infof 记录INFO级别格式化日志
func Infof(format string, args ...any) {
	if obj := object.Load(); obj != nil {
		obj.Outputf(2, LOG_INFO, "", format, args...)
	}
}

// Errorf 记录ERROR级别格式化日志
func Errorf(format string, args ...any) {
	if obj := object.Load(); obj != nil {
		obj.Outputf(2, LOG_ERROR, "", format, args...)
	}
}

// Fatalf 记录FATAL级别格式化日志
func Fatalf(format string, args ...any) {
	if obj := object.Load(); obj != nil {
		obj.Outputf(2, LOG_FATAL, "", format, args...)
	}
}

func Trace(args ...any) {
	if obj := object.Load(); obj != nil {
		obj.Output(2, LOG_TRACE, "", args...)
	}
}

func Debug(args ...any) {
	if obj := object.Load(); obj != nil {
		obj.Output(2, LOG_DEBUG, "", args...)
	}
}

func Warn(args ...any) {
	if obj := object.Load(); obj != nil {
		obj.Output(2, LOG_WARN, "", args...)
	}
}

func Info(args ...any) {
	if obj := object.Load(); obj != nil {
		obj.Output(2, LOG_INFO, "", args...)
	}
}

func Error(args ...any) {
	if obj := object.Load(); obj != nil {
		obj.Output(2, LOG_ERROR, "", args...)
	}
}

func Fatal(args ...any) {
	if obj := object.Load(); obj != nil {
		obj.Output(2, LOG_FATAL, "", args...)
	}
}

func Output(skip int, level int32, tag string, args ...any) {
	if obj := object.Load(); obj != nil {
		obj.Output(skip+1, level, tag, args...)
	}
}

func Outputf(skip int, level int32, tag, format string, args ...any) {
	if obj := object.Load(); obj != nil {
		obj.Outputf(skip+1, level, tag, format, args...)
	}
}

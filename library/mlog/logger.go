package mlog

import (
	"fmt"
	"path"
	"runtime"
	"sync/atomic"

	"github.com/spf13/cast"
)

type Logger struct {
	level int32
	list  []IWriter
}

func NewLogger(level any, ws ...IWriter) *Logger {
	var logLevel int32
	switch vv := level.(type) {
	case string:
		logLevel = StringToLevel(vv)
	default:
		logLevel = cast.ToInt32(vv)
	}
	return &Logger{
		level: logLevel,
		list:  ws,
	}
}

func (d *Logger) Close() {
	for _, w := range d.list {
		w.Close()
	}
}

func (d *Logger) SetLevel(level int32) {
	atomic.StoreInt32(&d.level, level)
}

func (d *Logger) Trace(skip int, format string, args ...any) {
	if atomic.LoadInt32(&d.level) <= LOG_TRACE {
		d.output(skip+1, LOG_TRACE, fmt.Sprintf(format, args...))
	}
}

func (d *Logger) Debug(skip int, format string, args ...any) {
	if atomic.LoadInt32(&d.level) <= LOG_DEBUG {
		d.output(skip+1, LOG_DEBUG, fmt.Sprintf(format, args...))
	}
}

func (d *Logger) Warn(skip int, format string, args ...any) {
	if atomic.LoadInt32(&d.level) <= LOG_WARN {
		d.output(skip+1, LOG_WARN, fmt.Sprintf(format, args...))
	}
}

func (d *Logger) Info(skip int, format string, args ...any) {
	if atomic.LoadInt32(&d.level) <= LOG_INFO {
		d.output(skip+1, LOG_INFO, fmt.Sprintf(format, args...))
	}
}

func (d *Logger) Error(skip int, format string, args ...any) {
	if atomic.LoadInt32(&d.level) <= LOG_ERROR {
		d.output(skip+1, LOG_ERROR, fmt.Sprintf(format, args...))
	}
}

func (d *Logger) Fatal(skip int, format string, args ...any) {
	if atomic.LoadInt32(&d.level) <= LOG_FATAL {
		d.output(skip+1, LOG_FATAL, fmt.Sprintf(format, args...))
	}
}

func (d *Logger) output(depth int, level int32, msg string) {
	meta := Meta{Level: level, Msg: msg}
	if depth > 0 {
		pc, file, line, _ := runtime.Caller(depth + 1)
		fname := path.Base(runtime.FuncForPC(pc).Name())
		meta.FileName = file
		meta.Line = line
		meta.FuncName = fname
	}
	data := get(len(d.list))
	data.Write(meta)
	for _, w := range d.list {
		w.Push(data)
	}
}

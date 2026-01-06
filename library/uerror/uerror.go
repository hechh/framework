package uerror

import (
	"fmt"
	"path"
	"runtime"
)

type UError struct {
	file  string
	fname string
	line  int
	code  int32
	msg   string
}

func New(code int32, format string, args ...any) *UError {
	pc, file, line, _ := runtime.Caller(1)
	fname := runtime.FuncForPC(pc).Name()
	return &UError{
		file:  path.Base(file),
		line:  line,
		fname: path.Base(fname),
		code:  code,
		msg:   fmt.Sprintf(format, args...),
	}
}

func Err(code int32, format string, args ...any) *UError {
	return &UError{
		code: code,
		msg:  fmt.Sprintf(format, args...),
	}
}

func Wrap(code int32, err error) *UError {
	return &UError{code: code, msg: err.Error()}
}

func Turn(code int32, err error) *UError {
	if vv, ok := err.(*UError); ok {
		return vv
	}
	return &UError{
		code: code,
		msg:  err.Error(),
	}
}

func (ue *UError) Error() string {
	if len(ue.file) <= 0 {
		return ue.msg
		//return fmt.Sprintf("[%d] %s", ue.code, ue.msg)
	}
	return fmt.Sprintf("%s:%d %s %s", ue.file, ue.line, ue.fname, ue.msg)
	//return fmt.Sprintf("%s:%d %s [%d] %s", ue.file, ue.line, ue.fname, ue.code, ue.msg)
}

func (ue *UError) GetFile() string {
	return ue.file
}

func (ue *UError) GetFunc() string {
	return ue.fname
}

func (ue *UError) GetLine() int {
	return ue.line
}

func (ue *UError) GetCode() int32 {
	return ue.code
}

func (ue *UError) GetMsg() string {
	return ue.msg
}

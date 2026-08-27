package uerror

import (
	"fmt"
	"path"
	"runtime"

	"github.com/bytedance/sonic"
)

type UError struct {
	file  string
	fname string
	line  int
	code  int32
	msg   string
}

func (d *UError) Printf(format string, args ...any) {
	d.msg = fmt.Sprintf(format, args...)
}

func (d *UError) Print(args ...any) {
	buf, _ := sonic.Marshal(args)
	d.msg = string(buf)
}

func (d *UError) GetCode() int32 {
	return d.code
}

func (d *UError) GetMsg() string {
	return d.msg
}

func (d *UError) Error() string {
	return fmt.Sprintf("[%d] %s:%d/%s %s", d.code, d.file, d.line, d.fname, d.msg)
}

type IRsp interface {
	SetRspHead(int32, string)
}

func SetRspHead(irsp any, err error) {
	if rsp, ok := irsp.(IRsp); ok && rsp != nil {
		switch vv := err.(type) {
		case *UError:
			rsp.SetRspHead(vv.GetCode(), vv.GetMsg())
		case nil:
		default:
			rsp.SetRspHead(-1, err.Error())
		}
	}
}

func Err(code int32, format string, args ...any) *UError {
	pc, file, line, _ := runtime.Caller(2)
	return &UError{
		file:  path.Base(file),
		line:  line,
		fname: path.Base(runtime.FuncForPC(pc).Name()),
		code:  code,
		msg:   fmt.Sprintf(format, args...),
	}
}

func Wrap(code int32, err error) *UError {
	if uerr, ok := err.(*UError); ok && uerr != nil {
		return uerr
	}
	pc, file, line, _ := runtime.Caller(2)
	return &UError{
		file:  path.Base(file),
		line:  line,
		fname: path.Base(runtime.FuncForPC(pc).Name()),
		code:  code,
		msg:   err.Error(),
	}
}

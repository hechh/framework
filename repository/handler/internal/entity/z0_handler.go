package entity

import (
	"framework/define"
	"framework/repository/handler/domain"
	"reflect"
	"time"
)

type Z0Handler[Actor any] struct {
	*Common
	EmptyCrypto
	method domain.Z0Func[Actor]
}

func NewZ0Handler[Actor any](nodeType int32, cmd int32, method domain.Z0Func[Actor]) *Z0Handler[Actor] {
	return &Z0Handler[Actor]{
		Common: NewCommon(nodeType, cmd, ParseActorFunc(reflect.ValueOf(method))),
		method: method,
	}
}

func (d *Z0Handler[Actor]) Call(obj any, head define.IContext, args ...any) func() {
	return func() {
		// 参数解析
		actor, ok := any(obj).(*Actor)
		if !ok {
			head.Errorf("参数类型错误")
			return
		}

		// 接口调用
		startMs := time.Now().UnixMilli()
		err := d.method(actor, head)
		endMs := time.Now().UnixMilli()
		if err != nil {
			head.Errorf("接口耗时%d毫秒\terror:%v", endMs-startMs, err)
		} else {
			head.Tracef("接口耗时%d毫秒", endMs-startMs)
		}
	}
}

func (d *Z0Handler[Actor]) Rpc(obj any, head define.IContext, body []byte) func() {
	return func() {
		// 参数解析
		actor, ok := any(obj).(*Actor)
		if !ok {
			head.Errorf("参数类型错误")
			return
		}

		// 接口调用
		startMs := time.Now().UnixMilli()
		err := d.method(actor, head)
		endMs := time.Now().UnixMilli()
		if err != nil {
			head.Errorf("接口耗时%d毫秒\terror:%v", endMs-startMs, err)
		} else {
			head.Tracef("接口耗时%d毫秒", endMs-startMs)
		}
	}
}

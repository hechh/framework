package entity

import (
	"framework/define"
	"framework/repository/handler/domain"
	"reflect"
	"time"
)

type Z1Handler[T any] struct {
	*Common
	EmptyCrypto
	method domain.Z1Func[T]
}

func NewZ1Handler[T any](nodeType int32, cmd int32, method domain.Z1Func[T]) *Z1Handler[T] {
	return &Z1Handler[T]{
		Common: NewCommon(nodeType, cmd, ParseActorFunc(reflect.ValueOf(method))),
		method: method,
	}
}

func (d *Z1Handler[T]) Call(obj any, head define.IContext, args ...any) func() {
	return func() {
		// 参数解析
		actor, ok := any(obj).(*T)
		if !ok {
			head.Errorf("调用接口参数错误")
			return
		}

		// 接口调用
		startMs := time.Now().UnixMilli()
		d.method(actor, head, args[0])
		endMs := time.Now().UnixMilli()
		head.Tracef("接口耗时%d毫秒\treq:%v", endMs-startMs, args[0])
	}
}

func (d Z1Handler[T]) Rpc(obj any, head define.IContext, body []byte) func() {
	return func() {
		// 参数解析
		actor, ok := any(obj).(*T)
		if !ok {
			head.Errorf("调用接口参数错误")
			return
		}

		// 接口调用
		startMs := time.Now().UnixMilli()
		d.method(actor, head, body)
		endMs := time.Now().UnixMilli()
		head.Tracef("接口耗时%d毫秒\treq:%v", endMs-startMs, body)
	}
}

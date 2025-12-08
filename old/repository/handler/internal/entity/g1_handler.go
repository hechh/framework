package entity

import (
	"framework/define"
	"framework/repository/handler/domain"
	"reflect"
	"time"
)

type G1Handler[T any, A any] struct {
	*Common
	GobCrypto
	method domain.G1Func[T, A]
}

func NewG1Handler[T any, A any](nodeType int32, cmd int32, method domain.G1Func[T, A]) *G1Handler[T, A] {
	return &G1Handler[T, A]{
		Common: NewCommon(nodeType, cmd, ParseActorFunc(reflect.ValueOf(method))),
		method: method,
	}
}

func (d *G1Handler[T, A]) Call(obj any, head define.IContext, args ...any) func() {
	return func() {
		// 参数解析
		req, ok := args[0].(A)
		if !ok {
			head.Errorf("调用接口参数错误")
			return
		}
		actor, ok := any(obj).(*T)
		if !ok {
			head.Errorf("调用接口参数错误")
			return
		}

		// 接口调用
		startMs := time.Now().UnixMilli()
		err := d.method(actor, head, req)
		endMs := time.Now().UnixMilli()
		if err != nil {
			head.Errorf("接口耗时%d毫秒\treq:%v\terror:%v", endMs-startMs, req, err)
		} else {
			head.Tracef("接口耗时%d毫秒\treq:%v", endMs-startMs, req)
		}
	}
}

func (d G1Handler[T, A]) Rpc(obj any, head define.IContext, body []byte) func() {
	return func() {
		// 参数解析
		req := new(A)
		if err := d.Unmarshal(body, req); err != nil {
			head.Errorf("调用接口参数错误:%v", err)
			return
		}
		actor, ok := any(obj).(*T)
		if !ok {
			head.Errorf("调用接口参数错误")
			return
		}

		// 接口调用
		startMs := time.Now().UnixMilli()
		err := d.method(actor, head, *req)
		endMs := time.Now().UnixMilli()
		if err != nil {
			head.Errorf("接口耗时%d毫秒\treq:%v\terror:%v", endMs-startMs, *req, err)
		} else {
			head.Tracef("接口耗时%d毫秒\treq:%v", endMs-startMs, *req)
		}
	}
}

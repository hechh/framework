package entity

import (
	"framework/define"
	"framework/repository/handler/domain"
	"reflect"
	"time"
)

type L1Handler[Actor any, L1 any] struct {
	*Common
	ErrorCrypto
	method domain.L1Func[Actor, L1]
}

func NewL1Handler[Actor any, L1 any](nodeType int32, cmd int32, method domain.L1Func[Actor, L1]) *L1Handler[Actor, L1] {
	return &L1Handler[Actor, L1]{
		Common: NewCommon(nodeType, cmd, ParseActorFunc(reflect.ValueOf(method))),
		method: method,
	}
}

func (d *L1Handler[Actor, L1]) Call(obj any, head define.IContext, args ...any) func() {
	return func() {
		// 参数解析
		req1, ok := args[0].(L1)
		if !ok {
			head.Errorf("参数类型错误")
			return
		}
		actor, ok := any(obj).(*Actor)
		if !ok {
			head.Errorf("参数类型错误")
			return
		}

		// 接口调用
		startMs := time.Now().UnixMilli()
		err := d.method(actor, req1)
		endMs := time.Now().UnixMilli()
		if err != nil {
			head.Errorf("接口耗时%d毫秒\terror:%v", endMs-startMs, err)
		} else {
			head.Tracef("接口耗时%d毫秒", endMs-startMs)
		}
	}
}

func (d *L1Handler[Actor, L1]) Rpc(obj any, head define.IContext, body []byte) func() {
	return func() {
		head.Errorf("该接口不支持远程调用")
	}
}

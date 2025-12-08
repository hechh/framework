package entity

import (
	"framework/define"
	"framework/repository/handler/domain"
	"reflect"
	"time"
)

type L2Handler[Actor any, L1 any, L2 any] struct {
	*Common
	ErrorCrypto
	method domain.L2Func[Actor, L1, L2]
}

func NewL2Handler[Actor any, L1 any, L2 any](nodeType int32, cmd int32, method domain.L2Func[Actor, L1, L2]) *L2Handler[Actor, L1, L2] {
	return &L2Handler[Actor, L1, L2]{
		Common: NewCommon(nodeType, cmd, ParseActorFunc(reflect.ValueOf(method))),
		method: method,
	}
}

func (d *L2Handler[Actor, L1, L2]) Call(obj any, head define.IContext, args ...any) func() {
	return func() {
		// 参数解析
		req1, ok := args[0].(L1)
		if !ok {
			head.Errorf("参数类型错误")
			return
		}
		req2, ok := args[0].(L2)
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
		err := d.method(actor, req1, req2)
		endMs := time.Now().UnixMilli()
		if err != nil {
			head.Errorf("接口耗时%d毫秒\terror:%v", endMs-startMs, err)
		} else {
			head.Tracef("接口耗时%d毫秒", endMs-startMs)
		}
	}
}

func (d *L2Handler[Actor, L1, L2]) Rpc(obj any, head define.IContext, body []byte) func() {
	return func() {
		head.Errorf("该接口不支持远程调用")
	}
}

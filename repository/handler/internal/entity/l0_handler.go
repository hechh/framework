package entity

import (
	"framework/define"
	"framework/repository/handler/domain"
	"reflect"
	"time"
)

type L0Handler[Actor any] struct {
	*Common
	ErrorCrypto
	method domain.L0Func[Actor]
}

func NewL0Handler[Actor any](nodeType int32, cmd int32, method domain.L0Func[Actor]) *L0Handler[Actor] {
	return &L0Handler[Actor]{
		Common: NewCommon(nodeType, cmd, ParseActorFunc(reflect.ValueOf(method))),
		method: method,
	}
}

func (d *L0Handler[Actor]) Call(obj any, head define.IContext, args ...any) func() {
	return func() {
		// 参数解析
		actor, ok := any(obj).(*Actor)
		if !ok {
			head.Errorf("参数类型错误")
			return
		}

		// 接口调用
		startMs := time.Now().UnixMilli()
		err := d.method(actor)
		endMs := time.Now().UnixMilli()
		if err != nil {
			head.Errorf("接口耗时%d毫秒\terror:%v", endMs-startMs, err)
		} else {
			head.Tracef("接口耗时%d毫秒", endMs-startMs)
		}
	}
}

func (d *L0Handler[Actor]) Rpc(obj any, head define.IContext, body []byte) func() {
	return func() {
		head.Errorf("该接口不支持远程调用")
	}
}

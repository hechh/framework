package entity

import (
	"framework/define"
	"framework/repository/handler/domain"
	"reflect"
	"time"
)

type G2Handler[Actor any, V1 any, V2 any] struct {
	*Common
	GobCrypto
	method domain.G2Func[Actor, V1, V2]
}

func NewG2Handler[Actor any, V1 any, V2 any](nodeType int32, cmd int32, method domain.G2Func[Actor, V1, V2]) *G2Handler[Actor, V1, V2] {
	return &G2Handler[Actor, V1, V2]{
		Common: NewCommon(nodeType, cmd, ParseActorFunc(reflect.ValueOf(method))),
		method: method,
	}
}

func (d *G2Handler[Actor, V1, V2]) Call(obj any, head define.IContext, args ...any) func() {
	return func() {
		// 参数解析
		req1, ok := args[0].(V1)
		if !ok {
			head.Errorf("调用接口参数错误")
			return
		}
		req2, ok := args[1].(V2)
		if !ok {
			head.Errorf("调用接口参数错误")
			return
		}
		actor, ok := any(obj).(*Actor)
		if !ok {
			head.Errorf("调用接口参数错误")
			return
		}

		// 接口调用
		startMs := time.Now().UnixMilli()
		err := d.method(actor, head, req1, req2)
		endMs := time.Now().UnixMilli()
		if err != nil {
			head.Errorf("接口耗时%d毫秒\treq1:%v\treq2:%v\terror:%v", endMs-startMs, req1, req2, err)
		} else {
			head.Tracef("接口耗时%d毫秒\treq1:%v\treq2:%v", endMs-startMs, req1, req2)
		}
	}
}

func (d *G2Handler[Actor, V1, V2]) Rpc(obj any, head define.IContext, body []byte) func() {
	return func() {
		// 参数解析
		req1 := new(V1)
		req2 := new(V2)
		if err := d.Unmarshal(body, req1, req2); err != nil {
			head.Errorf("参数解析错误:%v", err)
			return
		}
		actor, ok := any(obj).(*Actor)
		if !ok {
			head.Errorf("调用接口参数错误")
			return
		}

		// 接口调用
		startMs := time.Now().UnixMilli()
		err := d.method(actor, head, *req1, *req2)
		endMs := time.Now().UnixMilli()
		if err != nil {
			head.Errorf("接口耗时%d毫秒\treq1:%v\treq2:%v\terror:%v", endMs-startMs, *req1, *req2, err)
		} else {
			head.Tracef("接口耗时%d毫秒\treq1:%v\treq2:%v", endMs-startMs, *req1, *req2)
		}
	}
}

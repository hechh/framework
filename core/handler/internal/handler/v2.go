package handler

import (
	"framework/core"
	"reflect"
	"time"
)

type V2Handler[Actor any, V1 any, V2 any] struct {
	*Base
	core.ISerialize
	method core.V2Func[Actor, V1, V2]
}

func NewV2Handler[Actor any, V1 any, V2 any](en core.ISerialize, nodeType uint32, cmd uint32, f core.V2Func[Actor, V1, V2]) *V2Handler[Actor, V1, V2] {
	return &V2Handler[Actor, V1, V2]{
		Base:       NewBase(nodeType, cmd, reflect.ValueOf(f)),
		ISerialize: en,
		method:     f,
	}
}

func (d *V2Handler[Actor, V1, V2]) Call(obj any, ctx core.IContext, args ...any) func() {
	return func() {
		startTime := time.Now().UnixMilli()

		var err error
		defer func() {
			endTime := time.Now().UnixMilli()
			if err != nil {
				ctx.Errorf("调用%s耗时%d毫秒, req:%v|%v, error:%v", d.GetName(), endTime-startTime, args[0], args[1], err)
			} else {
				ctx.Tracef("调用%s耗时%d毫秒, req:%v|%v", d.GetName(), endTime-startTime, args[0], args[1])
			}
		}()

		err = d.method(obj.(*Actor), ctx, args[0].(V1), args[1].(V2))
	}
}

func (d *V2Handler[Actor, V1, V2]) Rpc(obj any, ctx core.IContext, body []byte) func() {
	return func() {
		startTime := time.Now().UnixMilli()
		req1 := new(V1)
		req2 := new(V2)

		var err error
		defer func() {
			endTime := time.Now().UnixMilli()
			if err != nil {
				ctx.Errorf("调用%s耗时%d毫秒, req:%v|%v, error:%v", d.GetName(), endTime-startTime, *req1, *req2, err)
			} else {
				ctx.Tracef("调用%s耗时%d毫秒, req:%v|%v", d.GetName(), endTime-startTime, *req1, *req2)
			}
		}()

		// 解析参数
		if err = d.Unmarshal(body, req1, req2); err != nil {
			return
		}

		// 调用接口
		err = d.method(obj.(*Actor), ctx, *req1, *req2)
	}
}

package handler

import (
	"framework/define"
	"framework/repository/handler/internal/base"
	"reflect"
	"time"
)

type P3Handler[Actor any, V1 any, V2 any, V3 any] struct {
	*base.Base
	define.ISerialize
	method define.P3Func[Actor, V1, V2, V3]
}

func NewP3Handler[Actor any, V1 any, V2 any, V3 any](en define.ISerialize, nodeType uint32, cmd uint32, f define.P3Func[Actor, V1, V2, V3]) *P3Handler[Actor, V1, V2, V3] {
	return &P3Handler[Actor, V1, V2, V3]{
		Base:       base.NewBase(nodeType, cmd, reflect.ValueOf(f)),
		ISerialize: en,
		method:     f,
	}
}

func (d *P3Handler[Actor, V1, V2, V3]) Call(obj any, ctx define.IContext, args ...any) func() {
	return func() {
		startTime := time.Now().UnixMilli()

		var err error
		defer func() {
			endTime := time.Now().UnixMilli()
			if err != nil {
				ctx.Errorf("调用%s耗时%d毫秒, req:%v|%v|%v, error:%v", d.GetName(), endTime-startTime, args[0], args[1], args[2], err)
			} else {
				ctx.Tracef("调用%s耗时%d毫秒, req:%v|%v|%v", d.GetName(), endTime-startTime, args[0], args[1], args[2])
			}
		}()

		err = d.method(obj.(*Actor), ctx, args[0].(*V1), args[1].(*V2), args[2].(*V3))
	}
}

func (d *P3Handler[Actor, V1, V2, V3]) Rpc(obj any, ctx define.IContext, body []byte) func() {
	return func() {
		startTime := time.Now().UnixMilli()
		req1 := new(V1)
		req2 := new(V2)
		req3 := new(V3)

		var err error
		defer func() {
			endTime := time.Now().UnixMilli()
			if err != nil {
				ctx.Errorf("调用%s耗时%d毫秒, req:%v|%v|%v, error:%v", d.GetName(), endTime-startTime, *req1, *req2, *req3, err)
			} else {
				ctx.Tracef("调用%s耗时%d毫秒, req:%v|%v|%v", d.GetName(), endTime-startTime, *req1, *req2, *req3)
			}
		}()

		// 解析参数
		if err = d.Unmarshal(body, req1, req2, req3); err != nil {
			return
		}

		// 调用接口
		err = d.method(obj.(*Actor), ctx, req1, req2, req3)
	}
}

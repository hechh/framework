package handler

import (
	"framework/define"
	"framework/repository/handler/internal/base"
	"reflect"
	"time"
)

type V1Handler[Actor any, V1 any] struct {
	*base.Base
	define.ISerialize
	method define.V1Func[Actor, V1]
}

func NewV1Handler[Actor any, V1 any](en define.ISerialize, nodeType uint32, cmd uint32, f define.V1Func[Actor, V1]) *V1Handler[Actor, V1] {
	return &V1Handler[Actor, V1]{
		Base:       base.NewBase(nodeType, cmd, reflect.ValueOf(f)),
		ISerialize: en,
		method:     f,
	}
}

func (d *V1Handler[Actor, V1]) Call(obj any, ctx define.IContext, args ...any) func() {
	return func() {
		startTime := time.Now().UnixMilli()

		var err error
		defer func() {
			endTime := time.Now().UnixMilli()
			if err != nil {
				ctx.Errorf("调用%s耗时%d毫秒, req:%v, error:%v", d.GetName(), endTime-startTime, args[0], err)
			} else {
				ctx.Tracef("调用%s耗时%d毫秒, req:%v", d.GetName(), endTime-startTime, args[0])
			}
		}()

		err = d.method(obj.(*Actor), ctx, args[0].(V1))
	}
}

func (d *V1Handler[Actor, V1]) Rpc(obj any, ctx define.IContext, body []byte) func() {
	return func() {
		startTime := time.Now().UnixMilli()
		req := new(V1)

		var err error
		defer func() {
			endTime := time.Now().UnixMilli()
			if err != nil {
				ctx.Errorf("调用%s耗时%d毫秒, req:%v, error:%v", d.GetName(), endTime-startTime, *req, err)
			} else {
				ctx.Tracef("调用%s耗时%d毫秒, req:%v", d.GetName(), endTime-startTime, *req)
			}
		}()

		// 解析参数
		if err = d.Unmarshal(body, req); err != nil {
			return
		}

		// 调用接口
		err = d.method(obj.(*Actor), ctx, *req)
	}
}

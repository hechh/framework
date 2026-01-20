package entity

import (
	"reflect"
	"time"

	"github.com/hechh/framework"
)

type V2Handler[Actor any, V any, R any] struct {
	*Base
	f framework.V2Func[Actor, V, R]
}

func NewV2Handler[Actor any, V any, R any](en framework.ISerialize, f framework.V2Func[Actor, V, R]) *V2Handler[Actor, V, R] {
	return &V2Handler[Actor, V, R]{
		Base: NewBase(en, "", reflect.ValueOf(f)),
		f:    f,
	}
}

func (d *V2Handler[Actor, V, R]) Call(obj any, ctx framework.IContext, args ...any) func() {
	return func() {
		startTime := time.Now().UnixMilli()

		var err error
		defer func() {
			endTime := time.Now().UnixMilli()
			if err != nil {
				ctx.Errorf("调用%s耗时%d毫秒, arg1:%v, arg2:%v, error:%v", d.GetName(), endTime-startTime, args[0], args[1], err)
			} else {
				ctx.Tracef("调用%s耗时%d毫秒, arg1:%v, arg2:%v", d.GetName(), endTime-startTime, args[0], args[1])
			}
		}()

		err = d.f(obj.(*Actor), ctx, args[0].(V), args[1].(R))
	}
}

func (d *V2Handler[Actor, V, R]) Rpc(obj any, ctx framework.IContext, body []byte) func() {
	return func() {
		startTime := time.Now().UnixMilli()
		req1 := new(V)
		req2 := new(R)

		var err error
		defer func() {
			endTime := time.Now().UnixMilli()
			if err != nil {
				ctx.Errorf("调用%s耗时%d毫秒, arg1:%v, arg2:%v, error:%v", d.GetName(), endTime-startTime, *req1, *req2, err)
			} else {
				ctx.Tracef("调用%s耗时%d毫秒, arg1:%v, arg2:%v", d.GetName(), endTime-startTime, *req1, *req2)
			}
		}()

		// 解析参数
		if err = d.Unmarshal(body, req1, req2); err != nil {
			return
		}

		// 调用接口
		err = d.f(obj.(*Actor), ctx, *req1, *req2)
	}
}

package entity

import (
	"reflect"
	"time"

	"github.com/hechh/framework"
	"github.com/hechh/library/mlog"
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
				mlog.Error(-1, "[result] 调用%s耗时%d毫秒, error:%v, head:%v, arg1:%v, arg2:%v", d.GetName(), endTime-startTime, err, ctx.GetHead(), args[0], args[1])
			} else {
				mlog.Trace(-1, "[result] 调用%s耗时%d毫秒, head:%v, arg1:%v, arg2:%v", d.GetName(), endTime-startTime, ctx.GetHead(), args[0], args[1])
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
				mlog.Error(-1, "[result] 调用%s耗时%d毫秒, error:%v, head:%v, arg1:%v, arg2:%v", d.GetName(), endTime-startTime, err, ctx.GetHead(), *req1, *req2)
			} else {
				mlog.Trace(-1, "[result] 调用%s耗时%d毫秒, head:%v, arg1:%v, arg2:%v", d.GetName(), endTime-startTime, ctx.GetHead(), *req1, *req2)
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

package entity

import (
	"reflect"
	"time"

	"github.com/hechh/framework"
)

type P2Handler[Actor any, V1 any, V2 any] struct {
	*Base
	method framework.P2Func[Actor, V1, V2]
}

func NewP2Handler[Actor any, V1 any, V2 any](en framework.ISerialize, f framework.P2Func[Actor, V1, V2]) *P2Handler[Actor, V1, V2] {
	return &P2Handler[Actor, V1, V2]{
		Base:   NewBase(en, "", reflect.ValueOf(f)),
		method: f,
	}
}

func (d *P2Handler[Actor, V1, V2]) Call(obj any, ctx framework.IContext, args ...any) func() {
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

		err = d.method(obj.(*Actor), ctx, args[0].(*V1), args[1].(*V2))
	}
}

func (d *P2Handler[Actor, V1, V2]) Rpc(obj any, ctx framework.IContext, body []byte) func() {
	return func() {
		startTime := time.Now().UnixMilli()
		req1 := new(V1)
		req2 := new(V2)

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
		err = d.method(obj.(*Actor), ctx, req1, req2)
	}
}

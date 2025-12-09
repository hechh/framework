package handler

import (
	"framework/define"
	"framework/repository/handler/domain"
	"framework/repository/handler/internal/base"
	"reflect"
	"time"
)

type P1Handler[Actor any, V1 any] struct {
	*base.Base
	define.ISerialize
	method domain.P1Func[Actor, V1]
}

func NewP1Handler[Actor any, V1 any](en define.ISerialize, nodeType uint32, cmd uint32, f domain.P1Func[Actor, V1]) *P1Handler[Actor, V1] {
	return &P1Handler[Actor, V1]{
		Base:       base.NewBase(nodeType, cmd, reflect.ValueOf(f)),
		ISerialize: en,
		method:     f,
	}
}

func (d *P1Handler[Actor, V1]) Call(obj any, ctx define.IContext, args ...any) func() {
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

		err = d.method(obj.(*Actor), ctx, args[0].(*V1))
	}
}

func (d *P1Handler[Actor, V1]) Rpc(obj any, ctx define.IContext, body []byte) func() {
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
		err = d.method(obj.(*Actor), ctx, req)
	}
}

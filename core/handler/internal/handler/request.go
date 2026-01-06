package handler

import (
	"framework/core/bus"
	"framework/core/define"
	"reflect"
	"time"
)

type ReqHandler[Actor any, V1 any, V2 any] struct {
	*Base
	define.ISerialize
	method define.P2Func[Actor, V1, V2]
}

func NewReqHandler[Actor any, V1 any, V2 any](en define.ISerialize, nodeType uint32, cmd uint32, f define.P2Func[Actor, V1, V2]) *ReqHandler[Actor, V1, V2] {
	return &ReqHandler[Actor, V1, V2]{
		Base:       NewBase(nodeType, cmd, reflect.ValueOf(f)),
		ISerialize: en,
		method:     f,
	}
}

func (d *ReqHandler[Actor, V1, V2]) Call(obj any, ctx define.IContext, args ...any) func() {
	ref := ctx.AddDepth(1)
	return func() {
		startTime := time.Now().UnixMilli()

		var reterr, err error
		defer func() {
			endTime := time.Now().UnixMilli()
			if err != nil {
				ctx.Errorf("调用%s耗时%d毫秒, req:%v|%v, error:%v, reterr:%v", d.GetName(), endTime-startTime, args[0], args[1], err, reterr)
			} else {
				ctx.Tracef("调用%s耗时%d毫秒, req:%v|%v, reterr:%v", d.GetName(), endTime-startTime, args[0], args[1], reterr)
			}
		}()

		err = d.method(obj.(*Actor), ctx, args[0].(*V1), args[1].(*V2))

		// 应答
		if ctx.CompareAndSwapDepth(ref, ref) && ctx.IsRsp() {
			reterr = bus.Send(ctx.GetPacket().Rsp(define.GATE, err, args[1].(define.IRspHead)))
		}
	}
}

func (d *ReqHandler[Actor, V1, V2]) Rpc(obj any, ctx define.IContext, body []byte) func() {
	ref := ctx.AddDepth(1)
	return func() {
		startTime := time.Now().UnixMilli()
		req1 := new(V1)
		req2 := new(V2)

		var reterr, err error
		defer func() {
			endTime := time.Now().UnixMilli()
			if err != nil {
				ctx.Errorf("调用%s耗时%d毫秒, req:%v|%v, error:%v, reterr:%v", d.GetName(), endTime-startTime, *req1, *req2, err, reterr)
			} else {
				ctx.Tracef("调用%s耗时%d毫秒, req:%v|%v, reterr:%v", d.GetName(), endTime-startTime, *req1, *req2, reterr)
			}
		}()

		// 解析参数
		if err = d.Unmarshal(body, req1, req2); err != nil {
			return
		}

		// 调用接口
		err = d.method(obj.(*Actor), ctx, req1, req2)

		// 应答
		if ctx.CompareAndSwapDepth(ref, ref) && ctx.IsRsp() {
			reterr = bus.Send(ctx.GetPacket().Rsp(define.GATE, err, any(req2).(define.IRspHead)))
		}
	}
}

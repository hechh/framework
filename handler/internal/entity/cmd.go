package entity

import (
	"reflect"
	"time"

	"github.com/hechh/framework"
)

type CmdHandler[Actor any, V1 any, V2 any] struct {
	*Base
	method framework.P2Func[Actor, V1, V2]
}

func NewCmdHandler[Actor any, V1 any, V2 any](en framework.ISerialize, f framework.P2Func[Actor, V1, V2]) *CmdHandler[Actor, V1, V2] {
	return &CmdHandler[Actor, V1, V2]{
		Base:   NewBase(en, "", reflect.ValueOf(f)),
		method: f,
	}
}

func (d *CmdHandler[Actor, V1, V2]) Call(obj any, ctx framework.IContext, args ...any) func() {
	ref := ctx.AddDepth(1)
	return func() {
		startTime := time.Now().UnixMilli()

		var reterr, err error
		defer func() {
			endTime := time.Now().UnixMilli()
			if err != nil {
				ctx.Errorf("[result] 调用%s耗时%d毫秒, error:%v, reterr:%v, head:%v, req:%v, rsp:%v", d.GetName(), endTime-startTime, err, reterr, ctx.GetHead(), args[0], args[1])
			} else {
				ctx.Tracef("[result] 调用%s耗时%d毫秒, reterr:%v, req:%v, head:%v, rsp:%v", d.GetName(), endTime-startTime, reterr, ctx.GetHead(), args[0], args[1])
			}
		}()

		err = d.method(obj.(*Actor), ctx, args[0].(*V1), args[1].(*V2))

		// 应答
		if ctx.CompareAndSwapDepth(ref, ref) {
			if rsp, ok := args[1].(framework.IResponse); err != nil && ok && rsp != nil {
				rsp.SetRspHead(framework.ToRspHead(err))
			}
			reterr = framework.SendResponse(ctx, framework.Rsp(d, nil, args[1]))
		}
	}
}

func (d *CmdHandler[Actor, V1, V2]) Rpc(obj any, ctx framework.IContext, body []byte) func() {
	ref := ctx.AddDepth(1)
	return func() {
		startTime := time.Now().UnixMilli()
		req1 := new(V1)
		req2 := new(V2)

		var reterr, err error
		defer func() {
			endTime := time.Now().UnixMilli()
			if err != nil {
				ctx.Errorf("[result] 调用%s耗时%d毫秒, error:%v, reterr:%v, head:%v, req:%v, rsp:%v", d.GetName(), endTime-startTime, err, reterr, ctx.GetHead(), *req1, *req2)
			} else {
				ctx.Tracef("[result] 调用%s耗时%d毫秒, reterr:%v, req:%v, head:%v, rsp:%v", d.GetName(), endTime-startTime, reterr, ctx.GetHead(), *req1, *req2)
			}
		}()

		// 解析参数
		if err = d.Unmarshal(body, req1, req2); err != nil {
			return
		}

		// 调用接口
		err = d.method(obj.(*Actor), ctx, req1, req2)

		// 应答
		if ctx.CompareAndSwapDepth(ref, ref) {
			if rsp, ok := any(req2).(framework.IResponse); err != nil && ok && rsp != nil {
				rsp.SetRspHead(framework.ToRspHead(err))
			}
			reterr = framework.SendResponse(ctx, framework.Rsp(d, nil, req2))
		}
	}
}

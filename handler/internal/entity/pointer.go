package entity

import (
	"reflect"
	"time"

	"github.com/hechh/framework"
	"google.golang.org/protobuf/proto"
)

var (
	pbType  = reflect.TypeOf((*proto.Message)(nil)).Elem()
	rspType = reflect.TypeOf((*framework.IResponse)(nil)).Elem()
)

type P1Handler[Actor any, N any] struct {
	framework.ISerialize
	name string
	id   uint32
	fun  framework.P1Func[Actor, N]
}

func NewP1Handler[Actor any, N any](en framework.ISerialize, f framework.P1Func[Actor, N]) *P1Handler[Actor, N] {
	name := framework.ParseActorFunc(reflect.ValueOf(f))
	return &P1Handler[Actor, N]{
		ISerialize: en,
		name:       name,
		id:         framework.GetCrc32(name),
		fun:        f,
	}
}

func (d *P1Handler[Actor, N]) GetName() string {
	return d.name
}

func (d *P1Handler[Actor, N]) GetCrc32() uint32 {
	return d.id
}

func (d *P1Handler[Actor, N]) Call(obj any, ctx framework.IContext, args ...any) func() {
	return func() {
		// 调用接口
		startTime := time.Now().UnixMilli()
		err := d.fun(obj.(*Actor), ctx, args[0].(*N))

		// 输出
		endTime := time.Now().UnixMilli()
		if err != nil {
			ctx.Errorf("[%s] %dms, arg:%v, error:%v", d.GetName(), endTime-startTime, args[0], err)
		} else {
			ctx.Tracef("[%s] %dms, arg:%v", d.GetName(), endTime-startTime, args[0])
		}
	}
}

func (d *P1Handler[Actor, N]) Rpc(obj any, ctx framework.IContext, body []byte) func() {
	return func() {
		// 解析参数
		startTime := time.Now().UnixMilli()
		req := new(N)
		err := d.Unmarshal(body, req)

		// 调用接口
		if err == nil {
			err = d.fun(obj.(*Actor), ctx, req)
		}

		// 输出
		endTime := time.Now().UnixMilli()
		if err != nil {
			ctx.Errorf("[%s] %dms, arg:%v, error:%v", d.GetName(), endTime-startTime, *req, err)
		} else {
			ctx.Tracef("[%s] %dms, arg:%v", d.GetName(), endTime-startTime, *req)
		}
	}
}

type P2Handler[Actor any, N any, R any] struct {
	framework.ISerialize
	name  string
	id    uint32
	fun   framework.P2Func[Actor, N, R]
	iscmd bool
}

func NewP2Handler[Actor any, N any, R any](en framework.ISerialize, f framework.P2Func[Actor, N, R]) *P2Handler[Actor, N, R] {
	name := framework.ParseActorFunc(reflect.ValueOf(f))
	nType := reflect.TypeOf((*N)(nil)).Elem()
	rType := reflect.TypeOf((*R)(nil)).Elem()
	return &P2Handler[Actor, N, R]{
		ISerialize: en,
		name:       name,
		id:         framework.GetCrc32(name),
		fun:        f,
		iscmd:      nType.Implements(pbType) && rType.Implements(rspType),
	}
}

func (d *P2Handler[Actor, N, R]) GetName() string {
	return d.name
}

func (d *P2Handler[Actor, N, R]) GetCrc32() uint32 {
	return d.id
}

func (d *P2Handler[Actor, N, R]) Call(obj any, ctx framework.IContext, args ...any) func() {
	ref := int32(0)
	if d.iscmd {
		ref = ctx.AddDepth(1)
	}
	return func() {
		// 调用接口
		startTime := time.Now().UnixMilli()
		arg1 := args[0].(*N)
		arg2 := args[1].(*R)
		err := d.fun(obj.(*Actor), ctx, args[0].(*N), args[1].(*R))

		// 自动回复
		if d.iscmd && ctx.CompareAndSwapDepth(ref, ref) {
			if rsp, ok := any(arg2).(framework.IResponse); err != nil && ok && rsp != nil {
				rsp.SetRspHead(framework.ToRspHead(err))
			}
			if reterr := framework.SendResponse(ctx, framework.Rsp(d, nil, arg2)); reterr != nil {
				ctx.Errorf("[自动回复] rsp:%v, error:%v", arg2, reterr)
			}
		}

		// 输出
		endTime := time.Now().UnixMilli()
		if err != nil {
			ctx.Errorf("[%s] %dms, arg1:%v, arg2:%v, error:%v", d.GetName(), endTime-startTime, *arg1, *arg2, err)
		} else {
			ctx.Tracef("[%s] %dms, arg1:%v, arg2:%v", d.GetName(), endTime-startTime, *arg1, *arg2)
		}
	}
}

func (d *P2Handler[Actor, N, R]) Rpc(obj any, ctx framework.IContext, body []byte) func() {
	ref := int32(0)
	if d.iscmd {
		ref = ctx.AddDepth(1)
	}
	return func() {
		// 解析参数
		startTime := time.Now().UnixMilli()
		arg1 := new(N)
		arg2 := new(R)
		err := d.Unmarshal(body, arg1, arg2)

		// 调用接口
		if err == nil {
			err = d.fun(obj.(*Actor), ctx, arg1, arg2)
		}

		// 自动回复
		if d.iscmd && ctx.CompareAndSwapDepth(ref, ref) {
			if rsp, ok := any(arg2).(framework.IResponse); err != nil && ok && rsp != nil {
				rsp.SetRspHead(framework.ToRspHead(err))
			}
			if reterr := framework.SendResponse(ctx, framework.Rsp(d, nil, arg2)); reterr != nil {
				ctx.Errorf("[自动回复] rsp:%v, error:%v", arg2, reterr)
			}
		}

		// 输出
		endTime := time.Now().UnixMilli()
		if err != nil {
			ctx.Errorf("[%s] %dms, arg1:%v, arg2:%v, error:%v", d.GetName(), endTime-startTime, *arg1, *arg2, err)
		} else {
			ctx.Tracef("[%s] %dms, arg1:%v, arg2:%v", d.GetName(), endTime-startTime, *arg1, *arg2)
		}
	}
}

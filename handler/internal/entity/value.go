package entity

import (
	"reflect"
	"time"

	"github.com/hechh/framework"
)

type V1Handler[Actor any, N any] struct {
	framework.ISerialize
	name string
	id   uint32
	fun  framework.V1Func[Actor, N]
}

func NewV1Handler[Actor any, N any](en framework.ISerialize, f framework.V1Func[Actor, N]) *V1Handler[Actor, N] {
	name := framework.ParseActorFunc(reflect.ValueOf(f))
	return &V1Handler[Actor, N]{
		ISerialize: en,
		name:       name,
		id:         framework.GetCrc32(name),
		fun:        f,
	}
}

func (d *V1Handler[Actor, N]) GetName() string {
	return d.name
}

func (d *V1Handler[Actor, N]) GetCrc32() uint32 {
	return d.id
}

func (d *V1Handler[Actor, N]) Call(obj any, ctx framework.IContext, args ...any) func() {
	return func() {
		// 调用接口
		startTime := time.Now().UnixMilli()
		arg1 := args[0].(N)
		err := d.fun(obj.(*Actor), ctx, arg1)
		// 输出
		endTime := time.Now().UnixMilli()
		if err != nil {
			ctx.Errorf("[%s] %dms, arg:%v, error:%v", d.GetName(), endTime-startTime, args[0], err)
		} else {
			ctx.Tracef("[%s] %dms, arg:%v", d.GetName(), endTime-startTime, args[0])
		}
	}
}

func (d *V1Handler[Actor, N]) Rpc(obj any, ctx framework.IContext, body []byte) func() {
	return func() {
		// 解析参数
		startTime := time.Now().UnixMilli()
		req := new(N)
		err := d.Unmarshal(body, req)
		// 调用接口
		if err == nil {
			err = d.fun(obj.(*Actor), ctx, *req)
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

type V2Handler[Actor any, V any, R any] struct {
	framework.ISerialize
	name string
	id   uint32
	fun  framework.V2Func[Actor, V, R]
}

func NewV2Handler[Actor any, V any, R any](en framework.ISerialize, f framework.V2Func[Actor, V, R]) *V2Handler[Actor, V, R] {
	name := framework.ParseActorFunc(reflect.ValueOf(f))
	return &V2Handler[Actor, V, R]{
		ISerialize: en,
		name:       name,
		id:         framework.GetCrc32(name),
		fun:        f,
	}
}

func (d *V2Handler[Actor, V, R]) GetName() string {
	return d.name
}

func (d *V2Handler[Actor, V, R]) GetCrc32() uint32 {
	return d.id
}

func (d *V2Handler[Actor, V, R]) Call(obj any, ctx framework.IContext, args ...any) func() {
	return func() {
		// 调用接口
		startTime := time.Now().UnixMilli()
		arg1 := args[0].(V)
		arg2 := args[1].(R)
		err := d.fun(obj.(*Actor), ctx, arg1, arg2)
		// 输出
		endTime := time.Now().UnixMilli()
		if err != nil {
			ctx.Errorf("[%s] %dms, arg1:%v, arg2:%v, error:%v", d.GetName(), endTime-startTime, args[0], args[1], err)
		} else {
			ctx.Tracef("[%s] %dms, arg1:%v, arg2:%v", d.GetName(), endTime-startTime, args[0], args[1])
		}
	}
}

func (d *V2Handler[Actor, V, R]) Rpc(obj any, ctx framework.IContext, body []byte) func() {
	return func() {
		// 解析参数
		startTime := time.Now().UnixMilli()
		req1 := new(V)
		req2 := new(R)
		err := d.Unmarshal(body, req1, req2)
		// 调用接口
		if err == nil {
			err = d.fun(obj.(*Actor), ctx, *req1, *req2)
		}
		// 输出
		endTime := time.Now().UnixMilli()
		if err != nil {
			ctx.Errorf("[%s] %dms, arg1:%v, arg2:%v, error:%v", d.GetName(), endTime-startTime, *req1, *req2, err)
		} else {
			ctx.Tracef("[%s] %dms, arg1:%v, arg2:%v", d.GetName(), endTime-startTime, *req1, *req2)
		}
	}
}

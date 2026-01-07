package handler

import (
	"framework/core"
	"reflect"
	"time"
)

type V0Handler[Actor any] struct {
	*Base
	EmptyEncoder
	method core.V0Func[Actor]
}

func NewV0Handler[Actor any](nodeType uint32, cmd uint32, f core.V0Func[Actor]) *V0Handler[Actor] {
	return &V0Handler[Actor]{
		Base:   NewBase(nodeType, cmd, reflect.ValueOf(f)),
		method: f,
	}
}

func (d *V0Handler[Actor]) Call(obj any, ctx core.IContext, args ...any) func() {
	return func() {
		var err error
		startTime := time.Now().UnixMilli()
		defer func() {
			endTime := time.Now().UnixMilli()
			if err != nil {
				ctx.Errorf("调用%s耗时%d毫秒, error:%v", d.GetName(), endTime-startTime, err)
			} else {
				ctx.Tracef("调用%s耗时%d毫秒", d.GetName(), endTime-startTime)
			}
		}()

		err = d.method(obj.(*Actor), ctx)
	}
}

func (d *V0Handler[Actor]) Rpc(obj any, ctx core.IContext, body []byte) func() {
	return func() {
		var err error
		startTime := time.Now().UnixMilli()
		defer func() {
			endTime := time.Now().UnixMilli()
			if err != nil {
				ctx.Errorf("调用%s耗时%d毫秒, head:%v, error:%v", d.GetName(), endTime-startTime, err)
			} else {
				ctx.Tracef("调用%s耗时%d毫秒, head:%v", d.GetName(), endTime-startTime)
			}
		}()

		err = d.method(obj.(*Actor), ctx)
	}
}

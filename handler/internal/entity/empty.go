package entity

import (
	"reflect"
	"time"

	"github.com/hechh/framework"
)

type EmptyHandler[Actor any] struct {
	framework.ISerialize
	name string
	id   uint32
	fun  framework.EmptyFunc[Actor]
}

func NewEmptyHandler[Actor any](en framework.ISerialize, f framework.EmptyFunc[Actor]) *EmptyHandler[Actor] {
	name := framework.ParseActorFunc(reflect.ValueOf(f))
	return &EmptyHandler[Actor]{
		ISerialize: en,
		name:       name,
		id:         framework.GetCrc32(name),
		fun:        f,
	}
}

func (d *EmptyHandler[Actor]) GetName() string {
	return d.name
}

func (d *EmptyHandler[Actor]) GetCrc32() uint32 {
	return d.id
}

func (d *EmptyHandler[Actor]) Call(obj any, ctx framework.IContext, args ...any) func() {
	return func() {
		startTime := time.Now().UnixMilli()
		err := d.fun(obj.(*Actor), ctx)
		endTime := time.Now().UnixMilli()
		if err != nil {
			ctx.Errorf("[%s] %dms, error:%v", d.GetName(), endTime-startTime, err)
		} else {
			ctx.Tracef("[%s] %dms", d.GetName(), endTime-startTime)
		}
	}
}

func (d *EmptyHandler[Actor]) Rpc(obj any, ctx framework.IContext, body []byte) func() {
	return d.Call(obj, ctx)
}

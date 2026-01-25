package entity

import (
	"reflect"
	"strings"
	"time"

	"github.com/hechh/framework"
	"github.com/hechh/library/mlog"
)

type EmptyHandler[Actor any] struct {
	*Base
	framework.ISerialize
	method framework.EmptyFunc[Actor]
}

func NewV0Handler[Actor any](en framework.ISerialize, f framework.EmptyFunc[Actor]) *EmptyHandler[Actor] {
	return &EmptyHandler[Actor]{
		Base:   NewBase(en, "", reflect.ValueOf(f)),
		method: f,
	}
}

func (d *EmptyHandler[Actor]) Call(obj any, ctx framework.IContext, args ...any) func() {
	return func() {
		var err error
		startTime := time.Now().UnixMilli()
		defer func() {
			endTime := time.Now().UnixMilli()
			if !strings.HasSuffix(d.name, "OnTick") {
				if err != nil {
					mlog.Error(-1, "调用%s耗时%d毫秒, error:%v, head:%v", d.GetName(), endTime-startTime, err, ctx.GetHead())
				} else {
					mlog.Trace(-1, "调用%s耗时%d毫秒, head:%v", d.GetName(), endTime-startTime, ctx.GetHead())
				}
			}
		}()

		err = d.method(obj.(*Actor), ctx)
	}
}

func (d *EmptyHandler[Actor]) Rpc(obj any, ctx framework.IContext, body []byte) func() {
	return func() {
		var err error
		startTime := time.Now().UnixMilli()
		defer func() {
			endTime := time.Now().UnixMilli()
			if !strings.HasSuffix(d.name, "OnTick") {
				if err != nil {
					mlog.Error(-1, "调用%s耗时%d毫秒, error:%v, head:%v", d.GetName(), endTime-startTime, err, ctx.GetHead())
				} else {
					mlog.Trace(-1, "调用%s耗时%d毫秒, head:%v", d.GetName(), endTime-startTime, ctx.GetHead())
				}
			}
		}()

		err = d.method(obj.(*Actor), ctx)
	}
}

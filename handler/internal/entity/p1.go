package entity

import (
	"reflect"
	"time"

	"github.com/hechh/framework"
	"github.com/hechh/library/mlog"
)

type P1Handler[Actor any, V1 any] struct {
	*Base
	method framework.P1Func[Actor, V1]
}

func NewP1Handler[Actor any, V1 any](en framework.ISerialize, f framework.P1Func[Actor, V1]) *P1Handler[Actor, V1] {
	return &P1Handler[Actor, V1]{
		Base:   NewBase(en, "", reflect.ValueOf(f)),
		method: f,
	}
}

func (d *P1Handler[Actor, V1]) Call(obj any, ctx framework.IContext, args ...any) func() {
	return func() {
		startTime := time.Now().UnixMilli()

		var err error
		defer func() {
			endTime := time.Now().UnixMilli()
			if err != nil {
				mlog.Errorf("[result] 调用%s耗时%d毫秒, error:%v, head:%v, req:%v", d.GetName(), endTime-startTime, err, ctx.GetHead(), args[0])
			} else {
				mlog.Tracef("[result] 调用%s耗时%d毫秒, head:%v, req:%v", d.GetName(), endTime-startTime, ctx.GetHead(), args[0])
			}
		}()

		err = d.method(obj.(*Actor), ctx, args[0].(*V1))
	}
}

func (d *P1Handler[Actor, V1]) Rpc(obj any, ctx framework.IContext, body []byte) func() {
	return func() {
		startTime := time.Now().UnixMilli()
		req := new(V1)

		var err error
		defer func() {
			endTime := time.Now().UnixMilli()
			if err != nil {
				mlog.Errorf("[result] 调用%s耗时%d毫秒, error:%v, head:%v, req:%v", d.GetName(), endTime-startTime, err, ctx.GetHead(), *req)
			} else {
				mlog.Tracef("[result] 调用%s耗时%d毫秒, head:%v, req:%v", d.GetName(), endTime-startTime, ctx.GetHead(), *req)
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

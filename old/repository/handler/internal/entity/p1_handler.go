package entity

import (
	"framework/define"
	"framework/internal/bus"
	"framework/internal/sender"
	"framework/library/uerror"
	"framework/repository/handler/domain"
	"reflect"
	"time"
)

type P1Handler[Actor any, P1 any] struct {
	*Common
	ProtoCrypto
	method domain.P1Func[Actor, P1]
}

func NewP1Handler[Actor any, P1 any](nodeType int32, cmd int32, method domain.P1Func[Actor, P1], rsp domain.RspFunc) *P1Handler[Actor, P1] {
	return &P1Handler[Actor, P1]{
		Common: NewCommon(nodeType, cmd, ParseActorFunc(reflect.ValueOf(method)), rsp),
		method: method,
	}
}

func (d *P1Handler[Actor, P1]) Call(obj any, head define.IContext, args ...any) func() {
	ref := head.AddDepth(1)
	return func() {
		// 参数解析
		req, ok := args[0].(*P1)
		if !ok {
			head.Errorf("参数类型错误")
			return
		}
		actor, ok := any(obj).(*Actor)
		if !ok {
			head.Errorf("参数类型错误")
			return
		}

		// 接口调用
		startMs := time.Now().UnixMilli()
		err := d.method(actor, head, req)
		endMs := time.Now().UnixMilli()
		if err != nil {
			head.Errorf("接口耗时%d毫秒\treq:%v\terror:%v", endMs-startMs, *req, err)
		} else {
			head.Tracef("接口耗时%d毫秒\treq:%v", endMs-startMs, *req)
		}

		// 是否自动回复
		if head.CompareAndSwapDepth(ref, ref) {
			if rsp, ok := any(req).(define.IRspHead); ok && err != nil {
				uerr := uerror.ToUError(err)
				rsp.SetRspHead(uerr.GetCode(), uerr.GetMsg())
			}
			err := bus.Response(head.GetHead(), req)
			if err != nil {
				head.Errorf("自动回复失败\trsp:%v\terror:%v", *req, err)
			} else {
				head.Tracef("自动回复成功\trsp:%v", *req)
			}
		}
	}
}

func (d *P1Handler[Actor, P1]) Rpc(obj any, head define.IContext, body []byte) func() {
	ref := head.AddDepth(1)
	return func() {
		// 参数解析
		req := new(P1)
		if err := d.Unmarshal(body, req); err != nil {
			head.Errorf("参数类型错误:%v", err)
			return
		}
		actor, ok := any(obj).(*Actor)
		if !ok {
			head.Errorf("参数类型错误")
			return
		}

		// 接口调用
		startMs := time.Now().UnixMilli()
		err := d.method(actor, head, req)
		endMs := time.Now().UnixMilli()
		if err != nil {
			head.Errorf("接口耗时%d毫秒\treq:%verror:%v", endMs-startMs, *req, err)
		} else {
			head.Tracef("接口耗时%d毫秒\treq:%v", endMs-startMs, *req)
		}

		// 是否自动回复
		if head.CompareAndSwapDepth(ref, ref) {
			if rsp, ok := any(req).(define.IRspHead); ok && err != nil {
				uerr := uerror.ToUError(err)
				rsp.SetRspHead(uerr.GetCode(), uerr.GetMsg())
			}
			err := sender.Response(head.GetHead(), req)
			if err != nil {
				head.Errorf("自动回复失败\trsp:%v\terror:%v", *req, err)
			} else {
				head.Tracef("自动回复成功\trsp:%v", *req)
			}
		}
	}
}

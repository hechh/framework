package entity

import (
	"framework/define"
	"framework/internal/sender"
	"framework/library/uerror"
	"framework/repository/handler/domain"
	"reflect"
	"time"

	"github.com/golang/protobuf/proto"
)

type P2Handler[Actor any, P1 any, P2 any] struct {
	*Common
	ProtoCrypto
	method domain.P2Func[Actor, P1, P2]
}

func NewP2Handler[Actor any, P1 any, P2 any](nodeType int32, cmd int32, method domain.P2Func[Actor, P1, P2]) *P2Handler[Actor, P1, P2] {
	return &P2Handler[Actor, P1, P2]{
		Common: NewCommon(nodeType, cmd, ParseActorFunc(reflect.ValueOf(method))),
		method: method,
	}
}

func (d *P2Handler[Actor, P1, P2]) Marshal(args ...any) ([]byte, error) {
	return proto.Marshal(args[0].(proto.Message))
}

func (d *P2Handler[Actor, P1, P2]) Unmarshal(buf []byte, args ...any) error {
	return proto.Unmarshal(buf, args[0].(proto.Message))
}

func (d *P2Handler[Actor, P1, P2]) Call(obj any, head define.IContext, args ...any) func() {
	ref := head.AddDepth(1)
	return func() {
		// 参数解析
		rsp := new(P2)
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
		err := d.method(actor, head, req, rsp)
		endMs := time.Now().UnixMilli()
		if err != nil {
			head.Errorf("接口耗时%d毫秒\treq:%v\trsp:%v\terror:%v", endMs-startMs, req, rsp, err)
		} else {
			head.Tracef("接口耗时%d毫秒\treq:%v\trsp:%v", endMs-startMs, req, rsp)
		}

		// 是否自动回复
		if head.CompareAndSwapDepth(ref, ref) {
			if rsp, ok := any(rsp).(define.IRspHead); ok && err != nil {
				uerr := uerror.ToUError(err)
				rsp.SetRspHead(uerr.GetCode(), uerr.GetMsg())
			}
			err := sender.Response(head.GetHead(), rsp)
			if err != nil {
				head.Errorf("自动回复失败\trsp:%v\terror:%v", rsp, err)
			} else {
				head.Tracef("自动回复成功\trsp:%v", rsp)
			}
		}
	}
}

func (d *P2Handler[Actor, P1, P2]) Rpc(obj any, head define.IContext, body []byte) func() {
	ref := head.AddDepth(1)
	return func() {
		// 参数解析
		rsp := new(P2)
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
		err := d.method(actor, head, req, rsp)
		endMs := time.Now().UnixMilli()
		startMs := time.Now().UnixMilli()
		if err != nil {
			head.Errorf("接口耗时%d毫秒\treq:%v\trsp:%v\terror:%v", endMs-startMs, req, rsp, err)
		} else {
			head.Tracef("接口耗时%d毫秒\treq:%v\trsp:%v", endMs-startMs, req, rsp)
		}

		// 是否自动回复
		if head.CompareAndSwapDepth(ref, ref) {
			if rsp, ok := any(rsp).(define.IRspHead); ok && err != nil {
				uerr := uerror.ToUError(err)
				rsp.SetRspHead(uerr.GetCode(), uerr.GetMsg())
			}
			err := sender.Response(head.GetHead(), rsp)
			if err != nil {
				head.Errorf("自动回复失败\trsp:%v\terror:%v", rsp, err)
			} else {
				head.Tracef("自动回复成功\trsp:%v", rsp)
			}
		}
	}
}

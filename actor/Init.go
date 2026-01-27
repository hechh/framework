package actor

import (
	"github.com/hechh/framework"
	"github.com/hechh/framework/context"
	"github.com/hechh/library/mlog"
	"github.com/hechh/library/uerror"
)

var (
	mapActor = make(map[string]framework.IActor)
)

func Register(act framework.IActor) {
	mapActor[act.GetActorName()] = act
}

func Send(ctx framework.IContext, body []byte) error {
	var err error
	if act, ok := mapActor[ctx.GetActorName()]; ok {
		err = act.Send(ctx, body)
	} else {
		err = uerror.Err(-1, "%s未注册", ctx.GetActorFunc())
	}
	mlog.Trace(-1, "[actor] 远程调用%s接口 head:%v, error:%v, body:%v", ctx.GetActorFunc(), ctx.GetHead(), err, body)
	return err
}

func SendMsg(ctx framework.IContext, args ...any) error {
	var err error
	if act, ok := mapActor[ctx.GetActorName()]; ok {
		err = act.SendMsg(ctx, args...)
	} else {
		err = uerror.Err(-1, "%s未注册", ctx.GetActorFunc())
	}
	mlog.Trace(-1, "[actor] 本地调用%s接口 head:%v, error:%v, args:%v", ctx.GetActorFunc(), ctx.GetHead(), err, args)
	return err
}

func SendTo(ctx framework.IContext, name string, buf []byte) error {
	ctx.To(name)
	return Send(ctx, buf)
}

func SendMsgTo(ctx framework.IContext, name string, args ...any) error {
	ctx.To(name)
	return SendMsg(ctx, args...)
}

func SendSimple(aid uint64, name string, body []byte) error {
	return Send(context.NewSimpleContext(aid, name), body)
}

func SendMsgSimple(aid uint64, name string, args ...any) error {
	return SendMsg(context.NewSimpleContext(aid, name), args...)
}

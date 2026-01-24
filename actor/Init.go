package actor

import (
	"github.com/hechh/framework"
	"github.com/hechh/framework/context"
	"github.com/hechh/library/uerror"
)

var (
	mapActor = make(map[string]framework.IActor)
)

func Register(act framework.IActor) {
	mapActor[act.GetActorName()] = act
}

func Send(ctx framework.IContext, body []byte) error {
	if act, ok := mapActor[ctx.GetActorName()]; ok {
		return act.Send(ctx, body)
	}
	return uerror.New(-1, "%s未注册", ctx.GetActorFunc())
}

func SendMsg(ctx framework.IContext, args ...any) error {
	if act, ok := mapActor[ctx.GetActorName()]; ok {
		return act.SendMsg(ctx, args...)
	}
	return uerror.New(-1, "%s未注册", ctx.GetActorFunc())
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

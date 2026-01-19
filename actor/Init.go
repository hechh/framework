package actor

import (
	"github.com/hechh/framework"
	"github.com/hechh/framework/context"
	"github.com/hechh/framework/packet"
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

func SendSimple(head *packet.Head, name string, buf []byte) error {
	return Send(context.NewContext(head, name), buf)
}

func SendMsgSimple(uid uint64, name string) error {
	return SendMsg(context.NewSimpleContext(uid, name), name)
}

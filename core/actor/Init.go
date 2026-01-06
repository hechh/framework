package actor

import (
	"framework/core/define"
	"framework/library/uerror"
)

var (
	mapActor = make(map[string]define.IActor)
)

func Register(act define.IActor) {
	mapActor[act.GetActorName()] = act
}

func SendMsg(ctx define.IContext, args ...any) error {
	if act, ok := mapActor[ctx.GetActorName()]; ok {
		return act.SendMsg(ctx, args...)
	}
	return uerror.New(-1, "%s.%s未注册", ctx.GetActorName(), ctx.GetFuncName())
}

func Send(ctx define.IContext, body []byte) error {
	if act, ok := mapActor[ctx.GetActorName()]; ok {
		return act.Send(ctx, body)
	}
	return uerror.New(-1, "%s.%s未注册", ctx.GetActorName(), ctx.GetFuncName())
}

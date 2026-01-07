package actor

import (
	"framework/core"
	"framework/library/uerror"
)

var (
	mapActor = make(map[string]core.IActor)
)

func Register(act core.IActor) {
	mapActor[act.GetActorName()] = act
}

func SendMsg(ctx core.IContext, args ...any) error {
	if act, ok := mapActor[ctx.GetActorName()]; ok {
		return act.SendMsg(ctx, args...)
	}
	return uerror.New(-1, "%s.%s未注册", ctx.GetActorName(), ctx.GetFuncName())
}

func Send(ctx core.IContext, body []byte) error {
	if act, ok := mapActor[ctx.GetActorName()]; ok {
		return act.Send(ctx, body)
	}
	return uerror.New(-1, "%s.%s未注册", ctx.GetActorName(), ctx.GetFuncName())
}

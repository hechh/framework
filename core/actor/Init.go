package actor

import (
	"fmt"
	"strings"

	"github.com/hechh/framework/packet"
	"github.com/hechh/library/base/datetime"
	"github.com/hechh/library/mlog"
)

var (
	mapActor = make(map[string]IActor)
)

func Register(actor IActor) {
	mapActor[actor.GetName()] = actor
}

func Send(head *packet.Head, body []byte) error {
	if act, ok := mapActor[GetActorName(head.ActorFuncName)]; ok {
		return act.Send(head, body)
	}
	mlog.Errorf("Send: Actor(%s)未注册", head.ActorFuncName)
	return fmt.Errorf("%s未注册", head.ActorFuncName)
}

func SendMsg(head *packet.Head, args ...any) error {
	if act, ok := mapActor[GetActorName(head.ActorFuncName)]; ok {
		return act.SendMsg(head, args...)
	}
	mlog.Errorf("SendMsg: Actor(%s)未注册", head.ActorFuncName)
	return fmt.Errorf("%s未注册", head.ActorFuncName)
}

func GetActorName(name string) string {
	pos := strings.Index(name, ".")
	if pos == -1 {
		return name
	}
	return name[:pos]
}

func Tag(head *packet.Head) string {
	now := datetime.NowUnixMilli()
	return fmt.Sprintf("%dms", now-datetime.NanoToMilli(head.CreateTime))
}

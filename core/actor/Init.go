package actor

import (
	"fmt"
	"strings"
	"time"

	"github.com/hechh/framework/packet"
	"github.com/hechh/library/base/cache"
	"github.com/hechh/library/mlog"
	"github.com/hechh/library/msgqueue"
)

var (
	mapActor = make(map[string]IActor)
)

type IActor interface {
	GetName() string                                   // Actor名字
	GetId() uint64                                     // Actor ID
	Start() bool                                       // 启动任务队列
	Stop()                                             // 关闭actor任务队列协程
	Register(IActor, cache.ICache, ...msgqueue.Option) // 派生类自我注册
	RegisterTimer(string, time.Duration, int32) error  // 注册定时器
	SendMsg(*packet.Head, ...any) error                // 异步调用派生类成员函数
	Send(*packet.Head, []byte) error                   // 异步调用派生类成员函数
}

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

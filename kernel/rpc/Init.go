package rpc

import (
	"github.com/hechh/framework/library/enum"
	"github.com/hechh/framework/library/logic"
	"github.com/hechh/framework/packet"
)

var single = NewRpcMgr()

func Register0(nodeType, cmd any, name string, flags ...uint32) {
	single.Register(NewR0(enum.ToUint32(nodeType), enum.ToUint32(cmd), name, logic.Or(flags...)))
}

func Register1[A any](nodeType, cmd any, name string, flags ...uint32) {
	single.Register(NewR1[A](enum.ToUint32(nodeType), enum.ToUint32(cmd), name, logic.Or(flags...)))
}

func Register2[A any, T any](nodeType, cmd any, name string, flags ...uint32) {
	single.Register(NewR2[A, T](enum.ToUint32(nodeType), enum.ToUint32(cmd), name, logic.Or(flags...)))
}

func GetByActorFuncName(nodeType uint32, name string) IRpc {
	return single.GetByActorFuncName(nodeType, name)
}

func GetByActorFunc(nodeType uint32, id uint32) IRpc {
	return single.GetByActorFunc(nodeType, id)
}

func GetByCmd(cmd uint32) IRpc {
	return single.GetByCmd(cmd)
}

func GetRpc(head *packet.Head) IRpc {
	return single.GetRpc(head)
}

func GetAllCmds() map[uint32]IRpc {
	return single.GetAllCmds()
}

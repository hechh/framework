package handler

import (
	"github.com/hechh/framework/library/logic"
	"github.com/hechh/framework/library/utils"
	"github.com/hechh/framework/packet"
)

var obj = NewHandlerMgr()

// Register0 注册0参数函数
func Register0[A any](f M0Func[A], flags ...uint32) {
	obj.Register(NewM0(f, logic.Or(flags...)))
}

// Register1 注册1参数函数
func Register1[A any, R any](f M1Func[A, R], flags ...uint32) {
	obj.Register(NewM1(f, logic.Or(flags...)))
}

// Register2 注册2参数函数
func Register2[A any, R any, U any](f M2Func[A, R, U], flags ...uint32) {
	obj.Register(NewM2(f, logic.Or(flags...)))
}

// Get 获取处理函数
func Get(head *packet.Head) IHandler {
	return obj.Get(head)
}

// GetByName 获取处理函数
func GetByActorFuncName(name string) IHandler {
	return obj.GetByActorFuncName(name)
}

// GetById 获取处理函数
func GetByActorFunc(id uint32) IHandler {
	return obj.GetByActorFunc(id)
}

func NameToId(name string) uint32 {
	return utils.GetCrc32(name)
}

package handler

import (
	"fmt"

	"github.com/hechh/framework/define"
	"github.com/hechh/framework/packet"
)

type IHandler interface {
	GetActorFuncName() string                // 获取处理函数名称
	GetActorFunc() uint32                    // 获取处理函数唯一ID
	GetMask() uint32                         // 获取屏蔽模式
	Call(any, define.IContext, ...any) error // 调用处理函数
}

type HandlerMgr struct {
	names map[string]IHandler
	apis  map[uint32]IHandler
}

func NewHandlerMgr() *HandlerMgr {
	return &HandlerMgr{
		names: make(map[string]IHandler),
		apis:  make(map[uint32]IHandler),
	}
}

// Register 注册handler
func (d *HandlerMgr) Register(handler IHandler) {
	// 注册名称
	name := handler.GetActorFuncName()
	if _, ok := d.names[name]; ok {
		panic(fmt.Sprintf("handler(%s) is already registered", name))
	}
	d.names[name] = handler

	// 注册API
	if _, ok := d.apis[handler.GetActorFunc()]; ok {
		panic(fmt.Sprintf("handler(%s) actorFunc(%d) is already registered", name, handler.GetActorFunc()))
	}
	d.apis[handler.GetActorFunc()] = handler
}

// Get 获取handler
// 优先按名称查找，失败时回退到按数字 ID 查找（兼容跨服务路由时 ActorFuncName 残留的场景）
func (d *HandlerMgr) Get(head *packet.Head) IHandler {
	if len(head.ActorFuncName) > 0 {
		if h, ok := d.names[head.ActorFuncName]; ok {
			return h
		}
	}
	return d.apis[head.ActorFunc]
}

// GetByName 通过名称获取handler
func (d *HandlerMgr) GetByActorFuncName(name string) IHandler {
	return d.names[name]
}

// GetById 通过ID获取handler
func (d *HandlerMgr) GetByActorFunc(id uint32) IHandler {
	return d.apis[id]
}

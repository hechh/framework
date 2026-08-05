package router

import (
	"github.com/hechh/framework/packet"
)

var object *Router

func SetObject(oj *Router) {
	object = oj
}

// Get 获取路由
func Get(idType uint32, id uint64) *Entity {
	if object != nil {
		return object.Get(idType, id)
	}
	return nil
}

// GetOrNew 获取或创建路由
func GetOrNew(idType uint32, id uint64) *Entity {
	if object != nil {
		return object.GetOrNew(idType, id)
	}
	return nil
}

// Remove 移除路由
func Remove(idType uint32, id uint64) {
	if object != nil {
		object.Remove(idType, id)
	}
}

func RouteHandler(f func(*packet.Packet)) func(*packet.Packet) {
	return func(msg *packet.Packet) {
		for _, item := range msg.List {
			if item == nil {
				continue
			}
			route := GetOrNew(item.IdType, item.Id)
			for _, info := range item.List {
				route.Set(info.GetType(), info.GetId())
			}
		}
		f(msg)
	}
}

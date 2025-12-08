package context

import (
	"framework/core/handler"
	"framework/library/uerror"
	"framework/packet"
)

type Rpc struct {
	*packet.Head                // 原始头
	routerId     uint64         // 路由id
	origin       *packet.Router // 源路由
	current      *packet.Router // 当前路由
}

func NewRpc(head *packet.Head, routerId uint64, idType uint32, id uint64) *Rpc {
	newHead := *head
	ret := &Rpc{
		Head:     &newHead,
		routerId: routerId,
	}
	if id > 0 {
		ret.origin = &packet.Router{
			IdType: idType,
			Id:     id,
		}
		ret.IdType = idType
		ret.Id = id
	}
	return ret
}

func (d *Rpc) GetHead() *packet.Head {
	return d.Head
}

func (d *Rpc) GetRouterId() uint64 {
	return d.routerId
}

func (d *Rpc) GetRouters() (rets []*packet.Router) {
	if d.current != nil {
		rets = append(rets, d.current)
	}
	if d.origin != nil {
		rets = append(rets, d.origin)
	}
	return
}

// 设置路由
func (d *Rpc) SetRouter(idType uint32, id uint64) {
	if d.origin == nil || d.origin.IdType != idType || d.origin.Id != id {
		d.current = &packet.Router{
			IdType: idType,
			Id:     id,
		}
	}
}

// 设置回调
func (d *Rpc) Callback(actorFunc string, actorId uint64) error {
	hh := handler.Get(actorFunc)
	if hh == nil {
		return uerror.New(-1, "远程接口(%s)未注册", actorFunc)
	}
	d.Back = &packet.Callback{
		ActorFunc: hh.GetId(),
		ActorId:   actorId,
	}
	return nil
}

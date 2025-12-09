package context

import (
	"framework/library/uerror"
	"framework/packet"
	"framework/repository/bus"
	"framework/repository/cluster"
	"framework/repository/handler"
	"framework/repository/router"
)

type SendRpc struct {
	packet.Packet
	routerId uint64         // 路由id
	current  *packet.Router // 路由
}

func NewSendRpc(head *packet.Head) *SendRpc {
	return &SendRpc{Packet: packet.Packet{Head: head}}
}

// 设置当前路由
func (d *SendRpc) SetRouter(idType uint32, id uint64, routerId uint64, isOrigin bool) {
	d.current = &packet.Router{
		IdType: idType,
		Id:     id,
	}
	d.List = append(d.List, d.current)
	d.routerId = routerId
	if isOrigin && idType != d.Head.IdType && id != d.Head.Id {
		d.List = append(d.List, &packet.Router{
			IdType: d.Head.IdType,
			Id:     d.Head.Id,
		})
	}
}

func (d *SendRpc) SetCallback(actorFunc string, actorId uint64) error {
	hh := handler.Get(actorFunc)
	if hh == nil {
		return uerror.New(-1, "远程接口(%s)未注册", actorFunc)
	}
	d.Head.Back = &packet.Callback{
		NodeType:  d.Head.SrcNodeType,
		NodeId:    d.Head.SrcNodeId,
		ActorFunc: hh.GetId(),
		ActorId:   actorId,
	}
	return nil
}

func (d *SendRpc) Rpc(nodeType uint32, actorId uint64, api string, args ...any) error {
	if err := d.dispatcher(nodeType, actorId, api, args...); err != nil {
		return err
	}
	return bus.Send(&d.Packet)
}

func (d *SendRpc) dispatcher(nodeType uint32, actorId uint64, api string, args ...any) (err error) {
	// 获取远程rpc
	hh := handler.GetByRpc(nodeType, api)
	if hh == nil {
		return uerror.New(-1, "远程接口%s未注册", api)
	}
	d.Head.ActorFunc = hh.GetId()
	d.Head.ActorId = actorId

	// 序列化
	if d.Body, err = hh.Marshal(args...); err != nil {
		return err
	}

	// 获取集群
	cls := cluster.Get(nodeType)
	if cls == nil {
		return uerror.New(-1, "集群(%d)不支持", nodeType)
	}

	// 更新路由
	for i, item := range d.List {
		rr := router.GetOrNew(item.IdType, item.Id)
		if i == 0 {
			node := cls.Get(rr.Get(nodeType))
			if node == nil {
				node = cls.Random(d.routerId)
			}
			if node == nil {
				return uerror.New(-1, "集群(%d)不存在任何服务节点", nodeType)
			} else {
				d.Head.DstNodeType = node.Type
				d.Head.DstNodeId = node.Id
			}
		}
		rr.Set(d.Head.DstNodeType, d.Head.DstNodeId)
		rr.Update()
		item.List = rr.GetRouter()
	}
	return
}

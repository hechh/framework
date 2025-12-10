package context

import (
	"framework/define"
	"framework/library/structure"
	"framework/library/uerror"
	"framework/packet"
	"framework/repository/cluster"
	"framework/repository/handler"
	"framework/repository/router"
)

type Packet struct {
	err      error
	routerId uint64
	head     *packet.Head
	body     []byte
	list     []*packet.Router
}

func NewPakcet(head *packet.Head) *Packet {
	return &Packet{head: head}
}

func (d *Packet) Router(isOrigin bool, routerId uint64, args ...any) define.IPacket {
	d.routerId = routerId
	for _, arg := range args {
		switch vv := arg.(type) {
		case *packet.Router:
			d.list = append(d.list, vv)
		case *structure.Pair[uint32, uint64]:
			d.list = append(d.list, &packet.Router{
				IdType: vv.First,
				Id:     vv.Second,
			})
		case uint32:
			d.list = append(d.list, &packet.Router{
				IdType: vv,
			})
		case uint64:
			d.list[len(d.list)-1].Id = vv
		}
	}
	if isOrigin && d.list[0].IdType != d.head.IdType && d.list[0].Id != d.head.Id {
		d.list = append(d.list, &packet.Router{
			IdType: d.head.IdType,
			Id:     d.head.Id,
		})
	}
	return d
}

func (d *Packet) Callback(actorFunc string, actorId uint64) define.IPacket {
	hh := handler.Get(actorFunc)
	if hh == nil {
		d.err = uerror.New(-1, "远程接口(%s)未注册", actorFunc)
		return d
	}
	d.head.Back = &packet.Callback{
		NodeType:  d.head.SrcNodeType,
		NodeId:    d.head.SrcNodeId,
		ActorFunc: hh.GetId(),
		ActorId:   actorId,
	}
	return d
}

func (d *Packet) Rpc(nodeType uint32, actorId uint64, api string, args ...any) define.IPacket {
	if d.err != nil {
		return d
	}
	// 获取远程rpc
	hh := handler.GetByRpc(nodeType, api)
	if hh == nil {
		d.err = uerror.New(-1, "远程接口%s未注册", api)
		return d
	}

	// 序列化
	if d.body, d.err = hh.Marshal(args...); d.err != nil {
		return d
	}

	d.head.DstNodeType = nodeType
	d.head.ActorFunc = hh.GetId()
	d.head.ActorId = actorId

	// 获取集群
	cls := cluster.Get(nodeType)
	if cls == nil {
		d.err = uerror.New(-1, "集群(%d)不支持", nodeType)
		return d
	}

	// 更新路由
	for i, item := range d.list {
		rr := router.GetOrNew(item.IdType, item.Id)
		if i == 0 {
			node := cls.Get(rr.Get(nodeType))
			if node == nil {
				node = cls.Random(d.routerId)
			}
			if node == nil {
				d.err = uerror.New(-1, "集群(%d)不存在任何服务节点", nodeType)
				return d
			} else {
				d.head.DstNodeId = node.Id
			}
		}
		rr.Set(d.head.DstNodeType, d.head.DstNodeId)
		rr.Update()
		item.List = rr.GetRouter()
	}
	return d
}

func (d *Packet) GetPacket() (*packet.Packet, error) {
	if d.err != nil {
		return nil, d.err
	}
	return &packet.Packet{
		Head: d.head,
		Body: d.body,
		List: d.list,
	}, nil
}

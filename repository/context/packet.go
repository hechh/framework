package context

import (
	"framework/define"
	"framework/library/uerror"
	"framework/packet"
	"framework/repository/cluster"
	"framework/repository/handler"
	"framework/repository/router"
)

type Packet struct {
	head *packet.Head
	body []byte
	list []*packet.Router
	err  error
}

func NewPacket(ctx define.IContext) *Packet {
	ctx.AddDepth(1)
	return &Packet{
		head: ctx.GetHead(),
	}
}

func (d *Packet) Head(head *packet.Head) define.IPacket {
	if d.err == nil {
		d.head = head
	}
	return d
}

func (d *Packet) SendType(val uint32) define.IPacket {
	if d.err == nil {
		d.head.SendType = val
	}
	return d
}

func (d *Packet) ID(idType uint32, id uint64) define.IPacket {
	if d.err == nil {
		d.head.IdType = idType
		d.head.Id = id
	}
	return d
}

func (d *Packet) Router(idType uint32, id uint64) define.IPacket {
	if d.err == nil {
		d.list = append(d.list, &packet.Router{IdType: idType, Id: id})
	}
	return d
}

func (d *Packet) Callback(actorFunc string, actorId uint64) define.IPacket {
	if d.err != nil {
		return d
	}
	hh := handler.Get(actorFunc)
	if hh == nil {
		d.err = uerror.New(-1, "远程接口(%s)未注册", actorFunc)
		return d
	}
	d.head.Back = &packet.Callback{
		ActorFunc: hh.GetId(),
		ActorId:   actorId,
	}
	return d
}

func (d *Packet) Rpc(nodeType uint32, actorId uint64, actorFunc string, args ...any) define.IPacket {
	if d.err != nil {
		return d
	}
	hh := handler.GetByRpc(nodeType, actorFunc)
	if hh == nil {
		d.err = uerror.New(-1, "远程接口(%s)未注册", actorFunc)
		return d
	}
	d.head.DstNodeType = nodeType
	d.head.ActorFunc = hh.GetId()
	d.head.ActorId = actorId
	d.body, d.err = hh.Marshal(args...)
	return d
}

func (d *Packet) Cmd(cmd uint32, actorId uint64, args ...any) define.IPacket {
	if d.err != nil {
		return d
	}
	hh := handler.GetByCmd(cmd)
	if hh == nil {
		d.err = uerror.New(-1, "命令字(%d)未注册", cmd)
		return d
	}
	d.head.Cmd = cmd
	d.head.DstNodeType = hh.GetType()
	d.head.ActorFunc = hh.GetId()
	d.head.ActorId = actorId
	d.body, d.err = hh.Marshal(args...)
	return d
}

func (d *Packet) Dispatch(routerId uint64) (*packet.Packet, error) {
	if d.err != nil {
		return nil, d.err
	}

	cls := cluster.Get(d.head.DstNodeType)
	if cls == nil {
		return nil, uerror.Err(-1, "集群(%d)不支持", d.head.DstNodeType)
	}

	for i, item := range d.list {
		rr := router.GetOrNew(item.IdType, item.Id)
		if i == 0 {
			node := cls.Get(rr.Get(d.head.DstNodeType))
			if node == nil {
				node = cls.Random(routerId)
			}
			if node == nil {
				return nil, uerror.Err(-1, "集群(%d)不存在任何服务节点", d.head.DstNodeType)
			}
			d.head.DstNodeId = node.Id
		}
		rr.Set(d.head.DstNodeType, d.head.DstNodeId)
		rr.Update()
		item.List = rr.GetRouter()
	}
	return &packet.Packet{Head: d.head, Body: d.body, List: d.list}, nil
}

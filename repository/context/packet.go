package context

import (
	"framework/library/uerror"
	"framework/packet"
	"framework/repository/cluster"
	"framework/repository/handler"
	"framework/repository/router"
)

type Packet struct {
	packet.Packet
	err error
}

func (d *Packet) Get() (*packet.Packet, error) {
	return &d.Packet, d.err
}

func (d *Packet) Header(head *packet.Head) *Packet {
	if d.err != nil {
		return d
	}
	d.Head.SendType = head.SendType
	d.Head.IdType = head.IdType
	d.Head.Id = head.Id
	d.Head.Cmd = head.Cmd
	d.Head.Seq = head.Seq
	d.Head.Version = head.Version
	d.Head.SocketId = head.SocketId
	d.Head.Reply = head.Reply
	return d
}

func (d *Packet) SendType(val uint32) *Packet {
	if d.err == nil {
		d.Head.SendType = val
	}
	return d
}

func (d *Packet) ID(idType uint32, id uint64) *Packet {
	if d.err == nil {
		d.Head.IdType = idType
		d.Head.Id = id
	}
	return d
}

func (d *Packet) SocketId(socketId uint32) *Packet {
	if d.err == nil {
		d.Head.SocketId = socketId
	}
	return d
}

func (d *Packet) Router(idType uint32, id uint64) *Packet {
	if d.err == nil {
		d.List = append(d.List, &packet.Router{IdType: idType, Id: id})
	}
	return d
}

func (d *Packet) Rpc(nodeType uint32, actorId uint64, actorFunc string, args ...any) *Packet {
	if d.err != nil {
		return d
	}
	hh := handler.GetByRpc(nodeType, actorFunc)
	if hh == nil {
		d.err = uerror.New(-1, "远程接口(%s)未注册", actorFunc)
		return d
	}
	d.Head.DstNodeType = nodeType
	d.Head.ActorFunc = hh.GetId()
	d.Head.ActorId = actorId
	d.Body, d.err = hh.Marshal(args...)
	return d
}

func (d *Packet) Cmd(cmd uint32, actorId uint64, args ...any) *Packet {
	if d.err != nil {
		return d
	}
	hh := handler.GetByCmd(cmd)
	if hh == nil {
		d.err = uerror.New(-1, "命令字(%d)未注册", cmd)
		return d
	}
	d.Head.Cmd = cmd
	d.Head.DstNodeType = hh.GetType()
	d.Head.ActorFunc = hh.GetId()
	d.Head.ActorId = actorId
	d.Body, d.err = hh.Marshal(args...)
	return d
}

func (d *Packet) Callback(actorFunc string, actorId uint64) *Packet {
	if d.err != nil {
		return d
	}
	hh := handler.Get(actorFunc)
	if hh == nil {
		d.err = uerror.New(-1, "远程接口(%s)未注册", actorFunc)
		return d
	}
	d.Head.Back = &packet.Callback{
		ActorFunc: hh.GetId(),
		ActorId:   actorId,
	}
	return d
}

func (d *Packet) Dispatch(routerId uint64) *Packet {
	if d.err != nil {
		return d
	}
	cls := cluster.Get(d.Head.DstNodeType)
	if cls == nil {
		d.err = uerror.Err(-1, "集群(%d)不支持", d.Head.DstNodeType)
		return d
	}
	for i, item := range d.List {
		rr := router.GetOrNew(item.IdType, item.Id)
		if i == 0 {
			node := cls.Get(rr.Get(d.Head.DstNodeType))
			if node == nil {
				node = cls.Random(routerId)
			}
			if node == nil {
				d.err = uerror.Err(-1, "集群(%d)不存在任何服务节点", d.Head.DstNodeType)
				return d
			}
			d.Head.DstNodeId = node.Id
		}
		rr.Set(d.Head.DstNodeType, d.Head.DstNodeId)
		rr.Update()
		item.List = rr.GetRouter()
	}
	return d
}

package core

import (
	"framework/library/uerror"
	"framework/packet"

	"github.com/gogo/protobuf/proto"
)

type Packet struct {
	head *packet.Head
	body []byte
	list []*packet.Router
	err  error
}

func NewPacket(head *packet.Head) *Packet {
	return &Packet{
		head: head,
	}
}

func (d *Packet) Router(idType uint32, id uint64) IPacket {
	if d.err == nil {
		d.list = append(d.list, &packet.Router{IdType: idType, Id: id})
	}
	return d
}

func (d *Packet) Callback(actorId uint64, actorFunc string) IPacket {
	if d.err != nil {
		return d
	}
	hh := GetHandler(actorFunc)
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

func (d *Packet) Rsp(nodeType uint32, err error, rsp IRspHead) IPacket {
	if d.err != nil {
		return d
	}
	if err != nil {
		uerr := uerror.Turn(-1, err)
		rsp.SetRspHead(&packet.RspHead{Code: uerr.GetCode(), Msg: uerr.Error()})
	}
	if d.head.Back != nil {
		d.head.DstNodeType = d.head.Back.NodeType
		d.head.DstNodeId = d.head.Back.NodeId
		d.head.ActorFunc = d.head.Back.ActorFunc
		d.head.ActorId = d.head.Back.ActorId
		d.head.Back = nil
	} else {
		d.head.DstNodeType = nodeType
		d.head.ActorFunc = 0
		d.head.ActorId = 0
	}
	d.body, d.err = proto.Marshal(rsp)
	return d
}

func (d *Packet) Rpc(nodeType uint32, actorId uint64, actorFunc string, args ...any) IPacket {
	if d.err != nil {
		return d
	}
	hh := GetHandlerByRpc(nodeType, actorFunc)
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

func (d *Packet) Cmd(cmd uint32, actorId uint64, args ...any) IPacket {
	if d.err != nil {
		return d
	}
	hh := GetHandlerByCmd(cmd)
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

func (d *Packet) Dispatch(sendType packet.SendType) (*packet.Packet, error) {
	if d.err != nil {
		return nil, d.err
	}

	cls := GetCluster(d.head.DstNodeType)
	if cls == nil || cls.Size() <= 0 {
		return nil, uerror.Err(-1, "集群(%d)不支持", d.head.DstNodeType)
	}

	switch sendType {
	case packet.SendType_Point:
		if len(d.list) <= 0 {
			d.list = append(d.list, &packet.Router{IdType: d.head.IdType, Id: d.head.Id})
		}
		for i, item := range d.list {
			rr := GetOrNewRouter(item.IdType, item.Id)
			if i == 0 {
				node := cls.Get(rr.Get(d.head.DstNodeType))
				if node == nil {
					node = cls.Random(d.head.ActorId)
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
	case packet.SendType_Broadcast:
	}
	return &packet.Packet{Head: d.head, Body: d.body, List: d.list}, nil
}

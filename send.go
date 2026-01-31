package framework

import (
	"github.com/hechh/framework/packet"
	"github.com/hechh/library/uerror"
)

func Copy(head *packet.Head) *packet.Head {
	return &packet.Head{
		SendType:    head.SendType,
		SrcNodeType: GetSelfType(),
		SrcNodeId:   GetSelfId(),
		DstNodeType: head.DstNodeType,
		DstNodeId:   head.DstNodeId,
		IdType:      head.IdType,
		Id:          head.Id,
		Cmd:         head.Cmd,
		Seq:         head.Seq,
		ActorFunc:   head.ActorFunc,
		ActorId:     head.ActorId,
		Version:     head.Version,
		SocketId:    head.SocketId,
		Extra:       head.Extra,
		Reply:       head.Reply,
		Back:        head.Back,
	}
}

func Router(idType IEnum, id uint64) PacketFunc {
	return func(d *packet.Packet) error {
		d.List = append(d.List, &packet.Router{
			IdType: idType.Integer(),
			Id:     id,
		})
		return nil
	}
}

func Callback(name string, aid uint64) PacketFunc {
	return func(d *packet.Packet) error {
		rpc := GetRpc(GetSelfType(), name)
		if rpc == nil {
			return uerror.New(-1, "远程接口(%s)未注册", name)
		}
		d.Head.Back = &packet.Callback{
			ActorFunc: rpc.GetCrc32(),
			ActorId:   aid,
			NodeType:  GetSelfType(),
			NodeId:    GetSelfId(),
		}
		return nil
	}
}

func Rpc(nodeType IEnum, name string, aid uint64, args ...any) PacketFunc {
	return func(d *packet.Packet) error {
		rpc := GetRpc(nodeType.Integer(), name)
		if rpc == nil {
			return uerror.New(-1, "远程接口(%s)未注册", name)
		}
		d.Head.DstNodeType = rpc.GetNodeType()
		d.Head.ActorFunc = rpc.GetCrc32()
		d.Head.ActorId = aid
		if buf, err := rpc.Marshal(args...); err != nil {
			return err
		} else {
			d.Body = buf
		}
		return nil
	}
}

func Cmd(cmd IEnum, aid uint64, args ...any) PacketFunc {
	return func(d *packet.Packet) error {
		rpc := GetCmdRpc(cmd.Integer())
		if rpc == nil {
			return uerror.New(-1, "Cmd(%d)未注册", cmd)
		}
		d.Head.Cmd = cmd.Integer()
		d.Head.DstNodeType = rpc.GetNodeType()
		d.Head.ActorFunc = rpc.GetCrc32()
		d.Head.ActorId = aid
		if buf, err := rpc.Marshal(args...); err != nil {
			return err
		} else {
			d.Body = buf
		}
		return nil
	}
}

func Rsp(en ISerialize, cmd IEnum, args ...any) PacketFunc {
	return func(d *packet.Packet) error {
		if cmd != nil {
			d.Head.Cmd = cmd.Integer()
		}
		if len(d.Head.Reply) <= 0 {
			if d.Head.Back != nil {
				d.Head.DstNodeType = d.Head.Back.NodeType
				d.Head.DstNodeId = d.Head.Back.NodeId
				d.Head.ActorFunc = d.Head.Back.ActorFunc
				d.Head.ActorId = d.Head.Back.ActorId
				d.Head.Back = nil
			} else {
				d.Head.DstNodeType = NodeTypeGate
				d.Head.ActorFunc = 0
				d.Head.ActorId = 0
			}
		}
		if buf, err := en.Marshal(args...); err != nil {
			return err
		} else {
			d.Body = buf
		}
		return nil
	}
}

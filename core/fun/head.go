package fun

import (
	"github.com/hechh/framework/packet"
	"github.com/hechh/library/base/datetime"
	"github.com/hechh/library/base/enum"
	"github.com/hechh/library/base/utils"
)

func ACTOR(name string, aid uint64) func(*packet.Head) {
	return func(head *packet.Head) {
		head.ActorId = aid
		head.ActorFuncName = name
	}
}

func RPC(node enum.IEnum, name string, aid uint64) func(*packet.Head) {
	return func(head *packet.Head) {
		head.DstType = uint32(node.Number())
		head.ActorId = aid
		head.ActorFuncName = name
	}
}

func CMD(cmd enum.IEnum, aid uint64) func(*packet.Head) {
	return func(head *packet.Head) {
		head.Cmd = uint32(cmd.Number())
		head.ActorId = aid
	}
}

func UID(uid uint64) func(*packet.Head) {
	return func(head *packet.Head) {
		head.Uid = uid
	}
}

func TRACE(head *packet.Head) {
	head.CreateTime = datetime.NowUnixNano()
	head.TraceId = utils.GetTraceId(head.Uid, head.CreateTime)
}

func COPY(src *packet.Head) func(*packet.Head) {
	return func(dst *packet.Head) {
		dst.TraceId = src.TraceId
		dst.CreateTime = src.CreateTime
		dst.SendType = src.SendType
		dst.SrcType = src.SrcType
		dst.SrcId = src.SrcId
		dst.DstType = src.DstType
		dst.DstId = src.DstId
		dst.Uid = src.Uid
		dst.Cmd = src.Cmd
		dst.Seq = src.Seq
		dst.ActorFunc = src.ActorFunc
		dst.ActorId = src.ActorId
		dst.SocketId = src.SocketId
		dst.Version = src.Version
		dst.ClientIp = src.ClientIp
		dst.Back = src.Back
		dst.Reply = src.Reply
	}
}

func DERIVE(src *packet.Head) func(*packet.Head) {
	return func(dst *packet.Head) {
		dst.TraceId = src.TraceId
		dst.CreateTime = datetime.NowUnixNano()
		dst.SendType = src.SendType
		dst.Uid = src.Uid
		dst.SocketId = src.SocketId
		dst.Version = src.Version
		dst.ClientIp = src.ClientIp
	}
}

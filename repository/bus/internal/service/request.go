package service

import (
	"mypoker/common/pb"
	"mypoker/framework/internal/extern/handler"
	"mypoker/framework/internal/extern/router"
	"mypoker/framework/library/uerror"
)

// 监听同步
func ReadReply(f func(head *pb.Head, body []byte)) error {
	return net.ReadReply(func(pac *pb.Packet) {
		router.SetRouter(pac.Head.Current)
		if len(pac.Head.Router) > 0 {
			router.SetRouter(&pb.Router{IdType: pac.Head.IdType, Id: pac.Head.Id, List: pac.Head.Router})
		}
		pac.Head.Current = nil
		pac.Head.Router = pac.Head.Router[:0]
		pac.Head.ActorFunc = handler.Id2Name(pac.Head.Dst.ActorFunc)
		pac.Head.ActorId = pac.Head.Dst.ActorId
		f(pac.Head, pac.Body)
	})
}

// 同步请求
func Request(head *pb.Head, cb func([]byte) error, reqs ...any) error {
	// 判断是否注册了rpc
	rpc := handler.Get(head.Dst.NodeType, head.Dst.ActorFunc)
	if rpc == nil {
		return uerror.New(-1, "接口(%s)未注册或注册错误", handler.Id2Name(head.Dst.ActorFunc))
	}
	// 序列化
	buf, err := rpc.Marshal(reqs...)
	if err != nil {
		return err
	}
	// 发送
	return net.Request(&pb.Packet{Head: head, Body: buf}, cb)
}

// 同步应答
func Response(head *pb.Head, args ...any) error {
	rpc := handler.Get(head.Dst.NodeType, head.Dst.ActorFunc)
	if rpc == nil {
		return uerror.New(-1, "接口(%s)未注册或注册错误", handler.Id2Name(head.Dst.ActorFunc))
	}
	// 序列化
	buf, err := rpc.Marshal(args...)
	if err != nil {
		return err
	}
	return net.Response(head.Reply, buf)
}

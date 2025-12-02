package service

import (
	"mypoker/common/pb"
	"mypoker/framework/internal/extern/global"
	"mypoker/framework/internal/extern/handler"
	"mypoker/framework/internal/extern/mlog"
	"mypoker/framework/internal/extern/router"
	"mypoker/framework/library/uerror"
	"mypoker/framework/library/util"

	"github.com/golang/protobuf/proto"
)

// 监听单播
func ReadUnicast(f func(head *pb.Head, body []byte)) error {
	return net.Read(func(pac *pb.Packet) {
		head, dst := pac.Head, pac.Head.Dst
		router.SetRouter(head.Current)
		if len(head.Router) > 0 {
			router.SetRouter(&pb.Router{IdType: head.IdType, Id: head.Id, List: head.Router})
		}
		head.Current = nil
		head.Router = pac.Head.Router[:0]
		head.ActorFunc = handler.Id2Name(pac.Head.Dst.ActorFunc)
		head.ActorId = util.Or(dst.ActorId <= 0 || head.Id == dst.ActorId, head.Id, dst.ActorId)
		dst.ActorFunc = 0
		dst.ActorId = 0
		f(head, pac.Body)
	})
}

// 发送点对点
func Send(head *pb.Head, args ...any) error {
	// 判断是否注册了rpc
	rpc := handler.Get(head.Dst.NodeType, head.Dst.ActorFunc)
	if rpc == nil {
		return uerror.New(-1, "接口(%s)未注册或注册错误", handler.Id2Name(head.Dst.ActorFunc))
	}
	// 序列化
	buf, err := rpc.Marshal(args...)
	if err != nil {
		return err
	}
	return net.Write(&pb.Packet{Head: head, Body: buf})
}

func SendRaw(head *pb.Head, buf []byte) error {
	return net.Write(&pb.Packet{Head: head, Body: buf})
}

func SendToCallback(head *pb.Head, args ...any) error {
	head.Dst = head.Callback
	head.Callback = nil
	head.Src = &pb.Address{NodeType: global.GetSelfType(), NodeId: global.GetSelfId()}
	// 判断
	if head.Src.NodeType == head.Dst.NodeType {
		return uerror.New(-1, "禁止同一集群节点之间相互转发")
	}
	// 解析
	rpc := handler.Get(head.Dst.NodeType, head.Dst.ActorFunc)
	if rpc == nil {
		return uerror.New(-1, "接口(%s)未注册或注册错误", handler.Id2Name(head.Dst.ActorFunc))
	}
	// 序列化
	buf, err := rpc.Marshal(args...)
	if err != nil {
		return err
	}
	// 路由
	rr := router.Load(head.IdType, head.Id)
	if rr == nil {
		return uerror.New(-1, "向未登录玩家推送消息")
	}
	head.Current = nil
	head.Router = rr.GetData()
	return net.Write(&pb.Packet{Head: head, Body: buf})
}

// 回复客户端
func SendToClient(head *pb.Head, rsp proto.Message) error {
	head.Src = &pb.Address{NodeType: global.GetSelfType(), NodeId: global.GetSelfId()}
	head.Dst = &pb.Address{
		NodeType: pb.NodeType_Gate,
	}
	// 判断
	if head.Src.NodeType == head.Dst.NodeType {
		return uerror.New(-1, "禁止同一集群节点之间相互转发")
	}
	if head.Id <= 0 {
		return uerror.New(-1, "玩家uid不存在")
	}
	// 加载路由
	rr := router.Load(head.IdType, head.Id)
	if rr == nil {
		return uerror.New(-1, "向未登录玩家推送消息")
	}
	head.Router = rr.GetData()
	head.Current = nil
	// 序列化
	buf, err := proto.Marshal(rsp)
	if err != nil {
		return err
	}
	return net.Write(&pb.Packet{Head: head, Body: buf})
}

func NotifyToClient(head *pb.Head, rsp proto.Message, uids ...uint64) error {
	head.Src = &pb.Address{NodeType: global.GetSelfType(), NodeId: global.GetSelfId()}
	head.Callback = nil
	if head.Id > 0 {
		uids = append(uids, head.Id)
	}
	// 判断
	if head.Src.NodeType == head.Dst.NodeType {
		return uerror.New(-1, "禁止同一集群节点之间相互转发")
	}
	// 序列化
	buf, err := proto.Marshal(rsp)
	if err != nil {
		return err
	}
	// 遍历发送
	for _, uid := range uids {
		rr := router.Load(head.IdType, uid)
		if rr == nil {
			mlog.Error(0, "向未登录玩家推送消息: %d", uid)
			continue
		}
		// 发送回复
		head.Id = uid
		head.Router = rr.GetData()
		if err := net.Write(&pb.Packet{Head: head, Body: buf}); err != nil {
			mlog.Error(0, "通知客户端失败：head:%v, error:%v", head, err)
		}
	}
	return nil
}

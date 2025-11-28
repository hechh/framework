package service

import (
	"framework/define"
	"framework/internal/cluster"
	"framework/internal/global"
	"framework/internal/handler"
	"framework/internal/router"
	"framework/library/uerror"
	"framework/library/yaml"
	"framework/repository/sender/internal/entity"

	"github.com/golang/protobuf/proto"
)

var (
	net *entity.Bus
)

func Init(cfg *yaml.NatsConfig) error {
	conn, err := entity.NewNatsBus(cfg.Topic, cfg.Endpoints)
	if err != nil {
		return err
	}
	net = entity.NewBus(conn)
	return nil
}

func Close() {
	net.Close()
}

// 设置源地址
func Source(head *pb.Head, idType pb.IdType, id uint64, actorFunc string, actorId uint64) error {
	if len(actorFunc) > 0 {
		if head.Callback != nil {
			return uerror.New(-1, "回调请求不能递归发送回调请求")
		}
		head.Callback = &pb.Address{
			NodeType:  global.GetSelfType(),
			NodeId:    global.GetSelfId(),
			ActorFunc: handler.Name2Id(actorFunc),
			ActorId:   actorId,
		}
	}
	if head.IdType != idType || head.Id != id {
		head.Current = &pb.Router{IdType: idType, Id: id}
	}
	// 设置回调
	head.Src = &pb.Address{
		NodeType: global.GetSelfType(),
		NodeId:   global.GetSelfId(),
	}
	return nil
}

// 设置目的地址
func Destination(head *pb.Head, nodeType pb.NodeType, actorFunc string, actorId uint64, routerId uint64) error {
	if global.GetSelfType() == nodeType {
		return uerror.New(-1, "禁止同一集群节点之间相互转发")
	}
	cluster := cluster.Get(nodeType)
	if cluster == nil {
		return uerror.New(-1, "节点类型(%d)不支持", nodeType)
	}
	if cluster.GetCount() <= 0 {
		return uerror.New(-1, "集群(%d)中不存在任何服务节点", nodeType)
	}
	// 设置目的地址
	head.Dst = &pb.Address{
		NodeType:  nodeType,
		ActorFunc: handler.Name2Id(actorFunc),
		ActorId:   actorId,
	}
	// 从路由中加载节点
	var rr define.IRouter
	if head.Current != nil {
		rr = router.LoadOrNew(head.Current.IdType, head.Current.Id)
	} else {
		rr = router.LoadOrNew(head.IdType, head.Id)
	}
	var node *pb.Node
	if node = cluster.Get(rr.Get(nodeType)); node == nil {
		if node = cluster.Random(routerId); node == nil {
			return uerror.New(-1, "服务节点不存在或者异常下线")
		}
	}
	// 更新路由信息
	rr.Set(node.Type, node.Id)
	rr.UpdateTime()
	head.Dst.NodeId = node.Id
	if head.Current != nil {
		head.Current.List = rr.GetData()
	} else {
		head.Router = rr.GetData()
	}
	return nil
}

func SendResponse(head *pb.Head, rsps ...any) error {
	// 回复同步请求
	if len(head.Reply) > 0 {
		return Response(head, rsps...)
	}
	// 回复cmd命令字
	if head.Cmd > 0 {
		return SendToClient(head, rsps[0].(proto.Message))
	}
	// 自动回复
	if head.Callback != nil {
		return SendToCallback(head, rsps...)
	}
	return nil
}

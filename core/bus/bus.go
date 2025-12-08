package bus

import (
	"framework/core/bus/internal/service"
	"framework/core/cluster"
	"framework/core/define"
	"framework/core/handler"
	"framework/core/router"
	"framework/library/uerror"
	"framework/library/util"
	"framework/library/yaml"
	"framework/packet"
)

var (
	serviceObj = service.NewService()
)

func Init(cfg *yaml.NatsConfig, nn *packet.Node) error {
	return serviceObj.Init(cfg, nn)
}

func Close() {
	serviceObj.Close()
}

func SubscribeBroadcast(f func(head *packet.Head, body []byte)) error {
	return serviceObj.SubscribeBroadcast(f)
}

func SubscribeUnicast(f func(head *packet.Head, body []byte)) error {
	return serviceObj.SubscribeUnicast(f)
}

func SubscribeReply(f func(head *packet.Head, body []byte)) error {
	return serviceObj.SubscribeReply(f)
}

func Broadcast(idType uint32, id uint64, nodeType uint32, api string, actorId uint64, args ...any) error {
	cls := cluster.Get(nodeType)
	if cls == nil || cls.Size() <= 0 {
		return uerror.New(-1, "集群(%d)不存在", nodeType)
	}
	// 获取远程rpc
	hh := handler.GetByRpc(nodeType, api)
	if hh == nil {
		return uerror.New(-1, "远程接口%s未注册", api)
	}
	// 序列化
	buf, err := hh.Marshal(args...)
	if err != nil {
		return err
	}
	// 发送
	return serviceObj.Broadcast(&packet.Head{
		SendType:    1,
		DstNodeType: nodeType,
		IdType:      idType,
		Id:          id,
		ActorFunc:   hh.GetId(),
		ActorId:     util.Or(actorId > 0, actorId, id),
	}, buf)
}

func Send(rpc define.IRpc, nodeType uint32, api string, actorId uint64, args ...any) error {
	// 获取远程rpc
	hh := handler.GetByRpc(nodeType, api)
	if hh == nil {
		return uerror.New(-1, "远程接口%s未注册", api)
	}
	body, err := hh.Marshal(args...)
	if err != nil {
		return err
	}

	// 获取集群s
	cls := cluster.Get(nodeType)
	if cls == nil || cls.Size() <= 0 {
		return uerror.New(-1, "集群(%d)不存在", nodeType)
	}

	// 获取路由
	routers := rpc.GetRouters()
	rr := router.GetOrNew(routers[0].IdType, routers[0].Id)

	// 获取发送节点
	node := cls.Get(rr.Get(nodeType))
	if node == nil {
		node = cls.Random(rpc.GetRouterId())
	}

	// 更新路由
	rr.Set(node.Type, node.Id)
	rr.Update()
	routers[0].List = rr.GetRouter()
	if len(routers) > 1 {
		for _, item := range routers[1:] {
			itemRouter := router.GetOrNew(item.IdType, item.Id)
			item.List = itemRouter.GetRouter()
			itemRouter.Update()
		}
	}

	head := rpc.GetHead()
	head.DstNodeType = nodeType
	head.DstNodeId = node.Id
	head.ActorId = util.Or(actorId > 0, actorId, head.Id)
	head.ActorFunc = hh.GetId()
	return serviceObj.Send(head, body, routers...)
}

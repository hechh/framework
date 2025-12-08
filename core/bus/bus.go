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
	return serviceObj.Broadcast(&packet.Packet{
		Head: &packet.Head{
			SendType:    1,
			DstNodeType: nodeType,
			IdType:      idType,
			Id:          id,
			ActorFunc:   hh.GetId(),
			ActorId:     util.Or(actorId > 0, actorId, id),
		},
		Body: buf,
	})
}

func Send(rpc define.IRpc, nodeType uint32, api string, actorId uint64, args ...any) error {
	pack, err := dispatcher(rpc, nodeType, api, actorId, args...)
	if err != nil {
		return err
	}
	return serviceObj.Send(pack)
}

func Request(cb func([]byte) error, rpc define.IRpc, nodeType uint32, api string, actorId uint64, args ...any) error {
	pack, err := dispatcher(rpc, nodeType, api, actorId, args...)
	if err != nil {
		return err
	}
	return serviceObj.Request(pack, cb)
}

func dispatcher(rpc define.IRpc, nodeType uint32, api string, actorId uint64, args ...any) (*packet.Packet, error) {
	// 获取远程rpc
	hh := handler.GetByRpc(nodeType, api)
	if hh == nil {
		return nil, uerror.New(-1, "远程接口%s未注册", api)
	}

	// 获取集群
	cls := cluster.Get(nodeType)
	if cls == nil || cls.Size() <= 0 {
		return nil, uerror.New(-1, "集群(%d)不存在", nodeType)
	}

	// 序列化
	body, err := hh.Marshal(args...)
	if err != nil {
		return nil, err
	}
	pack := &packet.Packet{
		Head: rpc.GetHead(),
		List: rpc.GetRouters(),
		Body: body,
	}

	// 更新路由
	var node *packet.Node
	for i, item := range pack.List {
		rr := router.GetOrNew(item.IdType, item.Id)
		if i == 0 {
			if node = cls.Get(rr.Get(nodeType)); node == nil {
				node = cls.Random(rpc.GetRouterId())
			}
			rr.Set(node.Type, node.Id)
		}
		item.List = rr.GetRouter()
		rr.Update()
	}

	// 设置值
	pack.Head.DstNodeType = nodeType
	pack.Head.DstNodeId = node.Id
	pack.Head.ActorFunc = hh.GetId()
	pack.Head.ActorId = actorId
	return pack, nil
}

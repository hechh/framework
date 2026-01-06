package bus

import (
	"framework/core/bus/internal/service"
	"framework/core/cluster"
	"framework/core/define"
	"framework/core/handler"
	"framework/library/uerror"
	"framework/library/util"
	"framework/library/yaml"
	"framework/packet"
)

var (
	serviceObj = service.NewService()
)

func Init(cfg *yaml.NatsConfig) error {
	return serviceObj.Init(cfg)
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

func Broadcast(idType uint32, id uint64, nodeType uint32, actorId uint64, api string, args ...any) error {
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

func Send(pack define.IPacket, routerId uint64) error {
	msg, err := pack.Dispatch(routerId)
	if err != nil {
		return err
	}
	return serviceObj.Send(msg)
}

func Request(vv define.IPacket, routerId uint64, cb func([]byte) error) error {
	msg, err := vv.Dispatch(routerId)
	if err != nil {
		return err
	}
	return serviceObj.Request(msg, cb)
}

func Response(head *packet.Head, buf []byte) error {
	return serviceObj.Response(head, buf)
}

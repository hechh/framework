package bus

import (
	"github.com/hechh/framework"
	"github.com/hechh/framework/bus/internal/service"
	"github.com/hechh/framework/packet"
	"github.com/hechh/library/uerror"
	"github.com/hechh/library/yaml"
)

var (
	serviceObj = service.NewService()
)

func init() {
	framework.SetBus(SendResponse)
}

func Init(cfg *yaml.NatsConfig) error {
	return serviceObj.Init(cfg)
}

func Close() {
	serviceObj.Close()
}

func SubscribeBroadcast(f func(framework.IContext, []byte)) error {
	return serviceObj.SubscribeBroadcast(f)
}

func SubscribeUnicast(f func(framework.IContext, []byte)) error {
	return serviceObj.SubscribeUnicast(f)
}

func SubscribeReply(f func(framework.IContext, []byte)) error {
	return serviceObj.SubscribeReply(f)
}

func to(msg any, sendType packet.SendType) (pack *packet.Packet) {
	switch vv := msg.(type) {
	case *packet.Packet:
		pack = vv
	case *packet.Head:
		pack = &packet.Packet{Head: vv}
	}
	pack.Head.SendType = sendType
	return
}

func Broadcast(ctx framework.IContext, funcs ...framework.PacketFunc) (err error) {
	pack := to(ctx, packet.SendType_BROADCAST)
	for _, f := range funcs {
		if err = f(pack); err != nil {
			return
		}
	}
	if err = serviceObj.Broadcast(pack); err == nil {
		ctx.AddDepth(1)
	}
	return
}

func Send(ctx framework.IContext, funcs ...framework.PacketFunc) (err error) {
	pack := to(ctx, packet.SendType_POINT)
	funcs = append(funcs, dispatcher)
	for _, f := range funcs {
		if err = f(pack); err != nil {
			return
		}
	}
	if err = serviceObj.Send(pack); err == nil {
		ctx.AddDepth(1)
	}
	return
}

func Request(ctx framework.IContext, cb func([]byte) error, funcs ...framework.PacketFunc) (err error) {
	pack := to(ctx, packet.SendType_POINT)
	funcs = append(funcs, dispatcher)
	for _, f := range funcs {
		if err = f(pack); err != nil {
			return
		}
	}
	if err = serviceObj.Request(pack, cb); err == nil {
		ctx.AddDepth(1)
	}
	return
}

func Response(head *packet.Head, buf []byte) error {
	return serviceObj.Response(head, buf)
}

func SendResponse(ctx framework.IContext, funcs ...framework.PacketFunc) (err error) {
	pack := to(ctx, packet.SendType_POINT)
	for _, f := range funcs {
		if err = f(pack); err != nil {
			return
		}
	}
	if len(pack.Head.Reply) > 0 {
		return Response(pack.Head, pack.Body)
	}

	if err = dispatcher(pack); err != nil {
		return
	}
	return serviceObj.Send(pack)
}

func dispatcher(d *packet.Packet) error {
	// 集群
	cls := framework.GetCluster(d.Head.DstNodeType)
	if cls == nil || cls.Size() <= 0 {
		return uerror.Err(-1, "集群(%d)不支持", d.Head.DstNodeType)
	}

	if len(d.List) <= 0 {
		d.List = append(d.List, &packet.Router{IdType: d.Head.IdType, Id: d.Head.Id})
	}

	for i, item := range d.List {
		rr := framework.GetOrNewRouter(item.IdType, item.Id)
		if i == 0 {
			node := cls.Get(rr.Get(d.Head.DstNodeType))
			if node == nil {
				node = cls.Random(d.Head.ActorId)
			}
			if node == nil {
				return uerror.Err(-1, "集群(%d)不存在任何服务节点", d.Head.DstNodeType)
			}
			d.Head.DstNodeId = node.Id
		}
		rr.Set(d.Head.DstNodeType, d.Head.DstNodeId)
		rr.Update()
		item.List = rr.GetRouter()
	}
	return nil
}

package bus

import (
	"github.com/hechh/framework"
	"github.com/hechh/framework/bus/internal/service"
	"github.com/hechh/framework/packet"
	"github.com/hechh/library/mlog"
	"github.com/hechh/library/uerror"
	"github.com/hechh/library/yaml"
	"google.golang.org/protobuf/proto"
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

func SubscribeBroadcast(f func(*packet.Head, []byte)) error {
	return serviceObj.SubscribeBroadcast(f)
}

func SubscribeUnicast(f func(*packet.Head, []byte)) error {
	return serviceObj.SubscribeUnicast(f)
}

func SubscribeReply(f func(*packet.Head, []byte)) error {
	return serviceObj.SubscribeReply(f)
}

func to(msg any, sendType packet.SendType) (pack *packet.Packet) {
	switch vv := msg.(type) {
	case *packet.Packet:
		pack = vv
	case *packet.Head:
		pack = &packet.Packet{Head: vv}
	case framework.IContext:
		pack = &packet.Packet{Head: vv.GetHead()}
	case uint64:
		pack = &packet.Packet{Head: &packet.Head{Id: vv}}
	case nil:
		pack = &packet.Packet{Head: &packet.Head{}}
	}
	pack.Head.SendType = sendType
	return
}

func addDepth(msg any) {
	switch vv := msg.(type) {
	case framework.IContext:
		vv.AddDepth(1)
	}
}

func Broadcast(msg any, funcs ...framework.PacketFunc) (err error) {
	pack := to(msg, packet.SendType_BROADCAST)
	for _, f := range funcs {
		if err = f(pack); err != nil {
			return
		}
	}
	if err = serviceObj.Broadcast(pack); err == nil {
		addDepth(msg)
	}
	return
}

func Send(msg any, funcs ...framework.PacketFunc) (err error) {
	pack := to(msg, packet.SendType_POINT)
	funcs = append(funcs, dispatcher)
	for _, f := range funcs {
		if err = f(pack); err != nil {
			return
		}
	}
	if err = serviceObj.Send(pack); err == nil {
		addDepth(msg)
	}
	return
}

func Request(msg any, cb func([]byte) error, funcs ...framework.PacketFunc) (err error) {
	pack := to(msg, packet.SendType_POINT)
	funcs = append(funcs, dispatcher)
	for _, f := range funcs {
		if err = f(pack); err != nil {
			return
		}
	}
	return serviceObj.Request(pack, cb)
}

func Response(head *packet.Head, buf []byte) error {
	return serviceObj.Response(head, buf)
}

func Notify(cmd framework.IEnum, data proto.Message, uids ...uint64) error {
	buf, err := proto.Marshal(data)
	if err != nil {
		return err
	}
	pack := &packet.Packet{
		Head: &packet.Head{Cmd: cmd.Integer(), DstNodeType: framework.NodeTypeGate},
		Body: buf,
	}
	for _, uid := range uids {
		pack.Head.Id = uid
		err := dispatcher(pack)
		if err == nil {
			err = serviceObj.Send(pack)
		}
		if err != nil {
			mlog.Errorf("[notify] uid:%v, error:%v, cmd:%d, event:%v", uid, err, cmd.Integer(), data)
		}
	}
	return nil
}

func SendResponse(msg any, funcs ...framework.PacketFunc) (err error) {
	pack := to(msg, packet.SendType_POINT)
	defer mlog.Tracef("[SendResponse] 自动回复：head:%v, body:%d, error:%v", pack.Head, len(pack.Body), err)
	for _, f := range funcs {
		if err = f(pack); err != nil {
			return
		}
	}
	if len(pack.Head.Reply) > 0 {
		err = Response(pack.Head, pack.Body)
		return
	}
	if err = dispatcher(pack); err != nil {
		return
	}
	err = serviceObj.Send(pack)
	return
}

func dispatcher(d *packet.Packet) error {
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

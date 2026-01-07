package bus

import (
	"framework/core"
	"framework/core/bus/internal/service"
	"framework/library/yaml"
	"framework/packet"
)

var (
	serviceObj = service.NewService()
)

func init() {
	core.SetBroadcast(Broadcast)
	core.SetSend(Send)
	core.SetRequest(Request)
}

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

func Broadcast(pack core.IPacket) error {
	msg, err := pack.Dispatch(packet.SendType_Broadcast)
	if err != nil {
		return err
	}
	return serviceObj.Broadcast(msg)
}

func Send(pack core.IPacket) error {
	msg, err := pack.Dispatch(packet.SendType_Point)
	if err != nil {
		return err
	}
	return serviceObj.Send(msg)
}

func Request(vv core.IPacket, cb func([]byte) error) error {
	msg, err := vv.Dispatch(packet.SendType_Point)
	if err != nil {
		return err
	}
	return serviceObj.Request(msg, cb)
}

func Response(head *packet.Head, buf []byte) error {
	return serviceObj.Response(head, buf)
}

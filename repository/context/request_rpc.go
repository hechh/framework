package context

import (
	"framework/packet"
	"framework/repository/bus"
)

type RequestRpc struct {
	SendRpc
	rspFunc func([]byte) error
}

func NewRequestRpc(head *packet.Head, rsp func([]byte) error) *RequestRpc {
	return &RequestRpc{
		SendRpc: SendRpc{
			Packet: packet.Packet{Head: head},
		},
		rspFunc: rsp,
	}
}

func (d *RequestRpc) Rpc(nodeType uint32, actorId uint64, api string, args ...any) error {
	if err := d.dispatcher(nodeType, actorId, api, args...); err != nil {
		return err
	}
	return bus.Request(&d.Packet, d.rspFunc)
}

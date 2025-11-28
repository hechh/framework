package sender

import (
	"framework/packet"
)

var (
	rspFunc func(*packet.Head, ...any) error // actor发送应答接口
)

func SetRspFunc(f func(*packet.Head, ...any) error) {
	rspFunc = f
}

func Response(head *packet.Head, rsps ...any) error {
	return rspFunc(head, rsps...)
}

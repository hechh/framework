package sender

import (
	"framework/internal/sender"
	"framework/library/yaml"
	"framework/repository/sender/internal/service"

	"github.com/golang/protobuf/proto"
)

func init() {
	sender.SetRspFunc(service.SendResponse)
}

func Init(cfg *yaml.NatsConfig) error {
	return service.Init(cfg)
}

func Close() {
	service.Close()
}

func ReadBroadcast(f func(*pb.Head, []byte)) error {
	return service.ReadBroadcast(f)
}

func ReadUnicast(f func(*pb.Head, []byte)) error {
	return service.ReadUnicast(f)
}

func ReadReply(f func(*pb.Head, []byte)) error {
	return service.ReadReply(f)
}

// 设置源地址的回调和路由
func Source(head *pb.Head, idType pb.IdType, id uint64, actorFunc string, actorId uint64) error {
	return service.Source(head, idType, id, actorFunc, actorId)
}

// 设置目的地址的接口
func Destination(head *pb.Head, nodeType pb.NodeType, actorFunc string, actorId uint64, routerId uint64) error {
	return service.Destination(head, nodeType, actorFunc, actorId, routerId)
}

// 向集群发送广播
func Broadcast(head *pb.Head, nodeType pb.NodeType, actorFunc string, args ...any) error {
	return service.Broadcast(head, nodeType, actorFunc, args...)
}

// 向节点发送消息
func Send(head *pb.Head, args ...any) error {
	return service.Send(head, args...)
}

func SendRaw(head *pb.Head, body []byte) error {
	return service.SendRaw(head, body)
}

// 向客户端回复
func SendToClient(head *pb.Head, rsp proto.Message) error {
	return service.SendToClient(head, rsp)
}

// 向多个客户端广播
func NotifyToClient(head *pb.Head, rsp proto.Message, uids ...uint64) error {
	return service.NotifyToClient(head, rsp, uids...)
}

// 发送同步请求
func Request(head *pb.Head, routerId uint64, cb func([]byte) error, reqs ...any) error {
	return service.Request(head, cb, reqs...)
}

// 回复同步请求
func Response(head *pb.Head, args ...any) error {
	return service.Response(head, args...)
}

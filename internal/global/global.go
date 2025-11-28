package global

import (
	"framework/library/uerror"
	"framework/packet"
	"reflect"
	"strings"
	"sync/atomic"
)

var (
	self             *packet.Node // 自身节点
	actorIdGenerator = uint64(0)  // actorId生成器
	rspFunc          func(*packet.Head, ...any) error
)

func SetRspFunc(f func(*packet.Head, ...any) error) {
	rspFunc = f
}

func SendResponse(head *packet.Head, args ...any) error {
	if rspFunc != nil {
		return rspFunc(head, args...)
	}
	return uerror.New(-1, "未注册自动回复接口")
}

// 生成 actorId
func GenerateActorId() uint64 {
	return atomic.AddUint64(&actorIdGenerator, 1)
}

func ParseActorName(rr reflect.Type) string {
	name := rr.String()
	if index := strings.Index(name, "."); index > -1 {
		name = name[index+1:]
	}
	return name
}

// 获取当前节点
func SetSelf(nn *packet.Node) {
	self = nn
}

func GetSelf() *packet.Node {
	return self
}

func GetSelfNodeType() int32 {
	return self.GetType()
}

func GetSelfNodeId() int32 {
	return self.GetId()
}

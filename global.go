package framework

import (
	"hash/crc32"
	"reflect"
	"runtime"
	"strings"
	"sync/atomic"

	"github.com/hechh/framework/packet"
	"github.com/hechh/library/uerror"
)

var (
	self               *packet.Node
	actorIdGenerator   = uint64(0) // actorId生成器
	socketIdGenerator  = uint32(0)
	NodeTypeGate       = uint32(1) // 网关节点类型
	HeartTimeExpire    = int64(6)  // 心跳过期时间
	RouterSyncInterval = int64(5)  // 路由同步间隔
)

func Init(gate uint32, nn *packet.Node) {
	self = nn
	NodeTypeGate = gate
}

func GetSelf() *packet.Node {
	return self
}

func GetSelfType() uint32 {
	return self.Type
}

func GetSelfName() string {
	return self.Name
}

func GetSelfId() uint32 {
	return self.Id
}

func ToRspHead(err error) (int32, string) {
	switch vv := err.(type) {
	case *uerror.UError:
		return vv.GetCode(), vv.GetMsg()
	case nil:
		return 0, ""
	default:
		return -1, vv.Error()
	}
}

func GetCrc32(name string) uint32 {
	return crc32.ChecksumIEEE([]byte(name))
}

func GenSocketId() uint32 {
	return atomic.AddUint32(&socketIdGenerator, 1)
}

// 生成 actorId
func GenActorId() uint64 {
	return atomic.AddUint64(&actorIdGenerator, 1)
}

func ParseActorName(rr reflect.Type) string {
	name := rr.String()
	if index := strings.Index(name, "."); index > -1 {
		name = name[index+1:]
	}
	return name
}

func ParseActorFunc(fun reflect.Value) string {
	runName := runtime.FuncForPC(fun.Pointer()).Name()
	strs := strings.Split(runName, "(*")
	return strings.ReplaceAll(strs[len(strs)-1], ")", "")
}

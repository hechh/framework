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
	self              *packet.Node
	NodeTypeGate      uint32
	actorIdGenerator  = uint64(0) // actorId生成器
	socketIdGenerator = uint32(0)
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

func ToRspHead(err error) *packet.RspHead {
	switch vv := err.(type) {
	case *uerror.UError:
		return &packet.RspHead{Code: vv.GetCode(), Msg: vv.GetMsg()}
	case nil:
		return nil
	default:
		return &packet.RspHead{Code: -1, Msg: vv.Error()}
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

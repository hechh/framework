package global

import (
	"framework/core/define"
	"framework/packet"
	"hash/crc32"
	"reflect"
	"runtime"
	"strings"
	"sync/atomic"
)

var (
	self              *packet.Node
	actorIdGenerator  = uint64(0) // actorId生成器
	socketIdGenerator = uint32(0)
	// handler
	getByName func(...string) define.IHandler
	getByCmd  func(uint32) define.IHandler
	getByRpc  func(uint32, any) define.IHandler
	// bus
	broadcast func(define.IPacket) error
	send      func(define.IPacket) error
	request   func(define.IPacket, func([]byte) error) error
	// router
	getRouter      func(uint32, uint64) define.IRouter
	getOrNewRouter func(uint32, uint64) define.IRouter
)

func SetSelf(nn *packet.Node) {
	self = nn
}

func GetSelf() *packet.Node {
	return self
}

func GetSelfType() uint32 {
	return self.Type
}

func GetSelfId() uint32 {
	return self.Id
}

// 生成 actorId
func GenerateActorId() uint64 {
	return atomic.AddUint64(&actorIdGenerator, 1)
}

func GenerateSocketId() uint32 {
	return atomic.AddUint32(&socketIdGenerator, 1)
}

func CopyTo(dst *packet.Head, src *packet.Head) {
	dst.SendType = src.SendType
	dst.SrcNodeType = src.SrcNodeType
	dst.SrcNodeId = src.SrcNodeId
	dst.DstNodeType = src.DstNodeType
	dst.DstNodeId = src.DstNodeId
	dst.IdType = src.IdType
	dst.Id = src.Id
	dst.Cmd = src.Cmd
	dst.Seq = src.Seq
	dst.ActorFunc = src.ActorFunc
	dst.ActorId = src.ActorId
	dst.Version = src.Version
	dst.SocketId = src.SocketId
	dst.Reply = src.Reply
}

func ParseActorFunc(fun reflect.Value) string {
	runName := runtime.FuncForPC(fun.Pointer()).Name()
	strs := strings.Split(runName, "(*")
	return strings.ReplaceAll(strs[len(strs)-1], ")", "")
}

func ParseActorName(rr reflect.Type) string {
	name := rr.String()
	if index := strings.Index(name, "."); index > -1 {
		name = name[index+1:]
	}
	return name
}

func GetCrc32(name string) uint32 {
	return crc32.ChecksumIEEE([]byte(name))
}

// 解决循环引用
func SetBroadcast(f func(define.IPacket) error) {
	broadcast = f
}

func SetSend(f func(define.IPacket) error) {
	send = f
}

func SetRequest(f func(define.IPacket, func([]byte) error) error) {
	request = f
}

func Broadcast(pack define.IPacket) error {
	return broadcast(pack)
}

func Send(pack define.IPacket) error {
	return send(pack)
}

func Request(pack define.IPacket, cb func([]byte) error) error {
	return request(pack, cb)
}

// ---------handler----------
func SetGetHandler(fn func(...string) define.IHandler) {
	getByName = fn
}

func SetGetHandlerByCmd(fn func(uint32) define.IHandler) {
	getByCmd = fn
}

func SetGetHandlerByRpc(fn func(uint32, any) define.IHandler) {
	getByRpc = fn
}

func GetHandler(names ...string) define.IHandler {
	return getByName(names...)
}

func GetHandlerByCmd(cmd uint32) define.IHandler {
	return getByCmd(cmd)
}

func GetHandlerByRpc(nodeType uint32, id any) define.IHandler {
	return getByRpc(nodeType, id)
}

// ----------------router-------------
func SetGetRouter(f func(uint32, uint64) define.IRouter) {
	getRouter = f
}

func SetGetOrNewRouter(f func(uint32, uint64) define.IRouter) {
	getOrNewRouter = f
}

func GetRouter(idType uint32, id uint64) define.IRouter {
	return getRouter(idType, id)
}

func GetOrNewRouter(idType uint32, id uint64) define.IRouter {
	return getOrNewRouter(idType, id)
}

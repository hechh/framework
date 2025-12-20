package common

import (
	"framework/packet"
	"hash/crc32"
	"reflect"
	"runtime"
	"strings"
)

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

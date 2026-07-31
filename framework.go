package framework

import (
	"os"

	"github.com/hechh/framework/core/msgbus"
	"github.com/hechh/framework/core/network"
	"github.com/hechh/framework/global"
	"github.com/hechh/framework/packet"
	"github.com/hechh/framework/service"
	"github.com/hechh/library/base/enum"
)

var (
	object = service.New()
)

func SetNodeConvertor(n map[string]int32, i map[int32]string) {
	global.NodeConvertor = enum.WrapConvertor(n, i)
}

func SetCmdConvertor(n map[string]int32, i map[int32]string) {
	global.CmdConvertor = enum.WrapConvertor(n, i)
}

func SetNetHandler(f func(*packet.Packet) error) {
	network.SetPacketFunc(f)
}

func SetNetDecoder(f func([]byte) (*packet.Packet, error)) {
	network.SetDecodeFunc(f)
}

func SetNetEncoder(f func(*packet.Packet) ([]byte, error)) {
	network.SetEncodeFunc(f)
}

func SetMsgHandler(f func(*packet.Packet)) {
	msgbus.SetPacketFunc(f)
}

func Register(c any) {
	object.Register(c)
}

func Init() error {
	return object.Init()
}

func Close() {
	object.Close()
}

func Run(sigs ...os.Signal) {
	object.Run(sigs...)
}

func GetApp() *service.Service {
	return object
}

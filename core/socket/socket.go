package socket

import (
	"framework/core/define"
	"framework/core/socket/internal/service"
	"framework/library/yaml"
	"framework/packet"
)

var (
	serviceObj *service.Websocket
)

func Init(cfg *yaml.NodeConfig, f define.IFrame, h func(*packet.Packet) error) error {
	serviceObj = service.NewWebsocket(f, h)
	return serviceObj.Init(cfg.Ip, cfg.Port)
}

func Close() {
	serviceObj.Close()
}

func Send(pac *packet.Packet) error {
	return serviceObj.Send(pac)
}

func Remove(socketId uint32) {
	serviceObj.Remove(socketId)
}

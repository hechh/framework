package socket

import (
	"fmt"
	"net"

	"github.com/hechh/framework"
	"github.com/hechh/framework/packet"
	"github.com/hechh/framework/socket/internal/entity"
	"github.com/hechh/framework/socket/internal/service"
	"github.com/hechh/library/yaml"

	"github.com/gorilla/websocket"
)

var (
	serviceObj = service.NewWebsocket()
)

func Init(cfg *yaml.NodeConfig, f framework.IFrame, h func(*packet.Packet) error) error {
	return serviceObj.Init(f, h, fmt.Sprintf("%s:%d", cfg.Ip, cfg.Port))
}

func Close() {
	serviceObj.Close()
}

func Send(pac *packet.Packet) error {
	return serviceObj.Send(pac)
}

func Stop(id uint32) {
	serviceObj.Stop(id)
}

func ConnWrapper(cc *websocket.Conn) net.Conn {
	return entity.NewWebsocketWrapper(cc)
}

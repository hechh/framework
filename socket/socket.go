package socket

import (
	"fmt"
	"net"

	"github.com/hechh/framework"
	"github.com/hechh/framework/packet"
	"github.com/hechh/framework/socket/internal/entity"
	"github.com/hechh/framework/socket/internal/service"
	"github.com/hechh/library/mlog"
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
	err := serviceObj.Send(pac)
	mlog.Tracef("[socket] 发送客户端消息 head:%v, body:%d, error:%v", pac.Head, len(pac.Body), err)
	return err
}

func Stop(id uint32) {
	serviceObj.Stop(id)
	mlog.Tracef("[socket] 关闭链接%d", id)
}

func ConnWrapper(cc *websocket.Conn) net.Conn {
	return entity.NewWebsocketWrapper(cc)
}

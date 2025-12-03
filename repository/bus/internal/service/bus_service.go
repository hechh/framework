package service

import (
	"framework/internal/handler"
	"framework/library/uerror"
	"framework/library/yaml"
	"framework/packet"
	"framework/repository/bus/internal/entity"
	"mypoker/common/pb"
)

type BusService struct {
	*entity.Bus
}

func (d *BusService) Init(cfg *yaml.NatsConfig) error {
	conn, err := entity.NewNatsBus(cfg.Topic, cfg.Endpoints)
	if err != nil {
		return err
	}
	d.Bus = entity.NewBus(conn)
	return nil
}

func (d *BusService) Close() {
	d.Bus.Close()
}

// 监听广播
func (d *BusService) ReadBroadcast(f func(head *packet.Head, body []byte)) error {
	return net.ReadBroadcast(func(pac packet.Packet) {
		f(pac.Head, pac.Body)
	})
}

// 发送广播
func (d *BusService) Broadcast(head packet.Head, args ...any) error {
	// 序列化
	hh := handler.Get(head.DstNodeType, head.ActorFunc)
	if hh == nil {
		return uerror.New(-1, "接口(%s)未注册或注册错误", hh.GetName())
	}
	buf, err := hh.Marshal(args...)
	if err != nil {
		return err
	}
	// 设置
	return net.Broadcast(pb.Packet{
		Head: head,
		Body: buf,
	})
}

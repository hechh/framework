package service

import (
	"framework/define"
	"framework/internal/cluster"
	"framework/internal/global"
	"framework/library/uerror"
	"framework/library/yaml"
	"framework/packet"
	"framework/repository/sender/internal/entity"
	"mypoker/common/pb"
	"mypoker/framework/internal/extern/handler"
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
func (d *BusService) Broadcast(ctx define.IContext, args ...any) error {
	// 判断是否合法
	cluster := cluster.Get(nodeType)
	if cluster == nil {
		return uerror.New(-1, "节点类型(%d)不支持", nodeType)
	}
	if cluster.GetCount() <= 0 {
		return uerror.New(-1, " 节点集群(%d)没有任何服务节点", nodeType)
	}
	if global.GetSelfType() == nodeType {
		return uerror.New(-1, "禁止同一集群节点之间相互转发")
	}
	// 序列化
	hh := handler.Get(nodeType, actorFunc)
	if hh == nil {
		return uerror.New(-1, "接口(%s)未注册或注册错误", actorFunc)
	}
	buf, err := hh.Marshal(args...)
	if err != nil {
		return err
	}

	head.Src = &pb.Address{
		NodeType: global.GetSelfType(),
		NodeId:   global.GetSelfId(),
	}

	head.Dst = &pb.Address{
		NodeType:  nodeType,
		ActorFunc: hh.GetApi(),
	}
	// 设置
	return net.Broadcast(&pb.Packet{Head: head, Body: buf})
}

package context

import (
	"framework/core/handler"
	"framework/library/uerror"
	"framework/packet"
)

type Rpc struct {
	*packet.Head                // 原始头
	routerId     uint64         // 路由id
	origin       *packet.Router // 源路由
	current      *packet.Router // 当前路由
}

func NewRpc(head *packet.Head, routerId uint64, idType uint32, id uint64) *Rpc {
	newHead := *head
	ret := &Rpc{
		Head:     &newHead,
		routerId: routerId,
	}
	if id > 0 {
		ret.origin = &packet.Router{
			IdType: idType,
			Id:     id,
		}
		ret.IdType = idType
		ret.Id = id
	}
	return ret
}

func (d *Rpc) GetHead() *packet.Head {
	return d.Head
}

// 设置路由
func (d *Rpc) SetRouter(idType uint32, id uint64) {
	if d.origin.IdType != idType || d.origin.Id != id {
		d.current = &packet.Router{
			IdType: idType,
			Id:     id,
		}
	}
}

// 设置回调
func (d *Rpc) Callback(nodeType uint32, actorFunc string, actorId uint64) error {
	hh := handler.Get(actorFunc)
	if hh == nil {
		return uerror.New(-1, "远程接口(%s)未注册", actorFunc)
	}
	d.Back = &packet.Callback{
		NodeType:  nodeType,
		ActorFunc: hh.GetId(),
		ActorId:   actorId,
	}
	return nil
}

func (d *Rpc) GetRouterId() uint64 {
	return d.routerId
}

func (d *Rpc) GetRouters() (rets []*packet.Router) {
	if d.current != nil {
		rets = append(rets, d.current)
	}
	if d.origin != nil {
		rets = append(rets, d.origin)
	}
	return
}

// 设置远程调用
/*
func (d *Rpc) Rpc(nodeType uint32, actorFunc string, actorId uint64, args ...any) (*packet.Head, []byte, []*packet.Router, error) {
	// 获取注册接口
	hh := handler.GetByRpc(nodeType, actorFunc)
	if hh == nil {
		return nil, nil, nil, uerror.New(-1, "远程接口(%s)未注册", actorFunc)
	}

	// 序列化参数
	body, err := hh.Marshal(args...)
	if err != nil {
		return nil, nil, nil, err
	}

	// 获取集群管理接口
	cls := cluster.Get(nodeType)
	if cls == nil || cls.Size() <= 0 {
		return nil, nil, nil, uerror.New(-1, "集群(%d)不存在", nodeType)
	}

	// 获取服务节点
	var rr define.IRouter
	if d.current != nil {
		rr = router.LoadOrNew(d.current.IdType, d.current.Id)
	} else {
		rr = router.LoadOrNew(d.origin.IdType, d.origin.Id)
	}
	node := cls.Get(rr.Get(nodeType))
	if node == nil {
		node = cls.Random(d.routerId)
	}
	if node == nil {
		err = uerror.New(-1, "服务节点不存在或者异常下线")
		return
	}
	// 更新路由信息
	rr.Set(node.Type, node.Id)
	rr.Update()
	// 设置参数
	if d.current != nil {
		d.origin.List = router.LoadOrNew(d.origin.IdType, d.origin.Id).GetRouter()
		d.current.List = rr.GetRouter()
		ret.List = append(ret.List, d.origin, d.current)
	} else {
		d.origin.List = rr.GetRouter()
		ret.List = append(ret.List, d.current)
	}
	ret.Head = &packet.Head{
		SendType:    sendType,
		SrcNodeType: global.GetSelfType(),
		SrcNodeId:   global.GetSelfId(),
		DstNodeType: node.Type,
		DstNodeId:   node.Id,
		IdType:      d.origin.IdType,
		Id:          d.origin.Id,
		Cmd:         d.Cmd,
		Seq:         d.Seq,
		ActorFunc:   hh.GetId(),
		ActorId:     actorId,
		Version:     d.Version,
		Extra:       d.Extra,
		Reply:       d.Reply,
		Back:        d.cb,
	}
	return
}
*/

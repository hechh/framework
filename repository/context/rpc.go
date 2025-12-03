package context

import (
	"framework/define"
	"framework/internal/cluster"
	"framework/internal/global"
	"framework/internal/handler"
	"framework/internal/router"
	"framework/library/uerror"
	"framework/packet"
)

type Rpc struct {
	*packet.Head                  // 原始头
	routerId     uint64           // 路由id
	origin       *packet.Router   // 源路由
	current      *packet.Router   // 当前路由
	cb           *packet.Callback // 回调
}

func NewRpc(head *packet.Head) *Rpc {
	ret := &Rpc{Head: head}
	ret.Router(head.IdType, head.Id, head.Id)
	return ret
}

// 设置路由
func (d *Rpc) Router(idType int32, id uint64, routerId uint64) {
	d.routerId = routerId
	if d.origin.IdType != idType || d.origin.Id != id {
		d.current = &packet.Router{
			IdType: idType,
			Id:     id,
		}
	}
}

// 设置回调
func (d *Rpc) Callback(actorFunc string, actorId uint64) error {
	if len(actorFunc) <= 0 {
		return nil
	}
	hh := handler.Get(actorFunc)
	if hh == nil {
		return uerror.New(-1, "回调接口(%s)未注册", actorFunc)
	}
	d.cb = &packet.Callback{
		NodeType:  global.GetSelfType(),
		NodeId:    global.GetSelfId(),
		ActorFunc: hh.GetId(),
		ActorId:   actorId,
	}
	return nil
}

// 设置远程调用
func (d *Rpc) Rpc(sendType int32, nodeType int32, actorFunc string, actorId uint64, args ...any) (ret packet.Packet, err error) {
	if global.GetSelfType() == nodeType {
		err = uerror.New(-1, "禁止同一集群节点之间相互转发")
		return
	}
	// 获取注册接口
	hh := handler.Get(nodeType, actorFunc)
	if hh == nil {
		err = uerror.New(-1, "Rpc接口(%s)未注册", actorFunc)
		return
	}
	// 序列化参数
	if ret.Body, err = hh.Marshal(args...); err != nil {
		return
	}
	// 获取集群管理接口
	cls := cluster.Get(nodeType)
	if cls == nil {
		err = uerror.New(-1, "节点类型(%d)不支持", nodeType)
		return
	}
	if cls.Size() <= 0 {
		err = uerror.New(-1, "集群(%d)中不存在任何服务节点", nodeType)
		return
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

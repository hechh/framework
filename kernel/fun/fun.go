package fun

import (
	"fmt"

	"github.com/hechh/framework/kernel/cluster"
	"github.com/hechh/framework/kernel/global"
	"github.com/hechh/framework/kernel/router"
	"github.com/hechh/framework/kernel/rpc"
	"github.com/hechh/framework/library/tplutil"
	"github.com/hechh/framework/packet"
)

func SetRouter(idType uint32, id uint64) func(*packet.Packet) error {
	return func(pack *packet.Packet) error {
		pack.List = append(pack.List, &packet.Router{
			IdType: idType,
			Id:     id,
		})
		return nil
	}
}

func SetBack(name string, actorId uint64) func(*packet.Packet) error {
	return func(pack *packet.Packet) error {
		api := rpc.GetByActorFuncName(global.GetSelfNodeType(), name)
		if api == nil {
			return fmt.Errorf("Rpc(%s)未注册", name)
		}
		pack.Head.Back = &packet.Callback{
			DstType:   global.GetSelfNodeType(),
			DstId:     global.GetSelfNodeId(),
			ActorFunc: api.GetActorFunc(),
			ActorId:   actorId,
		}
		return nil
	}
}

// 请求cmd接口
func SetCmd(pack *packet.Packet) error {
	api := rpc.GetByCmd(pack.Head.Cmd)
	if api == nil {
		return fmt.Errorf("CMD(%d)未注册", pack.Head.Cmd)
	}
	pack.Head.ActorFunc = api.GetActorFunc()
	pack.Head.DstType = api.GetNodeType()
	return nil
}

// 设置远程调用
func SetRpc(pack *packet.Packet) error {
	api := rpc.GetByActorFuncName(pack.Head.DstType, pack.Head.ActorFuncName)
	if api == nil {
		return fmt.Errorf("Rpc(%s)未注册", pack.Head.ActorFuncName)
	}
	pack.Head.ActorFunc = api.GetActorFunc()
	pack.Head.DstType = api.GetNodeType()
	return nil
}

// 返回设置
func SetRsp(pack *packet.Packet) error {
	if pack.Head.Back != nil {
		return SetRspBack(pack)
	} else if pack.Head.Cmd > 0 {
		return SetRspClient(pack)
	}
	return nil
}

func SetRspBack(pack *packet.Packet) error {
	back := pack.Head.Back
	pack.Head.DstType = back.DstType
	pack.Head.DstId = back.DstId
	pack.Head.ActorId = back.ActorId
	pack.Head.ActorFunc = back.ActorFunc
	return nil
}

func SetRspClient(pack *packet.Packet) error {
	pack.Head.DstType = global.GetGatewayNodeType()
	pack.Head.ActorId = 0
	pack.Head.ActorFunc = 0
	return nil
}

// 路由设置
func RandRouting(pack *packet.Packet) error {
	if len(pack.List) <= 0 {
		pack.List = append(pack.List, &packet.Router{Id: pack.Head.Uid})
	}

	// 进行路由计算
	actorId := tplutil.Or(pack.Head.ActorId > 0, pack.Head.ActorId, pack.Head.Uid)
	for i, item := range pack.List {
		if i == 0 {
			node := cluster.HashRoute(pack.Head.DstType, actorId)
			if node == nil {
				return fmt.Errorf("集群(%d)节点为空", pack.Head.DstType)
			} else {
				pack.Head.DstId = node.Id
			}
		}
		routeItem := router.GetOrNew(item.IdType, item.Id)
		routeItem.Set(pack.Head.DstType, pack.Head.DstId)
		item.List = routeItem.GetNodes()
	}
	return nil
}

// 缓存路由设置
func CacheRouting(pack *packet.Packet) error {
	if len(pack.List) <= 0 {
		pack.List = append(pack.List, &packet.Router{Id: pack.Head.Uid})
	}

	// 进行路由计算
	for i, item := range pack.List {
		routeItem := router.GetOrNew(item.IdType, item.Id)
		if i == 0 {
			nodeId := routeItem.Get(pack.Head.DstType)
			if cluster.Get(pack.Head.DstType, nodeId) == nil {
				return fmt.Errorf("玩家路由信息不存在 uid=%d", pack.Head.Uid)
			}
			pack.Head.DstId = nodeId
		}
		item.List = routeItem.GetNodes()
	}
	return nil
}

package rpc

import (
	"fmt"

	"github.com/hechh/framework/library/tplutil"
	"github.com/hechh/framework/packet"
)

type IRpc interface {
	GetNodeType() uint32        // 获取节点类型
	GetActorFuncName() string   // 获取处理函数名称
	GetActorFunc() uint32       // 获取处理函数唯一ID
	GetCmd() uint32             // 获取命令ID
	GetMask() uint32            // 获取屏蔽模式
	News([]byte) ([]any, error) // 创建请求/响应消息实例
}

type RpcMgr struct {
	names map[tplutil.Tuple2[uint32, string]]IRpc
	apis  map[tplutil.Tuple2[uint32, uint32]]IRpc
	cmds  map[uint32]IRpc
}

func NewRpcMgr() *RpcMgr {
	return &RpcMgr{
		names: make(map[tplutil.Tuple2[uint32, string]]IRpc),
		apis:  make(map[tplutil.Tuple2[uint32, uint32]]IRpc),
		cmds:  make(map[uint32]IRpc),
	}
}

func (m *RpcMgr) Register(rpc IRpc) {
	// 注册rpc
	nodeType := rpc.GetNodeType()
	name := rpc.GetActorFuncName()
	if nodeType != 0 && name != "" {
		nameKey := tplutil.T2(nodeType, name)
		if _, ok := m.names[nameKey]; ok {
			panic(fmt.Sprintf("接口重复注册 nodeType:%d, name:%s", nodeType, name))
		}
		m.names[nameKey] = rpc
		m.apis[tplutil.T2(nodeType, rpc.GetActorFunc())] = rpc
	}

	// 注册cmd
	if rpc.GetCmd() > 0 {
		if _, ok := m.cmds[rpc.GetCmd()]; ok {
			panic(fmt.Sprintf("接口重复注册 nodeType:%d, cmd:%d, name:%s", rpc.GetNodeType(), rpc.GetCmd(), rpc.GetActorFuncName()))
		}
		m.cmds[rpc.GetCmd()] = rpc
	}
}

func (m *RpcMgr) GetByCmd(cmd uint32) IRpc {
	if item, ok := m.cmds[cmd]; ok {
		return item
	}
	return nil
}

func (m *RpcMgr) GetByActorFunc(nodeType uint32, actorFunc uint32) IRpc {
	if item, ok := m.apis[tplutil.T2(nodeType, actorFunc)]; ok {
		return item
	}
	return nil
}

func (m *RpcMgr) GetByActorFuncName(nodeType uint32, name string) IRpc {
	if item, ok := m.names[tplutil.T2(nodeType, name)]; ok {
		return item
	}
	return nil
}

func (m *RpcMgr) GetAllCmds() map[uint32]IRpc {
	return m.cmds
}

func (m *RpcMgr) GetRpc(head *packet.Head) IRpc {
	if head.ActorFunc > 0 {
		if item, ok := m.apis[tplutil.T2(head.DstType, head.ActorFunc)]; ok {
			return item
		}
	}
	if item, ok := m.names[tplutil.T2(head.DstType, head.ActorFuncName)]; ok {
		return item
	}
	return nil
}

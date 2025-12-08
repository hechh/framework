package domain

import (
	"framework/core/define"
	"framework/packet"
)

// 编码器
type IEncoder interface {
	Marshal(...any) ([]byte, error)
	Unmarshal([]byte, ...any) error
}

// 开放接口
type IHandler interface {
	IEncoder
	GetType() uint32                          // 节点类型
	GetId() uint32                            // 唯一id
	GetCmd() uint32                           // 对应命令字
	GetName() string                          // handler名字
	Rpc(any, define.IContext, []byte) func()  // 远程调用
	Call(any, define.IContext, ...any) func() // 本地调用
}

// 回复接口
type RspFunc func(*packet.Head, ...any) error

// handler范式
type V0Func[Actor any] func(*Actor, define.IContext) error
type V1Func[Actor any, V1 any] func(*Actor, define.IContext, V1) error
type V2Func[Actor any, V1 any, V2 any] func(*Actor, define.IContext, V1, V2) error
type V3Func[Actor any, V1 any, V2 any, V3 any] func(*Actor, define.IContext, V1, V2, V3) error

type P1Func[Actor any, V1 any] func(*Actor, define.IContext, *V1) error
type P2Func[Actor any, V1 any, V2 any] func(*Actor, define.IContext, *V1, *V2) error
type P3Func[Actor any, V1 any, V2 any, V3 any] func(*Actor, define.IContext, *V1, *V2, *V3) error

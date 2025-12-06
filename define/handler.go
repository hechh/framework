package define

import "framework/packet"

// 开放接口
type IHandler interface {
	GetType() int32                    // handler类型，即节点类型
	GetId() uint32                     // 唯一id
	GetName() string                   // handler名字
	GetCmd() uint32                    // 对应命令字
	Marshal(...any) ([]byte, error)    // 参数序列化
	Unmarshal([]byte, ...any) error    // 参数反序列化
	Rpc(any, IContext, []byte) func()  // 调用封装
	Call(any, IContext, ...any) func() // 调用封装
}

// 本地
type L0Func[Actor any] func(*Actor) error
type L1Func[Actor any, L1 any] func(*Actor, L1) error
type L2Func[Actor any, L1 any, L2 any] func(*Actor, L1, L2) error

// 远程
type Z0Func[Actor any] func(*Actor, IContext) error
type Z1Func[Actor any] func(*Actor, IContext, any) error

type P1Func[Actor any, P1 any] func(*Actor, IContext, *P1) error
type P2Func[Actor any, P1 any, P2 any] func(*Actor, IContext, *P1, *P2) error

type G1Func[Actor any, V1 any] func(*Actor, IContext, V1) error
type G2Func[Actor any, V1 any, V2 any] func(*Actor, IContext, V1, V2) error

// 回复接口
type RspFunc func(*packet.Head, ...any) error

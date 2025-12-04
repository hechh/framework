package define

import (
	"framework/packet"
	"time"
)

const (
	MAX_NODE_TYPE_COUNT = 32  // 节点类型数量
	CLUSTER_BUCKET_SIZE = 256 // 集群桶的数量
	ETCD_GRANT_TTL      = 15
)

// 消息总线
type IBus interface {
	Listen(func(*packet.Head, []byte)) error                                   // 监听广播
	Broadcast(*packet.Head, []byte, ...*packet.Router) error                   // 发送广播
	Read(func(*packet.Head, []byte)) error                                     // 接受请求
	Write(*packet.Head, []byte, ...*packet.Router) error                       // 发送请求
	Reply(func(*packet.Head, []byte)) error                                    // 接受同步请求
	Request(func([]byte) error, *packet.Head, []byte, ...*packet.Router) error // 发送同步请求
	Response(*packet.Head, []byte) error                                       // 应答
}

type IRpc interface {
	Router(idType int32, id uint64, rid uint64)                                                               // 设置路由
	Callback(actorFunc string, actorId uint64) error                                                          // 设置回调
	Rpc(sendType int32, nodeType int32, actorFunc string, actorId uint64, args ...any) (packet.Packet, error) // 远程调用rpc
}

// 框架上下文接口
type IContext interface {
	ILog
	GetUid() uint64                                  // 获取玩家uid
	GetActorId() uint64                              // 获取actor id
	GetActorName() string                            // 获取actor名字
	GetFuncName() string                             // 获取函数名字
	AddDepth(add uint32) uint32                      // 添加调用深度
	CompareAndSwapDepth(old uint32, new uint32) bool // 原词操作
	NewRpc(isOrigin bool) IRpc                       // 创建IRpc接口
}

// 开放接口
type IHandler interface {
	GetType() int32                    // handler类型，即节点类型
	GetId() uint32                     // 唯一id
	GetName() string                   // handler名字
	GetCmd() int32                     // 对应命令字
	Marshal(...any) ([]byte, error)    // 参数序列化
	Unmarshal([]byte, ...any) error    // 参数反序列化
	Rpc(any, IContext, []byte) func()  // 调用封装
	Call(any, IContext, ...any) func() // 调用封装
}

// Actor接口
type IActor interface {
	GetName() string                               // Actor名字
	GetId() uint64                                 // Actor ID
	SetId(uint64)                                  // Actor ID
	Start()                                        // 启动actor任务队列协程
	Stop()                                         // 关闭actor任务队列协程
	Done()                                         // 出发业务关闭通知(业务层需要事项)
	Wait()                                         // 等待actor自行关闭(业务层需要实现)
	Register(IActor, ...int)                       // 派生类自我注册
	RegisterTimer(any, time.Duration, int32) error // 注册定时器
	SendMsg(any, ...any) error                     // 异步调用派生类成员函数
	Send(any, []byte) error                        // 异步调用派生类成员函数
}

// 应答接口
type IRspHead interface {
	SetRspHead(code int32, msg string)
}

// 包处理
type PacketHandler func(packet.Packet) error

// redis操作
type IRedis interface {
	Del(string) error
	Incr(string) (int64, error)
	IncrBy(string, int64) (int64, error)
	Get(string) (string, error)
	Set(string, any) error
	SetNX(string, any) (bool, error)
	SetEX(string, any, time.Duration) error
	MGet(...string) ([]any, error)
	MSet(...any) error
	HGetAll(string) (map[string]string, error)
	HGet(string, string) (string, error)
	HDel(string, ...string) error
	HKeys(string) ([]string, error)
	HIncrBy(string, string, int64) (int64, error)
	HMSet(string, ...any) error
}

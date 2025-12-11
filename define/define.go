package define

import (
	"framework/packet"
	"net"
	"time"
)

const (
	MAX_NODE_TYPE_COUNT = 32  // 节点类型数量
	CLUSTER_BUCKET_SIZE = 256 // 集群桶的数量
	ETCD_GRANT_TTL      = 15
)

// 服务发现接口
type IWatcher interface {
	Watch(func(string, []byte)) error // 监听k-v变更
	Close()                           // 关闭监听服务
}

// 服务注册接口
type IRegister interface {
	Register(string, []byte) error // 注册服务节点
	Close()                        // 关闭服务注册服务
}

// 集群接口
type ICluster interface {
	Size() int                       // 集群节点数量
	Add(node *packet.Node)           // 添加节点
	Get(nodeId uint32) *packet.Node  // 获取节点
	Del(nodeId uint32) *packet.Node  // 删除节点
	Random(seed uint64) *packet.Node // 路由一个节点
}

// 序列化
type ISerialize interface {
	Marshal(...any) ([]byte, error)
	Unmarshal([]byte, ...any) error
}

// 路由接口
type IRouter interface {
	ISerialize               // 序列化
	GetStatus() bool         // 设置变更标记位
	SetStatus(bool)          // 是否需要保存
	GetIdType() uint32       // 获取路由类型
	GetId() uint64           // 获取路由id
	Get(uint32) uint32       // 获取路由
	Set(uint32, uint32)      // 设置路由
	GetRouter() []uint32     // 获取路由
	SetRouter(...uint32)     // 设置路由
	Update()                 // 更新路由有效时间
	IsExpire(now int64) bool // 是否过期
}

// 消息队列接口
type IMsgQueue interface {
	Subscribe(topic string, handle func(*packet.Message)) error     // 读取消息
	Send(topic string, body []byte) error                           // 发送消息
	Request(topic string, body []byte, cb func([]byte) error) error // 发送同步消息
	Response(topic string, body []byte) error                       // 回复同步消息
	Close()                                                         // 关闭消息总线服务
}

// 消息总线
type IBus interface {
	Broadcast(*packet.Packet) error                   // 发送广播
	Send(*packet.Packet) error                        // 发送请求
	Request(*packet.Packet, func([]byte) error) error // 发送同步请求
	Response(*packet.Head, []byte) error              // 应答
}

// 远程请求接口
type IPacket interface {
	Router(isOrigin bool, routerId uint64, args ...any) IPacket
	Callback(actorFunc string, actorId uint64) IPacket
	Rpc(nodeType uint32, actorId uint64, api string, args ...any) IPacket
	GetPacket() (*packet.Packet, error)
}

// 通用上下文接口
type IContext interface {
	GetHead() *packet.Head                           // 获取包头
	GetIdType() uint32                               // 获取id类型
	GetId() uint64                                   // 获取玩家uid
	GetActorId() uint64                              // 获取actor id
	GetActorName() string                            // 获取actor名字
	GetFuncName() string                             // 获取函数名字
	AddDepth(add uint32) uint32                      // 添加调用深度
	CompareAndSwapDepth(old uint32, new uint32) bool // 原词操作
	Tracef(fmt string, args ...any)                  // 输出trace日志
	Debugf(fmt string, args ...any)                  // 输出debug日志
	Warnf(fmt string, args ...any)                   // 输出warn日志
	Infof(fmt string, args ...any)                   // 输出info日志
	Errorf(fmt string, args ...any)                  // 输出error日志
	Fatalf(fmt string, args ...any)                  // 输出fatal日志
}

// 开放接口
type IHandler interface {
	ISerialize
	GetType() uint32                   // 节点类型
	GetId() uint32                     // 唯一id
	GetCmd() uint32                    // 对应命令字
	GetName() string                   // handler名字
	Rpc(any, IContext, []byte) func()  // 远程调用
	Call(any, IContext, ...any) func() // 本地调用
}

// handler范式
type V0Func[Actor any] func(*Actor, IContext) error
type V1Func[Actor any, V1 any] func(*Actor, IContext, V1) error
type V2Func[Actor any, V1 any, V2 any] func(*Actor, IContext, V1, V2) error
type V3Func[Actor any, V1 any, V2 any, V3 any] func(*Actor, IContext, V1, V2, V3) error

type P1Func[Actor any, V1 any] func(*Actor, IContext, *V1) error
type P2Func[Actor any, V1 any, V2 any] func(*Actor, IContext, *V1, *V2) error
type P3Func[Actor any, V1 any, V2 any, V3 any] func(*Actor, IContext, *V1, *V2, *V3) error

// Actor接口
type IActor interface {
	Start()                                             // 启动actor任务队列协程
	Stop()                                              // 关闭actor任务队列协程
	Done()                                              // 出发业务关闭通知(业务层需要事项)
	Wait()                                              // 等待actor自行关闭(业务层需要实现)
	GetActorName() string                               // Actor名字
	GetActorId() uint64                                 // Actor ID
	SetActorId(uint64)                                  // Actor ID
	Register(IActor, ...int)                            // 派生类自我注册
	RegisterTimer(IContext, time.Duration, int32) error // 注册定时器
	SendMsg(IContext, ...any) error                     // 异步调用派生类成员函数
	Send(IContext, []byte) error                        // 异步调用派生类成员函数
}

// 数据包编码or解码
type IFrame interface {
	Encode(*packet.Packet) []byte
	Decode([]byte) *packet.Packet
}

// socket接口
type ISocket interface {
	Init(net.Conn, IFrame)
	Close()
	Stop()
	GetId() uint32
	Read(func(*packet.Packet) error)
	Write(*packet.Packet) error
}

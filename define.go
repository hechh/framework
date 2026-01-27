package framework

import (
	"net"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/hechh/framework/packet"
)

const (
	MAX_NODE_TYPE_COUNT = 32  // 节点类型数量
	CLUSTER_BUCKET_SIZE = 256 // 集群桶的数量
	ETCD_GRANT_TTL      = 15
)

type IEnum interface {
	Integer() uint32
}

type ISerialize interface {
	Marshal(...any) ([]byte, error)
	Unmarshal([]byte, ...any) error
}

// 消息队列接口
type IMessage interface {
	Subscribe(topic string, handle func(*packet.Message)) error     // 读取消息
	Send(topic string, body []byte) error                           // 发送消息
	Request(topic string, body []byte, cb func([]byte) error) error // 发送同步消息
	Response(topic string, body []byte) error                       // 回复同步消息
	Close()                                                         // 关闭消息总线服务
}

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

type IResponse interface {
	proto.Message
	SetRspHead(code int32, msg string)
	GetRspHead() (int32, string)
}

type IContext interface {
	To(string) IContext
	Copy() IContext
	GetHead() *packet.Head
	GetId() uint64
	GetActorId() uint64
	GetActorName() string
	GetActorFunc() string
	AddDepth(int32) int32
	CompareAndSwapDepth(int32, int32) bool
	Tracef(string, ...any)
	Debugf(string, ...any)
	Warnf(string, ...any)
	Infof(string, ...any)
	Errorf(string, ...any)
	Fatalf(string, ...any)
}

type IActor interface {
	Start()                                           // 启动actor任务队列协程
	Stop()                                            // 关闭actor任务队列协程
	Done()                                            // 出发业务关闭通知(业务层需要事项)
	Wait()                                            // 等待actor自行关闭(业务层需要实现)
	GetActorName() string                             // Actor名字
	GetActorId() uint64                               // Actor ID
	SetActorId(uint64)                                // Actor ID
	Register(IActor, ...int)                          // 派生类自我注册
	RegisterTimer(string, time.Duration, int32) error // 注册定时器
	SendMsg(IContext, ...any) error                   // 异步调用派生类成员函数
	Send(IContext, []byte) error                      // 异步调用派生类成员函数
}

type IRpc interface {
	ISerialize
	GetName() string     // handler名字
	GetCrc32() uint32    // 唯一id
	GetNodeType() uint32 // 节点类型
	GetCmd() uint32      // 对应命令字
	New(int) any         // 请求
}

type IHandler interface {
	ISerialize
	GetName() string                   // handler名字
	GetCrc32() uint32                  // 唯一id
	Rpc(any, IContext, []byte) func()  // 远程调用
	Call(any, IContext, ...any) func() // 本地调用
}

type IFrame interface {
	Decode([]byte) (*packet.Packet, error)
	Encode(*packet.Packet) []byte
}

// socket接口
type ISocket interface {
	Init(net.Conn, IFrame)
	Close()
	GetId() uint32
	IsExpire(int64, int64) bool
	Read(func(*packet.Packet) error)
	Write(*packet.Packet) error
}

type EmptyFunc[Actor any] func(*Actor, IContext) error
type P1Func[Actor any, V any] func(*Actor, IContext, *V) error
type P2Func[Actor any, V any, R any] func(*Actor, IContext, *V, *R) error
type V1Func[Actor any, V any] func(*Actor, IContext, V) error
type V2Func[Actor any, V any, R any] func(*Actor, IContext, V, R) error
type PacketFunc func(*packet.Packet) error
type SaveRouterFunc func(map[string]IRouter) error

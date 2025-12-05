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

// 日志接口
type ILog interface {
	Tracef(fmt string, args ...any) // 输出trace日志
	Debugf(fmt string, args ...any) // 输出debug日志
	Warnf(fmt string, args ...any)  // 输出warn日志
	Infof(fmt string, args ...any)  // 输出info日志
	Errorf(fmt string, args ...any) // 输出error日志
	Fatalf(fmt string, args ...any) // 输出fatal日志
}

// 数据包编码和解码
type IFrame interface {
	Encode(packet.Packet) []byte
	Decode([]byte) packet.Packet
}

// 服务发现接口
type IWatcher interface {
	Watch(func(key string, value []byte)) error // 监听k-v变更
	Close()                                     // 关闭监听服务
}

// 服务注册接口
type IRegister interface {
	Register(key string, value []byte) error // 注册服务节点
	Close()                                  // 关闭服务注册服务
}

// 消息队列接口
type IMsgQueue interface {
	Read(topic string, handle func(packet.Message)) error           // 读取消息
	Write(topic string, body []byte) error                          // 发送消息
	Request(topic string, body []byte, cb func([]byte) error) error // 发送同步消息
	Response(topic string, body []byte) error                       // 回复同步消息
	Close()                                                         // 关闭消息总线服务
}

// 集群接口
type ICluster interface {
	Size() int                       // 集群节点数量
	Add(node *packet.Node)           // 添加节点
	Get(nodeId int32) *packet.Node   // 获取节点
	Del(nodeId int32) *packet.Node   // 删除节点
	Random(seed uint64) *packet.Node // 路由一个节点
}

// 路由接口
type IRouter interface {
	GetType() int32                        // 获取路由类型
	GetId() uint64                         // 获取路由id
	Get(nodeType int32) int32              // 获取路由
	Set(nodeType int32, nodeId int32)      // 设置路由
	GetRouter() []uint32                   // 获取路由
	SetRouter(...uint32)                   // 设置路由
	Update()                               // 更新路由有效时间
	IsExpire(now int64, expire int64) bool // 是否过期
	Change()                               // 设置变更标记位
	IsChange() bool                        // 是否需要保存
	Save()                                 // 保存变更数据
	CopyTo(any)                            // 拷贝函数
}

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

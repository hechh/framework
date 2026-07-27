package define

import "github.com/hechh/framework/packet"

const (
	LOG_MASK        = 1 << 0 // 日志屏蔽模式
	UPDATETIME_MASK = 1 << 1 // 时间屏蔽模式
	CMD_FLAG        = 1 << 2 // 客户端交互命令
	NOTIFY_FLAG     = 1 << 3 // 推送消息
)

const (
	ACTOR_CACHE_FLAG  = 1 << 0 // actor缓存数据
	REDIS_GLOBAL_FLAG = 1 << 1 // 全局数据库
	REDIS_SHARDS_FLAG = 1 << 2 // 分片数据据库
)

const (
	CLUSTER_MASK         = (1 << 6)                    // 集群模式 0x40
	GATEWAY_MASK         = (1 << 7)                    // 网关模式 0x80
	CLUSTER_GATEWAY_MASK = CLUSTER_MASK | GATEWAY_MASK // 集群网关模式 0xC0
)

type IContext interface {
	ICache
	ILogger
	IHead
	Destroy()
	ReadOnly() *packet.Head
	Header(...func(*packet.Head)) *packet.Head // 转发
	Clone(...func(*packet.Head)) *packet.Head  // 派生
	AddDepth(int32) int32
	GetDepth() int32
}

type ICache interface {
	SetCache(string, any, uint32)
	GetCache(string) (any, bool)
	DelCache(string)
	IsChanged(string) bool
	Reset(string)
	Change(string)
}

type ILogger interface {
	Tracef(string, ...any)
	Debugf(string, ...any)
	Warnf(string, ...any)
	Infof(string, ...any)
	Errorf(string, ...any)
	Fatalf(string, ...any)
	Trace(...any)
	Debug(...any)
	Warn(...any)
	Info(...any)
	Error(...any)
	Fatal(...any)
}

type IHead interface {
	GetSendType() packet.SendType
	GetSrcType() uint32
	GetSrcId() uint32
	GetDstType() uint32
	GetDstId() uint32
	GetUid() uint64
	GetCmd() uint32
	GetSeq() uint32
	GetActorId() uint64
	GetActorFunc() uint32
	GetSocketId() uint32
	GetVersion() uint32
	GetCreateTime() int64
	GetBack() *packet.Callback
	GetReply() string
	GetActorFuncName() string
	GetClientIp() string
	SetSendType(packet.SendType)
	SetSrcType(uint32)
	SetSrcId(uint32)
	SetDstType(uint32)
	SetDstId(uint32)
	SetUid(uint64)
	SetCmd(uint32)
	SetSeq(uint32)
	SetActorId(uint64)
	SetActorFunc(uint32)
	SetSocketId(uint32)
	SetVersion(uint32)
	SetCreateTime(int64)
	SetBack(*packet.Callback)
	SetReply(string)
	SetActorFuncName(string)
	SetClientIp(string)
}

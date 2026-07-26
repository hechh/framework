package constant

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

const (
	STOPPED_STATUS = 0 // 已停止
	WAITING_STATUS = 1 // 等待启动中
	RUNNING_STATUS = 2 // 运行中
)

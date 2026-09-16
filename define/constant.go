package define

const (
	LOG_MASK        = 1 << 0 // 日志屏蔽模式
	UPDATETIME_MASK = 1 << 1 // 时间屏蔽模式
	CMD_FLAG        = 1 << 2 // 客户端交互命令
	NOTIFY_FLAG     = 1 << 3 // 推送消息
)

package define

const (
	MAX_NODE_TYPE_COUNT = 32 // 节点类型数量
)

type IContext interface {
	GetUid() uint64                                  // 获取玩家uid
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

type IRpc interface {
	SetRouter(idType uint32, id uint64, routerId uint64, isOrigin bool)
	SetCallback(actorFunc string, actorId uint64) error
	Rpc(nodeType uint32, actorId uint64, api string, args ...any) error
}

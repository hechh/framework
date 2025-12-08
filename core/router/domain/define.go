package domain

const (
	MAX_NODE_TYPE_COUNT = 32 // 节点类型数量
)

type IData interface {
	GetStatus() bool          // 设置变更标记位
	SetStatus(bool)           // 是否需要保存
	Marshal() (string, error) // 序列化
	Unmarshal(string) error   // 反序列化
}

type IRouter interface {
	IData
	GetIdType() uint32                     // 获取路由类型
	GetId() uint64                         // 获取路由id
	Get(uint32) uint32                     // 获取路由
	Set(uint32, uint32)                    // 设置路由
	GetRouter() []uint32                   // 获取路由
	SetRouter(...uint32)                   // 设置路由
	Update()                               // 更新路由有效时间
	IsExpire(now int64, expire int64) bool // 是否过期
}

type FilterFunc func(IRouter) bool        // 过滤函数
type NewFunc func(uint32, uint64) IRouter // 创建路由函数

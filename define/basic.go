package define

import "framework/packet"

/*********************************************************
****************框架基础接口，必须实现**********************
*********************************************************/

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

// 数据同步接口
type IDatabase interface {
	Change()        // 设置变更标记位
	IsChange() bool // 是否需要保存
	Save()          // 保存变更数据
	CopyTo(any)     // 拷贝函数
}

// 路由接口
type IRouter interface {
	IDatabase
	GetType() int32                        // 获取路由类型
	GetId() uint64                         // 获取路由id
	Get(nodeType int32) int32              // 获取路由
	Set(nodeType int32, nodeId int32)      // 设置路由
	GetRouter() []uint32                   // 获取路由
	SetRouter(...uint32)                   // 设置路由
	Update()                               // 更新路由有效时间
	IsExpire(now int64, expire int64) bool // 是否过期
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

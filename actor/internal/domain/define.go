package domain

import (
	"github.com/hechh/framework/common/define"
	"github.com/hechh/framework/packet"
	"strings"
	"sync/atomic"
	"time"

	"github.com/hechh/library/base/templ"
)

const (
	STOPPED_STATUS = 0 // 已停止
	WAITING_STATUS = 1 // 等待启动中
	RUNNING_STATUS = 2 // 运行中
)

type ITask interface {
	Do()
	GetMask() uint32
}

type IAsync interface {
	GetName() string       // Actor名字
	GetSize() int          // 获取数量
	GetId() uint64         // Actor ID
	SetId(uint64)          // Actor ID
	GetIdPointer() *uint64 // 获取指针
	Start() bool           // 启动actor任务队列协程
	Stop()                 // 关闭actor任务队列协程
	Wait()                 // 等待actor协程退出
	Push(ITask) bool       // 推送任务
}

type IActor interface {
	define.ICache                                     // 业务层必须实现
	Init() error                                      // 业务层必须实现
	Close()                                           // 业务层必须实现
	GetName() string                                  // Actor名字
	GetId() uint64                                    // Actor ID
	SetId(uint64)                                     // Actor ID
	Start() bool                                      // 启动actor任务队列协程
	Stop()                                            // 关闭actor任务队列协程
	Wait()                                            // 等待actor协程退出
	GetActor(uint64) IActor                           // 获取自身
	Register(IActor, ...Option)                       // 派生类自我注册
	RegisterTimer(string, time.Duration, int32) error // 注册定时器
	SendMsg(*packet.Head, ...any) error               // 异步调用派生类成员函数
	Send(*packet.Head, []byte) error                  // 异步调用派生类成员函数
}

var (
	actorIdGenerator = uint64(0)
)

func GenActorId() uint64 {
	return atomic.AddUint64(&actorIdGenerator, 1)
}

func GetActorId(head *packet.Head) uint64 {
	return templ.Or(head.ActorId > 0, head.ActorId, head.Uid)
}

func GetActorName(name string) string {
	pos := strings.Index(name, ".")
	if pos == -1 {
		return name
	}
	return name[:pos]
}

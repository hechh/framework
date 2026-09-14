package queue

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/hechh/framework/library/safe"
)

type MsgQueuePool[T ITask] struct {
	*Attribute                // 基础
	tasks      *Queue[T]      // 任务队列
	notifyCh   chan struct{}  // 通知
	exitCh     chan struct{}  // 退出
	taskCh     chan T         // 任务抢占队列
	updateTime int64          // 更新时间
	lockTime   int64          // 更新时间
	w1         sync.WaitGroup // 等待 run goroutine 退出
	w2         sync.WaitGroup // 等待 worker goroutine 退出
}

func NewMsgQueuePool[T ITask](opts ...Option) *MsgQueuePool[T] {
	attr := new(Attribute)
	for _, opt := range opts {
		opt(attr)
	}
	if attr.id <= 0 {
		attr.id = GenId()
	}
	return &MsgQueuePool[T]{
		Attribute: attr,
		tasks:     NewQueue[T](),
		notifyCh:  make(chan struct{}, 1),
		exitCh:    make(chan struct{}),
		taskCh:    make(chan T, 5*attr.size),
	}
}

func (d *MsgQueuePool[T]) Start(except func(string, ...any)) error {
	// 原子迁移 STOPPED → STARTING：重复/并发 Start 只允许一个成功（原因见 MsgQueue.Start）
	if !d.CompareAndSwapStatus(STOPPED, STARTING) {
		return fmt.Errorf("消息队列池(%s)启动失败: 当前状态(%d)不是已停止", d.GetName(), d.GetStatus())
	}
	// 先抢占锁（锁被其他节点持有时返回错误，属正常分支）
	if err := d.OnLock(); err != nil {
		// 回滚状态：否则永久停留在 STARTING，Start 静默失效且 Stop/Wait 永久不可用
		d.Closed()
		return err
	}
	// 启动协程池
	for range d.size {
		d.w2.Add(1)
		safe.SafeGo(except, func() {
			// 退出时才计数：w2 表示"worker 已排空并退出"。
			// 此前写在函数首条语句，导致 Wait() 里的 w2.Wait() 立即返回，
			// 在 worker 仍在消费 taskCh 时就宣告关闭完成。
			defer d.w2.Done()
			for task := range d.taskCh {
				if task.Do() {
					atomic.StoreInt64(&d.updateTime, time.Now().Unix())
				}
			}
		})
	}
	// 启动协程
	d.w1.Add(1)
	safe.SafeGo(except, d.run)
	// 正在运行
	d.Running()
	return nil
}

func (d *MsgQueuePool[T]) Stop() {
	// 原子地将 RUNNING → WAITING，仅切换成功者执行关闭动作。
	// OnDelete 回调可能重入 Stop 或存在并发调用，CAS 保证 exitCh 只 close 一次。
	if atomic.CompareAndSwapInt32(&d.status, RUNNING, WAITING) {
		close(d.exitCh)
		// 删除
		d.OnDelete()
	}
}

// Wait 等待 run 与全部 worker 退出后关闭对象。
// 注意：只应由单一调用者触发（当前经 gc.Destroy 串行调用），
// 因为 close(taskCh) 不可重复执行。
func (d *MsgQueuePool[T]) Wait() {
	if d.IsWaiting() {
		// w1 归零 = run 已完全退出（含 handle 收尾与 OnUnlock），此后 taskCh 不会再有发送者
		d.w1.Wait()
		d.SetId(0)
		// 关闭任务通道：worker 排空缓冲内的剩余任务后退出
		close(d.taskCh)
		// 等待 worker 全部退出
		d.w2.Wait()
		d.Closed()
	}
}

func (d *MsgQueuePool[T]) Push(t T) (flag bool) {
	if flag = d.IsRunning(); flag {
		d.tasks.Push(t, func() {
			select {
			case d.notifyCh <- struct{}{}:
			default:
			}
		})
	}
	return flag
}

func (d *MsgQueuePool[T]) run() {
	// id 快照：解锁必须用启动时的 actor id（不能读共享字段，见下方收尾顺序说明）
	id := d.GetId()
	tt := time.NewTicker(time.Second)
	// 收尾顺序是正确性约束，不要调整：
	//  1) OnUnlock 必须早于 w1.Done() 完成 —— Wait() 在 w1 归零后立即 SetId(0) 并 close(taskCh)，
	//     若解锁晚于 SetId(0)，解锁会拿到 id=0，真实的 actor_locker:<uid> 最长 TTL 内无人释放
	//  2) w1.Done() 必须是最后一条 —— 否则 Wait() 会在 handle() 仍向 taskCh 发送时 close(taskCh)，
	//     触发 panic: send on closed channel，且 panic 会跳过解锁
	//  3) 嵌套 defer 兜住 handle() panic —— 否则解锁被跳过，锁同样泄漏到 TTL
	defer func() {
		defer func() {
			_ = d.onUnlockWith(id)
			d.w1.Done()
		}()
		tt.Stop()  // 关闭定时器
		d.Stop()   // 发送停止消息
		d.handle() // 处理剩余请求
	}()
	d.updateTime = time.Now().Unix()
	d.lockTime = d.updateTime
	lockExpire := d.expireTime * 2 / 3

	// 循环处理任务
	for {
		select {
		case <-d.notifyCh:
			d.handle()
		case <-d.exitCh:
			return
		case tnow := <-tt.C:
			if d.idleTime > 0 && tnow.Unix()-atomic.LoadInt64(&d.updateTime) > d.idleTime {
				return
			}
			if d.expireTime > 0 && tnow.Unix()-d.lockTime >= lockExpire {
				if err := d.OnLock(); err != nil {
					return
				}
				d.lockTime = tnow.Unix()
			}
		}
	}
}

func (d *MsgQueuePool[T]) handle() {
	for range 100 {
		f, ok := d.tasks.Pop()
		if !ok {
			return
		}
		d.taskCh <- f
	}
	if d.tasks.GetCount() > 0 {
		select {
		case d.notifyCh <- struct{}{}:
		default:
		}
	}
}

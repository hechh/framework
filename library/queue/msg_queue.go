package queue

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/hechh/framework/library/safe"
)

type ITask interface {
	Do() bool
}

type MsgQueue[T ITask] struct {
	*Attribute                // 基础
	tasks      *Queue[T]      // 任务队列
	notifyCh   chan struct{}  // 通知
	exitCh     chan struct{}  // 退出
	updateTime int64          // 更新时间
	lockTime   int64          // 全局锁保活
	wg         sync.WaitGroup // 等待任务完成
}

func NewMsgQueue[T ITask](opts ...Option) *MsgQueue[T] {
	attr := new(Attribute)
	for _, opt := range opts {
		opt(attr)
	}
	if attr.id <= 0 {
		attr.id = GenId()
	}
	return &MsgQueue[T]{
		Attribute: attr,
		tasks:     NewQueue[T](),
		notifyCh:  make(chan struct{}, 1),
		exitCh:    make(chan struct{}),
	}
}

func (d *MsgQueue[T]) Start(except func(string, ...any)) error {
	// 原子迁移 STOPPED → STARTING：重复/并发 Start 只允许一个成功。
	// 此前用 IsClosed() 判断再 Starting()，处于 STARTING/WAITING 时会静默返回 nil，
	// 调用方会误以为启动成功（例如 ActorMgr.Start 重试时把"没起来"当成"起来了"）。
	if !d.CompareAndSwapStatus(STOPPED, STARTING) {
		return fmt.Errorf("消息队列(%s)启动失败: 当前状态(%d)不是已停止", d.GetName(), d.GetStatus())
	}
	// 先抢占锁（锁被其他节点持有时返回错误，属正常分支）
	if err := d.OnLock(); err != nil {
		// 回滚状态：否则永久停留在 STARTING —— 后续 Start 被挡掉且不报错，
		// Stop 的 RUNNING→WAITING CAS 必然失败，exitCh/OnDelete/Wait 全部永久失效
		d.Closed()
		return err
	}
	// 启动协程
	d.wg.Add(1)
	safe.SafeGo(except, d.run)
	// 正在运行
	d.Running()
	return nil
}

func (d *MsgQueue[T]) Stop() {
	// 原子地将 RUNNING → WAITING，仅切换成功者执行关闭动作。
	// OnDelete 回调（如 PlayerMgr.remove → Actor.Stop）可能在 Waiting 之前
	// 重入 Stop，或存在并发调用；CAS 保证 exitCh 只被 close 一次，
	// 避免 "close of closed channel" panic。
	if atomic.CompareAndSwapInt32(&d.status, RUNNING, WAITING) {
		close(d.exitCh)
		// 删除
		d.OnDelete()
	}
}

func (d *MsgQueue[T]) Wait() {
	if d.IsWaiting() {
		// 等待结束：run 的收尾以 wg.Done() 结尾，返回时 OnUnlock 必已完成
		d.wg.Wait()
		// 关闭定时器
		d.SetId(0)
		// 修改状态
		d.Closed()
	}
}

func (d *MsgQueue[T]) Push(t T) (flag bool) {
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

func (d *MsgQueue[T]) run() {
	// id 快照：解锁必须用启动时的 actor id（不能读共享字段，见下方收尾顺序说明）
	id := d.GetId()
	tt := time.NewTicker(time.Second)
	// 收尾顺序是正确性约束，不要调整：
	//  1) OnUnlock 必须早于 wg.Done() 完成 —— Wait() 在 wg 归零后立即 SetId(0)，
	//     若解锁晚于 SetId(0)，解锁会拿到 id=0，真实的 actor_locker:<uid> 最长 TTL 内无人释放（玩家无法重登/被接管）
	//  2) wg.Done() 必须是最后一条 —— 否则 Wait() 会在 handle()、解锁仍在执行时就返回
	//  3) 嵌套 defer 兜住 handle() panic —— 否则解锁被跳过，锁同样泄漏到 TTL
	defer func() {
		defer func() {
			_ = d.onUnlockWith(id)
			d.wg.Done()
		}()
		tt.Stop()
		d.Stop()
		d.handle()
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
			if d.idleTime > 0 && tnow.Unix()-d.updateTime > d.idleTime {
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

func (d *MsgQueue[T]) handle() {
	for range 100 {
		f, ok := d.tasks.Pop()
		if !ok {
			return
		}
		if f.Do() {
			d.updateTime = time.Now().Unix()
		}
	}
	if d.tasks.GetCount() > 0 {
		select {
		case d.notifyCh <- struct{}{}:
		default:
		}
	}
}

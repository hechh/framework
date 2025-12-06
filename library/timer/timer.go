package timer

import (
	"fmt"
	"framework/library/async"
	"framework/library/uerror"
	"framework/library/util"

	"sort"
	"sync/atomic"
	"time"
)

var (
	timerObj = NewTimer(4, 5)
)

func Register(taskId *uint64, ttl time.Duration, times int32, f func()) error {
	return timerObj.Register(taskId, ttl, times, f)
}

func Close() {
	timerObj.Close()
}

type Task struct {
	event  func()
	id     *uint64
	ttl    int64
	times  int32
	expire int64
	next   *Task
}

type Wheel struct {
	mask    int64
	shift   int64
	cursor  int64
	buckets []*Task
}

type Timer struct {
	startTime int64
	lastTime  int64
	caches    []*Task
	wheels    []*Wheel
	head      *Wheel
	tail      *Wheel
	tasks     *async.Queue[*Task]
	notify    chan struct{}
	exit      chan struct{}
}

func NewTimer(tick int64, size int) *Timer {
	nowMs := time.Now().UnixMilli()
	wls := []*Wheel{}
	for i := 0; i < size; i++ {
		bit := util.Or[int64](i == 0, 12, 5)
		wls = append(wls, &Wheel{mask: 1<<bit - 1, shift: tick, cursor: nowMs, buckets: make([]*Task, 1<<int(bit))})
		tick += bit
	}
	ret := &Timer{
		startTime: nowMs,
		lastTime:  nowMs,
		wheels:    wls,
		head:      wls[0],
		tail:      wls[size-1],
		tasks:     async.NewQueue[*Task](),
		notify:    make(chan struct{}, 1),
		exit:      make(chan struct{}),
	}
	go ret.run()
	return ret
}

// 注册定时器
func (d *Timer) Register(taskId *uint64, ttl time.Duration, times int32, f func()) error {
	tt := int64(ttl / time.Millisecond)
	if tt>>d.head.shift <= 0 {
		return uerror.New(-1, "最小时间间隔必须大于%dms", 1<<d.head.shift)
	}
	if (tt >> d.tail.shift) > d.tail.mask {
		return fmt.Errorf("最大时间间隔必须小于%dms", 1<<d.tail.shift)
	}
	d.tasks.Push(&Task{
		event: f,
		id:    taskId,
		ttl:   tt,
		times: times,
	})
	select {
	case d.notify <- struct{}{}:
	default:
	}
	return nil
}

func (d *Timer) Close() {
	close(d.exit)
}

func (d *Timer) run() {
	tick := int64(1 << d.wheels[0].shift)
	tt := time.NewTicker(time.Duration(tick) * time.Millisecond)
	defer tt.Stop()
	for {
		select {
		case <-d.notify:
			nowMs := atomic.LoadInt64(&d.lastTime)
			for tt := d.tasks.Pop(); tt != nil; tt = d.tasks.Pop() {
				tt.expire = nowMs + tt.ttl
				d.insert(tt)
			}
			d.flush()
		case <-tt.C:
			nowMs := atomic.AddInt64(&d.lastTime, tick)
			d.update(nowMs)
			d.flush()
		case <-d.exit:
			return
		}
	}
}

func (d *Timer) update(nowMs int64) {
	for _, w := range d.wheels {
		tasks := w.Get(nowMs)
		for tt := tasks; tt != nil; tt = tasks {
			tasks = tasks.next
			tt.next = nil
			if d.wheels[0].IsExpire(tt) {
				tt.Handle(nowMs)
			}
			if tt.IsEnable() {
				d.insert(tt)
			}
		}
		if !w.IsCarry() {
			break
		}
	}
}

func (d *Timer) insert(tt *Task) {
	d.caches = append(d.caches, tt)
	if len(d.caches) > 1000 {
		d.flush()
	}
}

func (d *Timer) flush() {
	if lnews := len(d.caches); lnews > 0 {
		sort.Slice(d.caches, func(i, j int) bool {
			return d.caches[i].expire < d.caches[j].expire
		})

		pos := 0
		for _, w := range d.wheels {
			for ; pos < lnews && w.IsMatch(d.caches[pos]); pos++ {
				w.Insert(d.caches[pos])
			}
			if lnews <= pos {
				break
			}
		}
		d.caches = d.caches[:0]
	}
}

// 是否进位
func (w *Wheel) IsCarry() bool {
	return (w.cursor>>w.shift)&w.mask <= 0
}

// 是否过期
func (w *Wheel) IsExpire(tt *Task) bool {
	return tt.expire <= w.cursor || (tt.expire>>w.shift) <= (w.cursor>>w.shift)
}

// 是否匹配
func (w *Wheel) IsMatch(tt *Task) bool {
	return (tt.expire>>w.shift)-(w.cursor>>w.shift) <= w.mask
}

// 插入数据
func (w *Wheel) Insert(tt *Task) {
	pos := (tt.expire >> w.shift) & w.mask
	tt.next = w.buckets[pos]
	w.buckets[pos] = tt
}

// 获取过期定时任务
func (w *Wheel) Get(nowMs int64) *Task {
	pos := (nowMs >> w.shift) & w.mask
	ret := w.buckets[pos]
	w.buckets[pos] = nil
	w.cursor = nowMs
	return ret
}

// 任务是否有效
func (d *Task) IsEnable() bool {
	return d.id != nil && *d.id > 0 && d.times != 0
}

// 执行任务
func (d *Task) Handle(nowMs int64) {
	if d.IsEnable() {
		async.Recover(d.event)
		if d.times > 0 {
			d.times--
		}
		if d.times != 0 {
			d.expire = nowMs + d.ttl
		}
	}
}

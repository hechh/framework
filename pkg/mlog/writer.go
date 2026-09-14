package mlog

import (
	"bytes"
	"fmt"
	"os"
	"path"
	"sync"
	"sync/atomic"
	"time"

	"github.com/hechh/framework/library/fileutil"
	"github.com/hechh/framework/library/queue"
)

// IWriter 写入器接口
type IWriter interface {
	Write(time.Time, []byte) (int, error)
	Close() error
}

type GroupWriter struct {
	list []IWriter
}

func (w *GroupWriter) Write(now time.Time, msg []byte) (n int, err error) {
	var reterr error
	for _, w := range w.list {
		if n, reterr = w.Write(now, msg); reterr != nil {
			fmt.Fprintf(os.Stderr, "Failed to write log, now:%d, error:%v\n", now.Unix(), err)
			err = reterr
		}
	}
	return
}

func (w *GroupWriter) Close() (err error) {
	var reterr error
	for _, writer := range w.list {
		if reterr = writer.Close(); reterr != nil {
			err = reterr
		}
	}
	return
}

type StdoutWriter struct{}

func (w *StdoutWriter) Write(t time.Time, p []byte) (n int, err error) {
	n, err = os.Stdout.Write(p)
	return
}

func (w *StdoutWriter) Close() error {
	return nil
}

type Data struct {
	now  time.Time
	buff *bytes.Buffer
}

// RollingStrategy 日志滚动策略
type RollingStrategy int

const (
	RollingByHour RollingStrategy = iota // 按小时滚动
	RollingByDay                         // 按天滚动
)

// logFallbackInterval 日志写失败告警的限流间隔：第 1 次立即告警，之后每 N 次告警一次
const logFallbackInterval = 1000

// RotateWriter 按小时滚动的日志写入器
type RotateWriter struct {
	strategy      RollingStrategy     // 滚动策略
	cache         *fileutil.Buffer    // 底层缓存
	flushInterval time.Duration       // 自动刷新间隔
	flushChan     chan struct{}       // 手动刷新信号
	dataChan      chan struct{}       // 写入信号
	exitChan      chan struct{}       // 停止信号
	list          *queue.Queue[*Data] // 数据队列
	lpath         string              // 日志路径
	lname         string              // 日志文件名前缀
	wg            sync.WaitGroup      // 等待组
	pendingCount  int32               // 无锁计数器减少通道通知竞争
	lastHour      int                 // 缓存最后一次小时
	lastDay       int                 // 缓存最后一次日期
	lastMonth     time.Month          // 缓存最后一次月份
	lastYear      int                 // 缓存最后一次年份
	dataPool      sync.Pool           // Data对象池
	errCount      atomic.Int64        // 写失败次数（限流降级告警）
}

// New 创建按小时滚动的日志写入器
func NewRotateWriter(lpath, lname string, bufferSize int, flushInterval time.Duration, val RollingStrategy) *RotateWriter {
	w := &RotateWriter{
		strategy:      val,
		cache:         fileutil.NewBuffer(bufferSize),
		flushInterval: flushInterval,
		flushChan:     make(chan struct{}, 1),
		dataChan:      make(chan struct{}, 1),
		exitChan:      make(chan struct{}),
		list:          queue.NewQueue[*Data](),
		lpath:         lpath,
		lname:         lname,
		dataPool: sync.Pool{
			New: func() any {
				return &Data{buff: &bytes.Buffer{}}
			},
		},
	}
	w.wg.Add(1)
	go w.run()
	return w
}

// Write 写入日志数据
func (d *RotateWriter) Write(now time.Time, p []byte) (n int, err error) {
	data := d.dataPool.Get().(*Data)
	data.now = now
	data.buff.Reset()
	data.buff.Write(p)
	n = len(p)

	// 放入队列
	d.list.Push(data, func() {
		if atomic.AddInt32(&d.pendingCount, 1) == 1 {
			select {
			case d.dataChan <- struct{}{}:
			default:
			}
		}
	})
	return
}

// Close 关闭写入器
func (d *RotateWriter) Close() error {
	close(d.exitChan)
	d.wg.Wait()
	return nil
}

func (d *RotateWriter) isRotate(t time.Time) bool {
	if d.strategy == RollingByHour {
		return d.lastYear != t.Year() || d.lastMonth != t.Month() || d.lastDay != t.Day() || d.lastHour != t.Hour()
	}
	return d.lastYear != t.Year() || d.lastMonth != t.Month() || d.lastDay != t.Day()
}

func (d *RotateWriter) getFilename(t time.Time) string {
	d.lastYear = t.Year()
	d.lastMonth = t.Month()
	d.lastDay = t.Day()
	d.lastHour = t.Hour()
	if d.strategy == RollingByHour {
		return path.Join(d.lpath, t.Format("20060102"), fmt.Sprintf("%s-%02d.log", d.lname, t.Hour()))
	}
	return path.Join(d.lpath, t.Format("20060102"), fmt.Sprintf("%s.log", d.lname))
}

// run 主循环
func (d *RotateWriter) run() {
	tt := time.NewTicker(d.flushInterval)
	defer func() {
		tt.Stop()
		d.handler()
		d.flush()
		d.cache.Close()
		d.wg.Done()
	}()

	for {
		select {
		case <-d.dataChan:
			atomic.StoreInt32(&d.pendingCount, 0)
			d.handler()
		case <-tt.C:
			d.flush()
		case <-d.flushChan:
			d.flush()
		case <-d.exitChan:
			return
		}
	}
}

func (d *RotateWriter) handler() {
	for {
		item, ok := d.list.Pop()
		if !ok {
			return
		}
		if d.isRotate(item.now) {
			// 目录不可写/磁盘满/文件被占用时文件无法打开：降级到 stderr，不静默丢日志
			if err := d.cache.Set(d.getFilename(item.now)); err != nil {
				d.fallback(item.buff.Bytes(), err)
				d.dataPool.Put(item)
				continue
			}
		}
		if _, err := d.cache.Write(item.buff.Bytes()); err != nil {
			d.fallback(item.buff.Bytes(), err)
		}
		d.dataPool.Put(item)
	}
}

// flush 刷新缓冲区；失败时缓冲区内容保留待下次重试（fileutil.Buffer 只在成功后清零），
// 但必须告警，否则日志静默丢失
func (d *RotateWriter) flush() {
	if err := d.cache.Flush(); err != nil {
		d.warn("日志刷盘失败", err)
	}
}

// fallback 文件写入失败时把日志降级输出到 stderr：日志目录不可写、磁盘满等故障下
// 日志不能静默丢弃（事故时无日志可查），告警按次数限流避免 stderr 被刷爆
func (d *RotateWriter) fallback(msg []byte, err error) {
	d.warn("日志写入文件失败, 降级输出到 stderr", err)
	_, _ = os.Stderr.Write(msg)
}

// warn 输出降级告警：第 1 次立即告警，之后每 logFallbackInterval 次告警一次
func (d *RotateWriter) warn(reason string, err error) {
	n := d.errCount.Add(1)
	if n == 1 || n%logFallbackInterval == 0 {
		fmt.Fprintf(os.Stderr, "%s %s(第%d次), error:%v\n", d.lname, reason, n, err)
	}
}

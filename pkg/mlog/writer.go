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
		d.cache.Flush()
		d.cache.Close()
		d.wg.Done()
	}()

	for {
		select {
		case <-d.dataChan:
			atomic.StoreInt32(&d.pendingCount, 0)
			d.handler()
		case <-tt.C:
			d.cache.Flush()
		case <-d.flushChan:
			d.cache.Flush()
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
			d.cache.Set(d.getFilename(item.now))
		}
		d.cache.Write(item.buff.Bytes())
		d.dataPool.Put(item)
	}
}

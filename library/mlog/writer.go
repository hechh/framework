package mlog

import (
	"fmt"
	"framework/library/queue"
	"os"
	"path"
	"sync"
	"time"
)

type StdWriter struct{}

func NewStdWriter() *StdWriter {
	return &StdWriter{}
}

func (d *StdWriter) Push(data IData) {
	defer put(data)
	tt := data.Now()
	fmt.Fprint(os.Stdout, tt.Format("2006-01-02 15:04:05")+"\t"+string(data.Read()))
}

func (d *StdWriter) Close() {}

type LogWriter struct {
	sync.WaitGroup
	lpath  string
	lname  string
	cache  *Cache
	datas  *queue.Queue[IData]
	notify chan struct{}
	exit   chan struct{}
}

func NewLogWriter(lpath, lname string) *LogWriter {
	ret := &LogWriter{
		lpath:  lpath,
		lname:  lname,
		cache:  NewCache(1024 * 1024),
		datas:  queue.NewQueue[IData](),
		notify: make(chan struct{}, 1),
		exit:   make(chan struct{}),
	}
	ret.Add(1)
	go ret.run()
	return ret
}

func (d *LogWriter) Push(data IData) {
	d.datas.Push(data)
	select {
	case d.notify <- struct{}{}:
	default:
	}
}

func (d *LogWriter) Close() {
	close(d.exit)
	d.Wait()
}

func (d *LogWriter) run() {
	tt := time.NewTicker(3 * time.Second)
	defer func() {
		tt.Stop()
		for mm := d.datas.Pop(); mm != nil; mm = d.datas.Pop() {
			d.cache.Set(d.getFileName(mm))
			d.cache.Write(mm.Read())
			put(mm)
		}
		d.cache.Flush()
		d.cache.Close()
		d.Done()
	}()

	for {
		select {
		case <-d.notify:
			for mm := d.datas.Pop(); mm != nil; mm = d.datas.Pop() {
				d.cache.Set(d.getFileName(mm))
				d.cache.Write(mm.Read())
				put(mm)
			}
		case <-tt.C:
			d.cache.Flush()
		case <-d.exit:
			return
		}
	}
}

func (d *LogWriter) getFileName(m IData) string {
	tt := m.Now()
	return path.Join(d.lpath, fmt.Sprintf("%s_%04d%02d%02d.log", d.lname, tt.Year(), tt.Month(), tt.Day()))
}

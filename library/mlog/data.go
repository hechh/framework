package mlog

import (
	"bytes"
	"path/filepath"
	"strconv"
	"sync/atomic"
	"time"
)

type Data struct {
	buffer    *bytes.Buffer
	time      time.Time
	reference int32
}

func NewData() IData {
	return &Data{buffer: bytes.NewBuffer(nil)}
}

func (d *Data) Now() time.Time {
	return d.time
}

func (d *Data) Add(delta int32) {
	atomic.AddInt32(&d.reference, delta)
}

func (d *Data) Done() int32 {
	return atomic.AddInt32(&d.reference, -1)
}

func (d *Data) Read() []byte {
	return d.buffer.Bytes()
}

func (d *Data) Write(data Meta) {
	d.buffer.Reset()
	d.time = time.Now()
	d.buffer.WriteByte('[')
	d.buffer.WriteString(d.time.Format("2006-01-02 15:04:05.000"))
	d.buffer.WriteByte(']')
	d.buffer.WriteByte('\t')
	d.buffer.WriteByte('[')
	d.buffer.WriteString(LevelToString(data.Level))
	d.buffer.WriteByte(']')
	d.buffer.WriteByte('\t')
	d.buffer.WriteString(filepath.Base(data.FileName))
	d.buffer.WriteByte(':')
	d.buffer.WriteString(strconv.Itoa(data.Line))
	d.buffer.WriteByte(' ')
	d.buffer.WriteString(data.FuncName)
	d.buffer.WriteByte('\t')
	d.buffer.WriteString(data.Msg)
	d.buffer.WriteByte('\n')
}

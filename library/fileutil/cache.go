package fileutil

import (
	"errors"
	"os"
)

type Buffer struct {
	len      int      // 当前使用量
	cap      int      // 缓冲区总大小
	buffer   []byte   // 缓冲区
	fp       *os.File // 文件操作
	filename string   // 写入文件名
}

func NewBuffer(size int) *Buffer {
	return &Buffer{
		cap:    size,
		buffer: make([]byte, size),
	}
}

func (d *Buffer) Set(filename string) error {
	if d.fp != nil && !IsSameFile(d.fp, filename) {
		if err := d.Flush(); err != nil {
			return err
		}
		if err := d.fp.Close(); err != nil {
			return err
		}
		d.fp = nil
		d.filename = ""
	}
	if d.fp == nil {
		fp, err := CreateFile(filename, os.O_CREATE|os.O_APPEND|os.O_RDWR)
		if err != nil {
			return err
		}
		d.fp = fp
		d.filename = filename
	}
	return nil
}

func (d *Buffer) Write(p []byte) (n int, err error) {
	if d.fp == nil {
		return 0, errors.New("futil: 文件未打开")
	}
	total := 0
	for len(p) > 0 {
		if d.len >= d.cap {
			if err := d.Flush(); err != nil {
				return total, err
			}
		}
		diff := d.cap - d.len
		if len(p) <= diff {
			copy(d.buffer[d.len:], p)
			d.len += len(p)
			total += len(p)
			return total, nil
		}
		copy(d.buffer[d.len:], p[:diff])
		d.len += diff
		total += diff
		p = p[diff:]
	}
	return total, nil
}

func (d *Buffer) Flush() error {
	if d.fp != nil && d.len > 0 {
		if _, err := d.fp.Write(d.buffer[:d.len]); err != nil {
			return err
		}
		d.len = 0
	}
	return nil
}

func (d *Buffer) Close() error {
	if d.fp != nil {
		defer d.fp.Close()
		return d.Flush()
	}
	return nil
}

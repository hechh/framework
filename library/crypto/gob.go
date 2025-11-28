package crypto

import (
	"bytes"
	"encoding/gob"
	"sync"
)

var (
	encPool = sync.Pool{
		New: func() interface{} {
			buf := bytes.NewBuffer(make([]byte, 0, 1024))
			enc := gob.NewEncoder(buf)
			return &GobEncoder{buf: buf, enc: enc}
		},
	}
	decPool = sync.Pool{
		New: func() interface{} {
			buf := bytes.NewBuffer(make([]byte, 0, 1024))
			dec := gob.NewDecoder(buf)
			return &GobDecoder{buf: buf, dec: dec}
		},
	}
)

type GobEncoder struct {
	buf *bytes.Buffer
	enc *gob.Encoder
}

type GobDecoder struct {
	buf *bytes.Buffer
	dec *gob.Decoder
}

// 编码
func Encode(args ...any) ([]byte, error) {
	item := encPool.Get().(*GobEncoder)
	defer encPool.Put(item)
	item.buf.Reset()
	for _, arg := range args {
		if err := item.enc.Encode(arg); err != nil {
			return nil, err
		}
	}
	rets := make([]byte, item.buf.Len())
	copy(rets, item.buf.Bytes())
	return rets, nil
}

// 解码
func Decode(data []byte, args ...any) error {
	item := decPool.Get().(*GobDecoder)
	defer decPool.Put(item)
	item.buf.Reset()
	item.buf.Write(data)
	for _, arg := range args {
		if err := item.dec.Decode(arg); err != nil {
			return err
		}
	}
	return nil
}

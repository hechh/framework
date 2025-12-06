package crypto

import (
	"bytes"
	"encoding/gob"
	"sync"
)

var (
	gobEncoder = sync.Pool{
		New: func() interface{} {
			buf := bytes.NewBuffer(make([]byte, 0, 1024))
			enc := gob.NewEncoder(buf)
			return &GobEncoder{buf: buf, enc: enc}
		},
	}
	gobDecoder = sync.Pool{
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
func GobEncrypto(args ...any) ([]byte, error) {
	item := gobEncoder.Get().(*GobEncoder)
	defer gobEncoder.Put(item)
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
func GobDecrypto(data []byte, args ...any) error {
	item := gobDecoder.Get().(*GobDecoder)
	defer gobDecoder.Put(item)
	item.buf.Reset()
	item.buf.Write(data)
	for _, arg := range args {
		if err := item.dec.Decode(arg); err != nil {
			return err
		}
	}
	return nil
}

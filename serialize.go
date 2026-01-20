package framework

import (
	"github.com/hechh/library/crypto"
	"github.com/hechh/library/uerror"
	"google.golang.org/protobuf/proto"
)

var (
	PROTO = &ProtoSerialize{}
	GOB   = &GobSerialize{}
	BYTES = &BytesSerialize{}
	EMPTY = &EmptySerialize{}
	ERROR = &ErrorSerialize{}
)

type GobSerialize struct{}

func (d *GobSerialize) Marshal(args ...any) ([]byte, error) {
	return crypto.GobEncrypto(args...)
}

func (d *GobSerialize) Unmarshal(buf []byte, args ...any) error {
	return crypto.GobDecrypto(buf, args...)
}

type ProtoSerialize struct{}

func (d *ProtoSerialize) Marshal(args ...any) ([]byte, error) {
	return proto.Marshal(args[0].(proto.Message))
}

func (d *ProtoSerialize) Unmarshal(buf []byte, args ...any) error {
	return proto.Unmarshal(buf, args[0].(proto.Message))
}

type BytesSerialize struct{}

func (d *BytesSerialize) Marshal(args ...any) ([]byte, error) {
	buf, ok := args[0].([]byte)
	if !ok {
		return nil, uerror.New(-1, "参数类型错误")
	}
	return buf, nil
}

func (d *BytesSerialize) Unmarshal(buf []byte, args ...any) error {
	dst, ok := args[0].(*[]byte)
	if !ok {
		return uerror.New(-1, "参数类型错误")
	}
	*dst = append(*dst, buf...)
	return nil
}

type EmptySerialize struct{}

func (d *EmptySerialize) Marshal(args ...any) ([]byte, error) {
	return nil, nil
}

func (d *EmptySerialize) Unmarshal(buf []byte, args ...any) error {
	return nil
}

type ErrorSerialize struct{}

func (d *ErrorSerialize) Marshal(args ...any) ([]byte, error) {
	return nil, uerror.New(-1, "接口未注册")
}

func (d *ErrorSerialize) Unmarshal(buf []byte, args ...any) error {
	return uerror.New(-1, "接口未注册")
}

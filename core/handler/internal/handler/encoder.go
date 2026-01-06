package handler

import (
	"framework/library/crypto"
	"framework/library/uerror"

	"github.com/gogo/protobuf/proto"
)

type GobEncoder struct{}

func (d *GobEncoder) Marshal(args ...any) ([]byte, error) {
	return crypto.GobEncrypto(args...)
}

func (d *GobEncoder) Unmarshal(buf []byte, args ...any) error {
	return crypto.GobDecrypto(buf, args...)
}

type ProtoEncoder struct{}

func (d *ProtoEncoder) Marshal(args ...any) ([]byte, error) {
	return proto.Marshal(args[0].(proto.Message))
}

func (d *ProtoEncoder) Unmarshal(buf []byte, args ...any) error {
	return proto.Unmarshal(buf, args[0].(proto.Message))
}

type EmptyEncoder struct{}

func (d *EmptyEncoder) Marshal(args ...any) ([]byte, error) {
	buf, ok := args[0].([]byte)
	if !ok {
		return nil, uerror.New(-1, "参数类型错误")
	}
	return buf, nil
}

func (d *EmptyEncoder) Unmarshal(buf []byte, args ...any) error {
	dst, ok := args[0].(*[]byte)
	if !ok {
		return uerror.New(-1, "参数类型错误")
	}
	*dst = append(*dst, buf...)
	return nil
}

type ErrorEncoder struct{}

func (d *ErrorEncoder) Marshal(args ...any) ([]byte, error) {
	return nil, uerror.New(-1, "接口未注册")
}

func (d *ErrorEncoder) Unmarshal(buf []byte, args ...any) error {
	return uerror.New(-1, "接口未注册")
}

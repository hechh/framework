package entity

import (
	"framework/library/crypto"
	"framework/library/uerror"
	"hash/crc32"
	"reflect"
	"runtime"
	"strings"

	"github.com/gogo/protobuf/proto"
)

func StringToUint32(str string) uint32 {
	return crc32.ChecksumIEEE([]byte(str))
}

func ParseActorFunc(val reflect.Value) string {
	name := runtime.FuncForPC(val.Pointer()).Name()
	strs := strings.Split(name, "(*")
	return strings.ReplaceAll(strs[len(strs)-1], ")", "")
}

type Common struct {
	nodeType int32
	id       uint32
	cmd      int32
	name     string
}

func NewCommon(nodeType int32, cmd int32, name string) *Common {
	return &Common{
		nodeType: nodeType,
		name:     name,
		cmd:      cmd,
		id:       StringToUint32(name),
	}
}

func (d *Common) GetType() int32 {
	return d.nodeType
}

func (d *Common) GetCmd() int32 {
	return d.cmd
}

func (d *Common) GetId() uint32 {
	return d.id
}

func (d *Common) GetName() string {
	return d.name
}

type GobCrypto struct{}

func (d *GobCrypto) Marshal(args ...any) ([]byte, error) {
	return crypto.Encode(args...)
}

func (d *GobCrypto) Unmarshal(buf []byte, args ...any) error {
	return crypto.Decode(buf, args...)
}

type ProtoCrypto struct{}

func (d *ProtoCrypto) Marshal(args ...any) ([]byte, error) {
	return proto.Marshal(args[0].(proto.Message))
}

func (d *ProtoCrypto) Unmarshal(buf []byte, args ...any) error {
	return proto.Unmarshal(buf, args[0].(proto.Message))
}

type EmptyCrypto struct{}

func (d *EmptyCrypto) Marshal(args ...any) ([]byte, error) {
	buf, ok := args[0].([]byte)
	if !ok {
		return nil, uerror.New(-1, "参数类型错误")
	}
	return buf, nil
}

func (d *EmptyCrypto) Unmarshal(buf []byte, args ...any) error {
	dst, ok := args[0].(*[]byte)
	if !ok {
		return uerror.New(-1, "参数类型错误")
	}
	*dst = append(*dst, buf...)
	return nil
}

type ErrorCrypto struct{}

func (d *ErrorCrypto) Marshal(args ...any) ([]byte, error) {
	return nil, uerror.New(-1, "接口未注册")
}

func (d *ErrorCrypto) Unmarshal(buf []byte, args ...any) error {
	return uerror.New(-1, "接口未注册")
}

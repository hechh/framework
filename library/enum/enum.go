package enum

import (
	"github.com/hechh/framework/library/tplutil"
	"github.com/spf13/cast"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type IEnum interface {
	String() string
	Number() protoreflect.EnumNumber
}

func ToInt32(val any) int32 {
	switch vv := val.(type) {
	case int32:
		return vv
	case IEnum:
		return int32(vv.Number())
	default:
		return cast.ToInt32(val)
	}
}

func ToUint32(val any) uint32 {
	switch vv := val.(type) {
	case IEnum:
		return uint32(vv.Number())
	default:
		return cast.ToUint32(val)
	}
}

type IConvertor interface {
	Has(any) bool
	ToUint32(string) uint32
	ToString(uint32) string
	GetTypes() []int32
}

type Convertor struct {
	names   map[string]int32
	numbers map[int32]string
}

func WrapConvertor(n map[string]int32, i map[int32]string) *Convertor {
	return &Convertor{
		names:   n,
		numbers: i,
	}
}

func (d *Convertor) Has(val any) bool {
	var ok bool
	switch vv := val.(type) {
	case uint32:
		_, ok = d.numbers[int32(vv)]
	case int32:
		_, ok = d.numbers[vv]
	case string:
		_, ok = d.names[vv]
	}
	return ok
}

func (d *Convertor) ToUint32(s string) uint32 {
	return uint32(d.names[s])
}

func (d *Convertor) ToString(i uint32) string {
	return d.numbers[int32(i)]
}

func (d *Convertor) GetTypes() []int32 {
	return tplutil.Map2Keys(d.numbers)
}

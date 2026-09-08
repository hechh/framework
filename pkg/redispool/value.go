package redispool

import (
	"github.com/hechh/framework/define"
	"github.com/hechh/framework/library/tplutil"
	"google.golang.org/protobuf/proto"
)

const (
	HASH   = 1
	STRING = 2
)

type Message interface {
	CloneMessageVT() proto.Message
	MarshalVT() ([]byte, error)
	UnmarshalVT([]byte) error
}

type Value struct {
	Message
	cli   IClient
	class uint32
	key   string
	field string
	times uint32
}

func (d *Value) SetClient(cli IClient) { d.cli = cli }
func (d *Value) GetClient() IClient    { return d.cli }
func (d *Value) GetType() uint32       { return d.class }
func (d *Value) GetKey() string        { return d.key }
func (d *Value) GetField() string      { return d.field }
func (d *Value) Get() any              { return d.Message }
func (d *Value) IsChanged() bool       { return d.times > 0 }
func (d *Value) Change()               { d.times++ }
func (d *Value) Reset()                { d.times = 0 }
func (d *Value) Clone() define.IValue {
	return &Value{
		Message: d.CloneMessageVT().(Message),
		cli:     d.cli,
		class:   d.class,
		key:     d.key,
		field:   d.field,
	}
}

func NewValue(cli IClient, obj Message, t uint32, args ...string) *Value {
	return &Value{
		Message: obj,
		cli:     cli,
		class:   t,
		key:     tplutil.Index(args, 0, ""),
		field:   tplutil.Index(args, 1, ""),
	}
}

func Map2Values[K comparable, V any](vals map[K]V) (rets []*Value) {
	rets = make([]*Value, 0, len(vals))
	for _, v := range vals {
		if vv, ok := any(v).(*Value); ok && vv != nil {
			rets = append(rets, vv)
		}
	}
	return
}

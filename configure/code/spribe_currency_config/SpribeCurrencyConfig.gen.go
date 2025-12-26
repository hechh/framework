/*
* 本代码由cfgtool工具生成，请勿手动修改
 */

package spribe_currency_config

import (
	"framework/configure/pb"
	"framework/library/structure"
	"sync/atomic"

	"github.com/gogo/protobuf/proto"
)

var obj = atomic.Pointer[SpribeCurrencyConfigData]{}

type SpribeCurrencyConfigData struct {
	list     []*pb.SpribeCurrencyConfig
	currency structure.Map1[pb.Currency, *pb.SpribeCurrencyConfig]
	code     structure.Map1[string, *pb.SpribeCurrencyConfig]
}

func DeepCopy(item *pb.SpribeCurrencyConfig) *pb.SpribeCurrencyConfig {
	buf, _ := proto.Marshal(item)
	ret := &pb.SpribeCurrencyConfig{}
	proto.Unmarshal(buf, ret)
	return ret
}

func parse(buf []byte) error {
	ary := &pb.SpribeCurrencyConfigAry{}
	if err := proto.UnmarshalText(string(buf), ary); err != nil {
		return err
	}

	data := &SpribeCurrencyConfigData{
		currency: make(structure.Map1[pb.Currency, *pb.SpribeCurrencyConfig]),
		code:     make(structure.Map1[string, *pb.SpribeCurrencyConfig]),
	}
	for _, item := range ary.Ary {
		data.list = append(data.list, item)
		data.currency.Put(item.Currency, item)
		data.code.Put(item.Code, item)
	}
	obj.Store(data)
	return nil
}

func SGet(pos int) *pb.SpribeCurrencyConfig {
	if pos < 0 {
		pos = 0
	}
	list := obj.Load().list
	if ll := len(list); ll-1 < pos {
		pos = ll - 1
	}
	return list[pos]
}

func LGet() (rets []*pb.SpribeCurrencyConfig) {
	list := obj.Load().list
	rets = make([]*pb.SpribeCurrencyConfig, len(list))
	copy(rets, list)
	return
}

func Walk(f func(*pb.SpribeCurrencyConfig) bool) {
	for _, item := range obj.Load().list {
		if !f(item) {
			return
		}
	}
}

func MGetCurrency(Currency pb.Currency) *pb.SpribeCurrencyConfig {
	data := obj.Load().currency
	if value, ok := data.Get(Currency); ok {
		return value
	}
	return nil
}

func MGetCode(Code string) *pb.SpribeCurrencyConfig {
	data := obj.Load().code
	if value, ok := data.Get(Code); ok {
		return value
	}
	return nil
}

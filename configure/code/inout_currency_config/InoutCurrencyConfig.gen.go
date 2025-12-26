/*
* 本代码由cfgtool工具生成，请勿手动修改
 */

package inout_currency_config

import (
	"framework/configure/pb"
	"framework/library/structure"
	"sync/atomic"

	"github.com/gogo/protobuf/proto"
)

var obj = atomic.Pointer[InoutCurrencyConfigData]{}

type InoutCurrencyConfigData struct {
	list     []*pb.InoutCurrencyConfig
	currency structure.Map1[pb.Currency, *pb.InoutCurrencyConfig]
	code     structure.Map1[string, *pb.InoutCurrencyConfig]
}

func DeepCopy(item *pb.InoutCurrencyConfig) *pb.InoutCurrencyConfig {
	buf, _ := proto.Marshal(item)
	ret := &pb.InoutCurrencyConfig{}
	proto.Unmarshal(buf, ret)
	return ret
}

func parse(buf []byte) error {
	ary := &pb.InoutCurrencyConfigAry{}
	if err := proto.UnmarshalText(string(buf), ary); err != nil {
		return err
	}

	data := &InoutCurrencyConfigData{
		currency: make(structure.Map1[pb.Currency, *pb.InoutCurrencyConfig]),
		code:     make(structure.Map1[string, *pb.InoutCurrencyConfig]),
	}
	for _, item := range ary.Ary {
		data.list = append(data.list, item)
		data.currency.Put(item.Currency, item)
		data.code.Put(item.Code, item)
	}
	obj.Store(data)
	return nil
}

func SGet(pos int) *pb.InoutCurrencyConfig {
	if pos < 0 {
		pos = 0
	}
	list := obj.Load().list
	if ll := len(list); ll-1 < pos {
		pos = ll - 1
	}
	return list[pos]
}

func LGet() (rets []*pb.InoutCurrencyConfig) {
	list := obj.Load().list
	rets = make([]*pb.InoutCurrencyConfig, len(list))
	copy(rets, list)
	return
}

func Walk(f func(*pb.InoutCurrencyConfig) bool) {
	for _, item := range obj.Load().list {
		if !f(item) {
			return
		}
	}
}

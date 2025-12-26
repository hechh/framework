/*
* 本代码由cfgtool工具生成，请勿手动修改
 */

package currency_exchange_config

import (
	"framework/configure/pb"
	"framework/library/structure"
	"sync/atomic"

	"github.com/gogo/protobuf/proto"
)

var obj = atomic.Pointer[CurrencyExchangeConfigData]{}

type CurrencyExchangeConfigData struct {
	list     []*pb.CurrencyExchangeConfig
	currency structure.Map1[pb.Currency, *pb.CurrencyExchangeConfig]
}

func DeepCopy(item *pb.CurrencyExchangeConfig) *pb.CurrencyExchangeConfig {
	buf, _ := proto.Marshal(item)
	ret := &pb.CurrencyExchangeConfig{}
	proto.Unmarshal(buf, ret)
	return ret
}

func parse(buf []byte) error {
	ary := &pb.CurrencyExchangeConfigAry{}
	if err := proto.UnmarshalText(string(buf), ary); err != nil {
		return err
	}

	data := &CurrencyExchangeConfigData{
		currency: make(structure.Map1[pb.Currency, *pb.CurrencyExchangeConfig]),
	}
	for _, item := range ary.Ary {
		data.list = append(data.list, item)
		data.currency.Put(item.Currency, item)
	}
	obj.Store(data)
	return nil
}

func SGet(pos int) *pb.CurrencyExchangeConfig {
	if pos < 0 {
		pos = 0
	}
	list := obj.Load().list
	if ll := len(list); ll-1 < pos {
		pos = ll - 1
	}
	return list[pos]
}

func LGet() (rets []*pb.CurrencyExchangeConfig) {
	list := obj.Load().list
	rets = make([]*pb.CurrencyExchangeConfig, len(list))
	copy(rets, list)
	return
}

func Walk(f func(*pb.CurrencyExchangeConfig) bool) {
	for _, item := range obj.Load().list {
		if !f(item) {
			return
		}
	}
}

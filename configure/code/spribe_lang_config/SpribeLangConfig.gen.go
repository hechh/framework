/*
* 本代码由cfgtool工具生成，请勿手动修改
 */

package spribe_lang_config

import (
	"framework/configure/pb"
	"framework/library/structure"
	"sync/atomic"

	"github.com/gogo/protobuf/proto"
)

var obj = atomic.Pointer[SpribeLangConfigData]{}

type SpribeLangConfigData struct {
	list     []*pb.SpribeLangConfig
	langType structure.Map1[pb.LangType, *pb.SpribeLangConfig]
	code     structure.Map1[string, *pb.SpribeLangConfig]
}

func DeepCopy(item *pb.SpribeLangConfig) *pb.SpribeLangConfig {
	buf, _ := proto.Marshal(item)
	ret := &pb.SpribeLangConfig{}
	proto.Unmarshal(buf, ret)
	return ret
}

func parse(buf []byte) error {
	ary := &pb.SpribeLangConfigAry{}
	if err := proto.UnmarshalText(string(buf), ary); err != nil {
		return err
	}

	data := &SpribeLangConfigData{
		langType: make(structure.Map1[pb.LangType, *pb.SpribeLangConfig]),
		code:     make(structure.Map1[string, *pb.SpribeLangConfig]),
	}
	for _, item := range ary.Ary {
		data.list = append(data.list, item)
		data.langType.Put(item.LangType, item)
		data.code.Put(item.Code, item)
	}
	obj.Store(data)
	return nil
}

func SGet(pos int) *pb.SpribeLangConfig {
	if pos < 0 {
		pos = 0
	}
	list := obj.Load().list
	if ll := len(list); ll-1 < pos {
		pos = ll - 1
	}
	return list[pos]
}

func LGet() (rets []*pb.SpribeLangConfig) {
	list := obj.Load().list
	rets = make([]*pb.SpribeLangConfig, len(list))
	copy(rets, list)
	return
}

func Walk(f func(*pb.SpribeLangConfig) bool) {
	for _, item := range obj.Load().list {
		if !f(item) {
			return
		}
	}
}

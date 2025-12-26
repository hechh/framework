/*
* 本代码由cfgtool工具生成，请勿手动修改
 */

package inout_lang_config

import (
	"framework/configure/pb"
	"framework/library/structure"
	"sync/atomic"

	"github.com/gogo/protobuf/proto"
)

var obj = atomic.Pointer[InoutLangConfigData]{}

type InoutLangConfigData struct {
	list     []*pb.InoutLangConfig
	langType structure.Map1[pb.LangType, *pb.InoutLangConfig]
	code     structure.Map1[string, *pb.InoutLangConfig]
}

func DeepCopy(item *pb.InoutLangConfig) *pb.InoutLangConfig {
	buf, _ := proto.Marshal(item)
	ret := &pb.InoutLangConfig{}
	proto.Unmarshal(buf, ret)
	return ret
}

func parse(buf []byte) error {
	ary := &pb.InoutLangConfigAry{}
	if err := proto.UnmarshalText(string(buf), ary); err != nil {
		return err
	}

	data := &InoutLangConfigData{
		langType: make(structure.Map1[pb.LangType, *pb.InoutLangConfig]),
		code:     make(structure.Map1[string, *pb.InoutLangConfig]),
	}
	for _, item := range ary.Ary {
		data.list = append(data.list, item)
		data.langType.Put(item.LangType, item)
		data.code.Put(item.Code, item)
	}
	obj.Store(data)
	return nil
}

func SGet(pos int) *pb.InoutLangConfig {
	if pos < 0 {
		pos = 0
	}
	list := obj.Load().list
	if ll := len(list); ll-1 < pos {
		pos = ll - 1
	}
	return list[pos]
}

func LGet() (rets []*pb.InoutLangConfig) {
	list := obj.Load().list
	rets = make([]*pb.InoutLangConfig, len(list))
	copy(rets, list)
	return
}

func Walk(f func(*pb.InoutLangConfig) bool) {
	for _, item := range obj.Load().list {
		if !f(item) {
			return
		}
	}
}

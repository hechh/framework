/*
* 本代码由cfgtool工具生成，请勿手动修改
 */

package global_config

import (
	"framework/configure/pb"
	"sync/atomic"

	"github.com/gogo/protobuf/proto"
)

var obj = atomic.Pointer[GlobalConfigData]{}

type GlobalConfigData struct {
	list []*pb.GlobalConfig
}

func DeepCopy(item *pb.GlobalConfig) *pb.GlobalConfig {
	buf, _ := proto.Marshal(item)
	ret := &pb.GlobalConfig{}
	proto.Unmarshal(buf, ret)
	return ret
}

func parse(buf []byte) error {
	ary := &pb.GlobalConfigAry{}
	if err := proto.UnmarshalText(string(buf), ary); err != nil {
		return err
	}

	data := &GlobalConfigData{}
	for _, item := range ary.Ary {
		data.list = append(data.list, item)
	}
	obj.Store(data)
	return nil
}

func SGet(pos int) *pb.GlobalConfig {
	if pos < 0 {
		pos = 0
	}
	list := obj.Load().list
	if ll := len(list); ll-1 < pos {
		pos = ll - 1
	}
	return list[pos]
}

func LGet() (rets []*pb.GlobalConfig) {
	list := obj.Load().list
	rets = make([]*pb.GlobalConfig, len(list))
	copy(rets, list)
	return
}

func Walk(f func(*pb.GlobalConfig) bool) {
	for _, item := range obj.Load().list {
		if !f(item) {
			return
		}
	}
}

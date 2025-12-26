/*
* 本代码由cfgtool工具生成，请勿手动修改
 */

package spribe_platform_config

import (
	"framework/configure/pb"
	"framework/library/structure"
	"sync/atomic"

	"github.com/gogo/protobuf/proto"
)

var obj = atomic.Pointer[SpribePlatformConfigData]{}

type SpribePlatformConfigData struct {
	list     []*pb.SpribePlatformConfig
	platform structure.Map1[pb.Platform, *pb.SpribePlatformConfig]
	code     structure.Map1[string, *pb.SpribePlatformConfig]
}

func DeepCopy(item *pb.SpribePlatformConfig) *pb.SpribePlatformConfig {
	buf, _ := proto.Marshal(item)
	ret := &pb.SpribePlatformConfig{}
	proto.Unmarshal(buf, ret)
	return ret
}

func parse(buf []byte) error {
	ary := &pb.SpribePlatformConfigAry{}
	if err := proto.UnmarshalText(string(buf), ary); err != nil {
		return err
	}

	data := &SpribePlatformConfigData{
		platform: make(structure.Map1[pb.Platform, *pb.SpribePlatformConfig]),
		code:     make(structure.Map1[string, *pb.SpribePlatformConfig]),
	}
	for _, item := range ary.Ary {
		data.list = append(data.list, item)
		data.platform.Put(item.Platform, item)
		data.code.Put(item.Code, item)
	}
	obj.Store(data)
	return nil
}

func SGet(pos int) *pb.SpribePlatformConfig {
	if pos < 0 {
		pos = 0
	}
	list := obj.Load().list
	if ll := len(list); ll-1 < pos {
		pos = ll - 1
	}
	return list[pos]
}

func LGet() (rets []*pb.SpribePlatformConfig) {
	list := obj.Load().list
	rets = make([]*pb.SpribePlatformConfig, len(list))
	copy(rets, list)
	return
}

func Walk(f func(*pb.SpribePlatformConfig) bool) {
	for _, item := range obj.Load().list {
		if !f(item) {
			return
		}
	}
}

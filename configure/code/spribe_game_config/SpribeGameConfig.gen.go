/*
* 本代码由cfgtool工具生成，请勿手动修改
 */

package spribe_game_config

import (
	"framework/configure/pb"
	"framework/library/structure"
	"sync/atomic"

	"github.com/gogo/protobuf/proto"
)

var obj = atomic.Pointer[SpribeGameConfigData]{}

type SpribeGameConfigData struct {
	list     []*pb.SpribeGameConfig
	gameType structure.Map1[pb.GameType, *pb.SpribeGameConfig]
	code     structure.Map1[string, *pb.SpribeGameConfig]
}

func DeepCopy(item *pb.SpribeGameConfig) *pb.SpribeGameConfig {
	buf, _ := proto.Marshal(item)
	ret := &pb.SpribeGameConfig{}
	proto.Unmarshal(buf, ret)
	return ret
}

func parse(buf []byte) error {
	ary := &pb.SpribeGameConfigAry{}
	if err := proto.UnmarshalText(string(buf), ary); err != nil {
		return err
	}

	data := &SpribeGameConfigData{
		gameType: make(structure.Map1[pb.GameType, *pb.SpribeGameConfig]),
		code:     make(structure.Map1[string, *pb.SpribeGameConfig]),
	}
	for _, item := range ary.Ary {
		data.list = append(data.list, item)
		data.gameType.Put(item.GameType, item)
		data.code.Put(item.Code, item)
	}
	obj.Store(data)
	return nil
}

func SGet(pos int) *pb.SpribeGameConfig {
	if pos < 0 {
		pos = 0
	}
	list := obj.Load().list
	if ll := len(list); ll-1 < pos {
		pos = ll - 1
	}
	return list[pos]
}

func LGet() (rets []*pb.SpribeGameConfig) {
	list := obj.Load().list
	rets = make([]*pb.SpribeGameConfig, len(list))
	copy(rets, list)
	return
}

func Walk(f func(*pb.SpribeGameConfig) bool) {
	for _, item := range obj.Load().list {
		if !f(item) {
			return
		}
	}
}

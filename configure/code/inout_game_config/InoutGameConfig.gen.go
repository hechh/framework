/*
* 本代码由cfgtool工具生成，请勿手动修改
 */

package inout_game_config

import (
	"framework/configure/pb"
	"framework/library/structure"
	"sync/atomic"

	"github.com/gogo/protobuf/proto"
)

var obj = atomic.Pointer[InoutGameConfigData]{}

type InoutGameConfigData struct {
	list     []*pb.InoutGameConfig
	gameType structure.Map1[pb.GameType, *pb.InoutGameConfig]
	code     structure.Map1[string, *pb.InoutGameConfig]
}

func DeepCopy(item *pb.InoutGameConfig) *pb.InoutGameConfig {
	buf, _ := proto.Marshal(item)
	ret := &pb.InoutGameConfig{}
	proto.Unmarshal(buf, ret)
	return ret
}

func parse(buf []byte) error {
	ary := &pb.InoutGameConfigAry{}
	if err := proto.UnmarshalText(string(buf), ary); err != nil {
		return err
	}

	data := &InoutGameConfigData{
		gameType: make(structure.Map1[pb.GameType, *pb.InoutGameConfig]),
		code:     make(structure.Map1[string, *pb.InoutGameConfig]),
	}
	for _, item := range ary.Ary {
		data.list = append(data.list, item)
		data.gameType.Put(item.GameType, item)
		data.code.Put(item.Code, item)
	}
	obj.Store(data)
	return nil
}

func SGet(pos int) *pb.InoutGameConfig {
	if pos < 0 {
		pos = 0
	}
	list := obj.Load().list
	if ll := len(list); ll-1 < pos {
		pos = ll - 1
	}
	return list[pos]
}

func LGet() (rets []*pb.InoutGameConfig) {
	list := obj.Load().list
	rets = make([]*pb.InoutGameConfig, len(list))
	copy(rets, list)
	return
}

func Walk(f func(*pb.InoutGameConfig) bool) {
	for _, item := range obj.Load().list {
		if !f(item) {
			return
		}
	}
}

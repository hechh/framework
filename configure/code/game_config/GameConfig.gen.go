/*
* 本代码由cfgtool工具生成，请勿手动修改
 */

package game_config

import (
	"framework/configure/pb"
	"framework/library/structure"
	"sync/atomic"

	"github.com/gogo/protobuf/proto"
)

var obj = atomic.Pointer[GameConfigData]{}

type GameConfigData struct {
	list     []*pb.GameConfig
	provider structure.Group1[pb.Provider, *pb.GameConfig]
	gameType structure.Group1[pb.GameType, *pb.GameConfig]
}

func DeepCopy(item *pb.GameConfig) *pb.GameConfig {
	buf, _ := proto.Marshal(item)
	ret := &pb.GameConfig{}
	proto.Unmarshal(buf, ret)
	return ret
}

func parse(buf []byte) error {
	ary := &pb.GameConfigAry{}
	if err := proto.UnmarshalText(string(buf), ary); err != nil {
		return err
	}

	data := &GameConfigData{
		provider: make(structure.Group1[pb.Provider, *pb.GameConfig]),
		gameType: make(structure.Group1[pb.GameType, *pb.GameConfig]),
	}
	for _, item := range ary.Ary {
		data.list = append(data.list, item)
		data.provider.Put(item.Provider, item)
		data.gameType.Put(item.GameType, item)
	}
	obj.Store(data)
	return nil
}

func SGet(pos int) *pb.GameConfig {
	if pos < 0 {
		pos = 0
	}
	list := obj.Load().list
	if ll := len(list); ll-1 < pos {
		pos = ll - 1
	}
	return list[pos]
}

func LGet() (rets []*pb.GameConfig) {
	list := obj.Load().list
	rets = make([]*pb.GameConfig, len(list))
	copy(rets, list)
	return
}

func Walk(f func(*pb.GameConfig) bool) {
	for _, item := range obj.Load().list {
		if !f(item) {
			return
		}
	}
}

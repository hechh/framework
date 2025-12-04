package router

import (
	"framework/define"
	"framework/packet"
)

type GetFunc func(int32, uint64) define.IRouter // 获取路由表

var (
	loadFunc      GetFunc
	loadOrNewFunc GetFunc
)

func Set(l, lon GetFunc) {
	loadFunc = l
	loadOrNewFunc = lon
}

func Load(idType int32, id uint64) define.IRouter {
	return loadFunc(idType, id)
}

func LoadOrNew(idType int32, id uint64) define.IRouter {
	return loadOrNewFunc(idType, id)
}

func GetRouter(idType int32, id uint64) *packet.Router {
	return &packet.Router{
		IdType: idType,
		Id:     id,
		List:   loadOrNewFunc(idType, id).GetRouter(),
	}
}

func SetRouter(rs ...*packet.Router) {
	for _, data := range rs {
		if data != nil {
			loadOrNewFunc(data.IdType, data.Id).SetRouter(data.List...)
		}
	}
}

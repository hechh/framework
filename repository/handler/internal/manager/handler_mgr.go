package manager

import (
	"framework/define"
	"framework/internal/global"
	"framework/repository/handler/internal/entity"
)

var (
	mapName  = make(map[string]uint32)
	mapValue = make(map[uint32]string)
	mapCmd   = make(map[int32]define.IHandler)
	mapRpc   = make(map[int32]map[uint32]define.IHandler)
	mapFunc  = make(map[string]define.IHandler)
)

func Name2Id(str string) uint32 {
	if val, ok := mapName[str]; ok {
		return val
	}
	return 0
}

func Id2Name(val uint32) string {
	if str, ok := mapValue[val]; ok {
		return str
	}
	return ""
}

func RegisgerName(str string) {
	val := entity.StringToUint32(str)
	mapValue[val] = str
	mapName[str] = val
}

// 注册全局Rpc
func RegisterRpc(hh define.IHandler) {
	nodeType := hh.GetType()
	if _, ok := mapRpc[nodeType]; !ok {
		mapRpc[nodeType] = make(map[uint32]define.IHandler)
	}
	mapRpc[nodeType][hh.GetId()] = hh

	// 命令字注册
	if cmd := hh.GetCmd(); cmd > 0 {
		mapCmd[hh.GetCmd()] = hh
	}

	// 设置name 映射 id
	RegisgerName(hh.GetName())
}

// 获取全局Rpc
func GetRpc(nodeType int32, id uint32) define.IHandler {
	if nodeType == 0 {
		nodeType = global.GetSelfType()
	}
	if items, ok := mapRpc[nodeType]; ok {
		return items[id]
	}
	return nil
}

// 通过命令字获取全局Rpc
func GetCmd(cmd int32) define.IHandler {
	if hh, ok := mapCmd[cmd]; ok {
		return hh
	}
	return nil
}

// 当前服务节点的注册(特殊)
func Register(h define.IHandler) {
	mapFunc[h.GetName()] = h

	// 设置name 映射 id
	RegisgerName(h.GetName())
}

func Get(name string) define.IHandler {
	if f, ok := mapFunc[name]; ok {
		return f
	}
	return nil
}

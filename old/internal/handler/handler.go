package handler

import "framework/define"

type HandlerService struct {
	names  map[string]uint32                    // name to value
	values map[uint32]string                    // value to name
	locals map[string]define.IHandler           // 本地接口
	rpcs   map[int32]map[uint32]define.IHandler // 远程接口
	cmds   map[uint32]define.IHandler           // 命令字
}

func NewHandlerService() *HandlerService {
	return &HandlerService{
		names:  make(map[string]uint32),
		values: make(map[uint32]string),
		locals: make(map[string]define.IHandler),
		rpcs:   make(map[int32]map[uint32]define.IHandler),
		cmds:   make(map[uint32]define.IHandler),
	}
}

func (d *HandlerService) Name2Id(name string) (uint32, bool) {
	val, ok := d.names[name]
	return val, ok
}

func (d *HandlerService) Id2Name(val uint32) (string, bool) {
	name, ok := d.values[val]
	return name, ok
}

func (d *HandlerService) Register(hh define.IHandler) {
	d.names[hh.GetName()] = hh.GetId()
	d.values[hh.GetId()] = hh.GetName()
	d.locals[hh.GetName()] = hh
}

func (d *HandlerService) RegisterRpc(hh define.IHandler) {
	d.names[hh.GetName()] = hh.GetId()
	d.values[hh.GetId()] = hh.GetName()
	if hh.GetCmd() > 0 {
		d.cmds[hh.GetCmd()] = hh
	}
	if _, ok := d.rpcs[hh.GetType()]; !ok {
		d.rpcs[hh.GetType()] = make(map[uint32]define.IHandler)
	}
	d.rpcs[hh.GetType()][hh.GetId()] = hh
}

func (d *HandlerService) Get(key any, values ...any) define.IHandler {
	switch v1 := key.(type) {
	case string:
		if val, ok := d.locals[v1]; ok {
			return val
		}
	case uint32:
		if val, ok := d.cmds[v1]; ok {
			return val
		}
	case int32:
		if vals, ok := d.rpcs[v1]; ok {
			switch v2 := values[0].(type) {
			case string:
				if val, ok := vals[d.names[v2]]; ok {
					return val
				}
			case uint32:
				if val, ok := vals[v2]; ok {
					return val
				}
			}
		}
	}
	return nil
}

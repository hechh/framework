package service

import "framework/core"

type Service struct {
	names  map[string]uint32                   // name to value
	values map[uint32]string                   // value to name
	locals map[string]core.IHandler            // 本地接口
	rpcs   map[uint32]map[uint32]core.IHandler // 远程接口
	cmds   map[uint32]core.IHandler            // 命令字
}

func NewService() *Service {
	return &Service{
		names:  make(map[string]uint32),
		values: make(map[uint32]string),
		locals: make(map[string]core.IHandler),
		rpcs:   make(map[uint32]map[uint32]core.IHandler),
		cmds:   make(map[uint32]core.IHandler),
	}
}

func (d *Service) Name2Id(name string) (uint32, bool) {
	val, ok := d.names[name]
	return val, ok
}

func (d *Service) Id2Name(val uint32) (string, bool) {
	name, ok := d.values[val]
	return name, ok
}

func (d *Service) Register(hh core.IHandler) {
	d.names[hh.GetName()] = hh.GetId()
	d.values[hh.GetId()] = hh.GetName()
	d.locals[hh.GetName()] = hh
}

func (d *Service) RegisterRpc(hh core.IHandler) {
	d.names[hh.GetName()] = hh.GetId()
	d.values[hh.GetId()] = hh.GetName()
	if hh.GetCmd() > 0 {
		d.cmds[hh.GetCmd()] = hh
	}
	if _, ok := d.rpcs[hh.GetType()]; !ok {
		d.rpcs[hh.GetType()] = make(map[uint32]core.IHandler)
	}
	d.rpcs[hh.GetType()][hh.GetId()] = hh
}

func (d *Service) Get(actorFunc string) core.IHandler {
	if val, ok := d.locals[actorFunc]; ok {
		return val
	}
	return nil
}

func (d *Service) GetByCmd(cmd uint32) core.IHandler {
	if val, ok := d.cmds[cmd]; ok {
		return val
	}
	return nil
}

func (d *Service) GetByRpc(nodeType uint32, id any) core.IHandler {
	if vals, ok := d.rpcs[nodeType]; ok {
		switch vv := id.(type) {
		case uint32:
			if val, ok := vals[vv]; ok {
				return val
			}
		case string:
			vid, _ := d.names[vv]
			if val, ok := vals[vid]; ok {
				return val
			}
		}
	}
	return nil
}

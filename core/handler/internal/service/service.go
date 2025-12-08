package service

import "framework/core/handler/domain"

type Service struct {
	names  map[string]uint32                     // name to value
	values map[uint32]string                     // value to name
	locals map[string]domain.IHandler            // 本地接口
	rpcs   map[uint32]map[uint32]domain.IHandler // 远程接口
	cmds   map[uint32]domain.IHandler            // 命令字
}

func NewService() *Service {
	return &Service{
		names:  make(map[string]uint32),
		values: make(map[uint32]string),
		locals: make(map[string]domain.IHandler),
		rpcs:   make(map[uint32]map[uint32]domain.IHandler),
		cmds:   make(map[uint32]domain.IHandler),
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

func (d *Service) Register(hh domain.IHandler) {
	d.names[hh.GetName()] = hh.GetId()
	d.values[hh.GetId()] = hh.GetName()
	d.locals[hh.GetName()] = hh
}

func (d *Service) RegisterRpc(hh domain.IHandler) {
	d.names[hh.GetName()] = hh.GetId()
	d.values[hh.GetId()] = hh.GetName()
	if hh.GetCmd() > 0 {
		d.cmds[hh.GetCmd()] = hh
	}
	if _, ok := d.rpcs[hh.GetType()]; !ok {
		d.rpcs[hh.GetType()] = make(map[uint32]domain.IHandler)
	}
	d.rpcs[hh.GetType()][hh.GetId()] = hh
}

func (d *Service) Get(actorFunc string) domain.IHandler {
	if val, ok := d.locals[actorFunc]; ok {
		return val
	}
	return nil
}

func (d *Service) GetByCmd(cmd uint32) domain.IHandler {
	if val, ok := d.cmds[cmd]; ok {
		return val
	}
	return nil
}

func (d *Service) GetByRpc(nodeType uint32, id uint32) domain.IHandler {
	if vals, ok := d.rpcs[nodeType]; ok {
		if val, ok := vals[id]; ok {
			return val
		}
	}
	return nil
}

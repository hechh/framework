package service

import (
	"github.com/hechh/framework"
	"github.com/hechh/library/util"
)

type Service struct {
	names    map[string]uint32
	values   map[uint32]string
	handlers map[string]framework.IHandler
	cmds     map[uint32]framework.IRpc
	rpcs     util.Map2[uint32, uint32, framework.IRpc]
}

func NewService() *Service {
	return &Service{
		names:    make(map[string]uint32),
		values:   make(map[uint32]string),
		handlers: make(map[string]framework.IHandler),
		cmds:     make(map[uint32]framework.IRpc),
		rpcs:     make(util.Map2[uint32, uint32, framework.IRpc]),
	}
}

func (s *Service) Name2Id(name string) (uint32, bool) {
	val, ok := s.names[name]
	return val, ok
}

func (s *Service) Id2Name(val uint32) (string, bool) {
	name, ok := s.values[val]
	return name, ok
}

func (s *Service) Register(hh framework.IHandler) {
	s.names[hh.GetName()] = hh.GetCrc32()
	s.values[hh.GetCrc32()] = hh.GetName()
	s.handlers[hh.GetName()] = hh
}

func (s *Service) Get(actorFunc string) framework.IHandler {
	if val, ok := s.handlers[actorFunc]; ok {
		return val
	}
	return nil
}

func (s *Service) RegisterRpc(hh framework.IRpc) {
	s.names[hh.GetName()] = hh.GetCrc32()
	s.values[hh.GetCrc32()] = hh.GetName()
	if cmd := hh.GetCmd(); cmd > 0 {
		s.cmds[cmd] = hh
	}
	s.rpcs.Put(hh.GetNodeType(), hh.GetCrc32(), hh)
}

func (s *Service) GetCmdRpc(cmd uint32) framework.IRpc {
	if val, ok := s.cmds[cmd]; ok {
		return val
	}
	return nil
}

func (s *Service) GetRpc(nodeType uint32, id any) framework.IRpc {
	switch vv := id.(type) {
	case uint32:
		if val, ok := s.rpcs.Get(nodeType, vv); ok {
			return val
		}
	case string:
		vid, _ := s.names[vv]
		if val, ok := s.rpcs.Get(nodeType, vid); ok {
			return val
		}
	}
	return nil
}

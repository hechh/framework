package rpc

import (
	"fmt"

	"github.com/hechh/framework/kernel/define"
	"github.com/hechh/framework/library/utils"
)

type R0 struct {
	nodeType      uint32
	cmd           uint32
	actorFuncName string
	actorFunc     uint32
	flag          uint32
}

// NewR0 创建R0实例
func NewR0(nodeType uint32, cmd uint32, actorFuncName string, flag uint32) *R0 {
	return &R0{
		nodeType:      nodeType,
		cmd:           cmd,
		actorFuncName: actorFuncName,
		actorFunc:     utils.GetCrc32(actorFuncName),
		flag:          flag,
	}
}

func (d *R0) GetActorFuncName() string { return d.actorFuncName }
func (d *R0) GetActorFunc() uint32     { return d.actorFunc }
func (d *R0) GetNodeType() uint32      { return d.nodeType }
func (d *R0) GetCmd() uint32           { return d.cmd }
func (d *R0) GetMask() uint32          { return d.flag }

func (d *R0) News([]byte) ([]any, error) {
	return nil, nil
}

type R1[R any] struct {
	nodeType      uint32
	cmd           uint32
	actorFuncName string
	actorFunc     uint32
	flag          uint32
}

// NewR1 创建R1实例
func NewR1[R any](nodeType uint32, cmd uint32, actorFuncName string, flag uint32) *R1[R] {
	return &R1[R]{
		actorFuncName: actorFuncName,
		actorFunc:     utils.GetCrc32(actorFuncName),
		nodeType:      nodeType,
		cmd:           cmd,
		flag:          flag,
	}
}

func (d *R1[R]) GetActorFuncName() string { return d.actorFuncName }
func (d *R1[R]) GetActorFunc() uint32     { return d.actorFunc }
func (d *R1[R]) GetNodeType() uint32      { return d.nodeType }
func (d *R1[R]) GetCmd() uint32           { return d.cmd }
func (d *R1[R]) GetMask() uint32          { return d.flag }

func (d *R1[R]) News(body []byte) ([]any, error) {
	val := new(R)
	var err error
	if len(body) > 0 {
		if obj, ok := any(val).(define.Message); ok {
			err = obj.UnmarshalVT(body)
		} else {
			err = fmt.Errorf("Rpc接口交互协议只能使用protobuf")
		}
	}
	return []any{val}, err
}

type R2[R any, T any] struct {
	nodeType      uint32
	cmd           uint32
	actorFuncName string
	actorFunc     uint32
	flag          uint32
}

// NewR2 创建R2实例
func NewR2[R any, T any](nodeType uint32, cmd uint32, actorFuncName string, flag uint32) *R2[R, T] {
	return &R2[R, T]{
		nodeType:      nodeType,
		cmd:           cmd,
		actorFuncName: actorFuncName,
		actorFunc:     utils.GetCrc32(actorFuncName),
		flag:          flag,
	}
}

func (d *R2[R, T]) GetActorFuncName() string { return d.actorFuncName }
func (d *R2[R, T]) GetActorFunc() uint32     { return d.actorFunc }
func (d *R2[R, T]) GetNodeType() uint32      { return d.nodeType }
func (d *R2[R, T]) GetCmd() uint32           { return d.cmd }
func (d *R2[R, T]) GetMask() uint32          { return d.flag }

func (d *R2[R, T]) News(body []byte) ([]any, error) {
	r, t := new(R), new(T)
	dst := make([]any, 0, 2)
	dst = append(dst, r, t)
	var err error
	if len(body) > 0 {
		if obj, ok := any(r).(define.Message); ok {
			err = obj.UnmarshalVT(body)
		} else {
			err = fmt.Errorf("Rpc接口交互协议只能使用protobuf")
		}
	}
	return dst, err
}

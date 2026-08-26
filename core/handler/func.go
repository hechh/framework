package handler

import (
	"reflect"
	"runtime"
	"strings"

	"github.com/hechh/framework/core/define"
	"github.com/hechh/framework/library/utils"
)

type (
	M0Func[A any]               func(A, define.IContext) error       // 处理函数0
	M1Func[A any, R any]        func(A, define.IContext, R) error    // 处理函数1
	M2Func[A any, R any, U any] func(A, define.IContext, R, U) error // 处理函数2
)

type M0[A any] struct {
	actorFuncName string
	actorFunc     uint32
	f             M0Func[A]
	flag          uint32
}

func NewM0[A any](f M0Func[A], flag uint32) *M0[A] {
	actorFuncName := parseActorFunc(reflect.ValueOf(f))
	return &M0[A]{
		actorFuncName: actorFuncName,
		actorFunc:     utils.GetCrc32(actorFuncName),
		f:             f,
		flag:          flag,
	}
}

func (d *M0[A]) GetActorFuncName() string { return d.actorFuncName }
func (d *M0[A]) GetActorFunc() uint32     { return d.actorFunc }
func (d *M0[A]) GetMask() uint32          { return d.flag }
func (d *M0[A]) Call(a any, ctx define.IContext, args ...any) error {
	return d.f(a.(A), ctx)
}

type M1[A any, R any] struct {
	actorFuncName string
	actorFunc     uint32
	f             M1Func[A, R]
	flag          uint32
}

func NewM1[A any, R any](f M1Func[A, R], flag uint32) *M1[A, R] {
	actorFuncName := parseActorFunc(reflect.ValueOf(f))
	return &M1[A, R]{
		actorFuncName: actorFuncName,
		actorFunc:     utils.GetCrc32(actorFuncName),
		f:             f,
		flag:          flag,
	}
}

func (d *M1[A, R]) GetActorFuncName() string { return d.actorFuncName }
func (d *M1[A, R]) GetActorFunc() uint32     { return d.actorFunc }
func (d *M1[A, R]) GetMask() uint32          { return d.flag }
func (d *M1[A, R]) Call(a any, ctx define.IContext, args ...any) error {
	return d.f(a.(A), ctx, args[0].(R))
}

type M2[A any, R any, U any] struct {
	actorFuncName string
	actorFunc     uint32
	f             M2Func[A, R, U]
	flag          uint32
}

func NewM2[A any, R any, U any](f M2Func[A, R, U], flag uint32) *M2[A, R, U] {
	actorFuncName := parseActorFunc(reflect.ValueOf(f))
	return &M2[A, R, U]{
		actorFuncName: actorFuncName,
		actorFunc:     utils.GetCrc32(actorFuncName),
		f:             f,
		flag:          flag,
	}
}

func (d *M2[A, R, U]) GetActorFuncName() string { return d.actorFuncName }
func (d *M2[A, R, U]) GetActorFunc() uint32     { return d.actorFunc }
func (d *M2[A, R, U]) GetMask() uint32          { return d.flag }
func (d *M2[A, R, U]) Call(a any, ctx define.IContext, args ...any) error {
	return d.f(a.(A), ctx, args[0].(R), args[1].(U))
}

func parseActorFunc(fun reflect.Value) string {
	runName := runtime.FuncForPC(fun.Pointer()).Name()
	strs := strings.Split(runName, "(*")
	return strings.ReplaceAll(strs[len(strs)-1], ")", "")
}

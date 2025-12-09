package domain

import (
	"framework/define"
	"framework/packet"
)

// 回复接口
type RspFunc func(*packet.Head, ...any) error

// handler范式
type V0Func[Actor any] func(*Actor, define.IContext) error
type V1Func[Actor any, V1 any] func(*Actor, define.IContext, V1) error
type V2Func[Actor any, V1 any, V2 any] func(*Actor, define.IContext, V1, V2) error
type V3Func[Actor any, V1 any, V2 any, V3 any] func(*Actor, define.IContext, V1, V2, V3) error

type P1Func[Actor any, V1 any] func(*Actor, define.IContext, *V1) error
type P2Func[Actor any, V1 any, V2 any] func(*Actor, define.IContext, *V1, *V2) error
type P3Func[Actor any, V1 any, V2 any, V3 any] func(*Actor, define.IContext, *V1, *V2, *V3) error

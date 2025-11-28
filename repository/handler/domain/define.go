package domain

import "framework/define"

// 本地
type L0Func[Actor any] func(*Actor) error
type L1Func[Actor any, L1 any] func(*Actor, L1) error
type L2Func[Actor any, L1 any, L2 any] func(*Actor, L1, L2) error

// 远程
type Z0Func[Actor any] func(*Actor, define.IContext) error
type Z1Func[Actor any] func(*Actor, define.IContext, any) error

type P1Func[Actor any, P1 any] func(*Actor, define.IContext, *P1) error
type P2Func[Actor any, P1 any, P2 any] func(*Actor, define.IContext, *P1, *P2) error

type G1Func[Actor any, V1 any] func(*Actor, define.IContext, V1) error
type G2Func[Actor any, V1 any, V2 any] func(*Actor, define.IContext, V1, V2) error

package handler

import "framework/define"

type ToIdFunc func(string) uint32                   // string转uint32
type ToNameFunc func(uint32) string                 // uint32转string
type GetRpcFunc func(int32, uint32) define.IHandler // 获取全局rpc
type GetCmdFunc func(int32) define.IHandler         // 获取cmd的rpc
type GetActorFunc func(string) define.IHandler      // 获取本地节点rpc

var (
	toIdFunc   ToIdFunc
	toNameFunc ToNameFunc
	cmdFunc    GetCmdFunc
	rpcFunc    GetRpcFunc
	actorFunc  GetActorFunc
)

func SetHandler(i ToIdFunc, n ToNameFunc, r GetRpcFunc, c GetCmdFunc, a GetActorFunc) {
	toIdFunc = i
	toNameFunc = n
	cmdFunc = c
	rpcFunc = r
	actorFunc = a
}

func Name2Id(val string) uint32 {
	return toIdFunc(val)
}

func Id2Name(val uint32) string {
	return toNameFunc(val)
}

func Get(args ...any) define.IHandler {
	switch len(args) {
	case 1:
		switch vv := args[0].(type) {
		case string:
			return actorFunc(vv)
		case int32:
			return cmdFunc(vv)
		}
	case 2:
		switch v1 := args[0].(type) {
		case int32:
			switch vv := args[1].(type) {
			case string:
				return rpcFunc(v1, Name2Id(vv))
			case uint32:
				return rpcFunc(v1, vv)
			}
		}
	}
	return nil
}

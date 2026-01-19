package framework

type GetHandlerFunc func(string) IHandler
type GetCmdFunc func(uint32) IRpc
type GetRpcFunc func(uint32, any) IRpc
type GetClusterFunc func(uint32) ICluster
type GetRouterFunc func(uint32, uint64) IRouter
type SendRspFunc func(any, ...PacketFunc) error

var (
	handlerGet     GetHandlerFunc
	handlerCmd     GetCmdFunc
	handlerRpc     GetRpcFunc
	clusterGet     GetClusterFunc
	routerGet      GetRouterFunc
	routerGetOrNew GetRouterFunc
	sendRsp        SendRspFunc
)

func SetBus(f SendRspFunc) {
	sendRsp = f
}

func SendResponse(msg any, funcs ...PacketFunc) error {
	return sendRsp(msg, funcs...)
}

func SetHandler(h GetHandlerFunc, c GetCmdFunc, r GetRpcFunc) {
	handlerGet = h
	handlerCmd = c
	handlerRpc = r
}

func GetHandler(name string) IHandler {
	return handlerGet(name)
}

func GetCmdRpc(cmd uint32) IRpc {
	return handlerCmd(cmd)
}

func GetRpc(nodeType uint32, id any) IRpc {
	return handlerRpc(nodeType, id)
}

func SetCluster(f GetClusterFunc) {
	clusterGet = f
}

func GetCluster(nodeType uint32) ICluster {
	return clusterGet(nodeType)
}

func SetRouter(g, n GetRouterFunc) {
	routerGet = g
	routerGetOrNew = n
}

func GetRouter(idType uint32, id uint64) IRouter {
	return routerGet(idType, id)
}

func GetOrNewRouter(idType uint32, id uint64) IRouter {
	return routerGetOrNew(idType, id)
}

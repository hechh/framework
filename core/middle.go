package core

// 解决循环引用
var (
	// handler
	getByName func(...string) IHandler
	getByCmd  func(uint32) IHandler
	getByRpc  func(uint32, any) IHandler
	// bus
	broadcast func(IPacket) error
	send      func(IPacket) error
	request   func(IPacket, func([]byte) error) error
	// router
	getRouter      func(uint32, uint64) IRouter
	getOrNewRouter func(uint32, uint64) IRouter
	// cluster
	getCluster func(uint32) ICluster
)

// --------------------------cluster----------------------------
func SetGetCluster(f func(uint32) ICluster) {
	getCluster = f
}

func GetCluster(nodeType uint32) ICluster {
	return getCluster(nodeType)
}

// --------------------------bus----------------------------
func SetBroadcast(f func(IPacket) error) {
	broadcast = f
}

func SetSend(f func(IPacket) error) {
	send = f
}

func SetRequest(f func(IPacket, func([]byte) error) error) {
	request = f
}

func Broadcast(pack IPacket) error {
	return broadcast(pack)
}

func Send(pack IPacket) error {
	return send(pack)
}

func Request(pack IPacket, cb func([]byte) error) error {
	return request(pack, cb)
}

// ---------handler----------
func SetGetHandler(fn func(...string) IHandler) {
	getByName = fn
}

func SetGetHandlerByCmd(fn func(uint32) IHandler) {
	getByCmd = fn
}

func SetGetHandlerByRpc(fn func(uint32, any) IHandler) {
	getByRpc = fn
}

func GetHandler(names ...string) IHandler {
	return getByName(names...)
}

func GetHandlerByCmd(cmd uint32) IHandler {
	return getByCmd(cmd)
}

func GetHandlerByRpc(nodeType uint32, id any) IHandler {
	return getByRpc(nodeType, id)
}

// ----------------router-------------
func SetGetRouter(f func(uint32, uint64) IRouter) {
	getRouter = f
}

func SetGetOrNewRouter(f func(uint32, uint64) IRouter) {
	getOrNewRouter = f
}

func GetRouter(idType uint32, id uint64) IRouter {
	return getRouter(idType, id)
}

func GetOrNewRouter(idType uint32, id uint64) IRouter {
	return getOrNewRouter(idType, id)
}

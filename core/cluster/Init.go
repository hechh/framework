package cluster

import "github.com/hechh/framework/packet"

var object *Cluster

func SetObject(oj *Cluster) {
	object = oj
}

func Count(nodeType uint32) int {
	if object != nil {
		return object.Count(nodeType)
	}
	return 0
}

func Total() int {
	if object != nil {
		return object.Total()
	}
	return 0
}

func Get(nodeType, nodeId uint32) *packet.Node {
	if object != nil {
		return object.Get(nodeType, nodeId)
	}
	return nil
}

func Gets(nodeType uint32) []*packet.Node {
	if object != nil {
		return object.Gets(nodeType)
	}
	return nil
}

func HashRoute(nodeType uint32, seed uint64) *packet.Node {
	if object != nil {
		return object.Route(nodeType, seed)
	}
	return nil
}

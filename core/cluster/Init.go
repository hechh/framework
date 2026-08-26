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
		if node := object.Get(nodeType, nodeId); node != nil {
			if pn, ok := node.(*packet.Node); ok {
				return pn
			}
		}
	}
	return nil
}

func Gets(nodeType uint32) []*packet.Node {
	if object != nil {
		nodes := object.Gets(nodeType)
		result := make([]*packet.Node, 0, len(nodes))
		for _, n := range nodes {
			if pn, ok := n.(*packet.Node); ok {
				result = append(result, pn)
			}
		}
		return result
	}
	return nil
}

func HashRoute(nodeType uint32, seed uint64) *packet.Node {
	if object != nil {
		if node := object.Route(nodeType, seed); node != nil {
			if pn, ok := node.(*packet.Node); ok {
				return pn
			}
		}
	}
	return nil
}

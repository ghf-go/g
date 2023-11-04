package g

import "strings"

type router_web_node struct {
	mid   []GHandlerFunc
	nodes map[string]*router_web_node
	hf    map[string]map[string]GHandlerFunc
}

// 添加路由
func (n *router_web_node) add(name, method string, fen GHandlerFunc) {
	if strings.HasPrefix(name, "/") {
		name = name[1:]
	}
	if strings.HasPrefix(name, "/") {
		name = name[0 : len(name)-1]
	}
	if nnn, ok := n.hf[name]; !ok {
		nnn[method] = fen
		n.hf[name] = nnn
	} else {
		n.hf[name] = map[string]GHandlerFunc{
			method: fen,
		}
	}
}

// 添加分组
func (n *router_web_node) addGroup(name string, m ...GHandlerFunc) *router_web_node {
	ret := &router_web_node{
		mid:   m,
		nodes: map[string]*router_web_node{},
		hf:    map[string]map[string]GHandlerFunc{},
	}
	if strings.HasPrefix(name, "/") {
		name = name[1:]
	}
	if strings.HasPrefix(name, "/") {
		name = name[0 : len(name)-1]
	}
	n.nodes[name] = ret
	return ret
}

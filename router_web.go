package g

import (
	"errors"
	"net/http"
	"strings"
)

type router_web_node struct {
	mid   []GHandlerFunc
	nodes map[string]*router_web_node
	hf    map[string]map[string]GHandlerFunc
}

// 添加路由
func (n *router_web_node) add(name, method string, fen GHandlerFunc) {
	if nnn, ok := n.hf[name]; ok {
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
	n.nodes[name] = ret
	return ret
}

// 获取请求地址
func (n *router_web_node) getHandle(path, method string, m []GHandlerFunc) ([]GHandlerFunc, error) {
	ret := append(m, n.mid...)
	if p, ok := n.hf[path]; ok {
		if h, ok := p[method]; ok {
			ret = append(ret, h)
			return ret, nil
		} else {
			return ret, errors.New("Method is exists")
		}
	}
	ns := strings.Split(path, "/")
	lens := len(ns)
	for i := lens - 1; i >= 0; i-- {
		ps := strings.Join(ns[0:i], "/")
		if aa, ok := n.nodes[ps]; ok {
			return aa.getHandle("/"+strings.Join(ns[i:], "/"), method, ret)
		}
	}
	return ret, errors.New("path not found")
}

// 网页路由
func (ge *router_web_node) Any(name string, fen GHandlerFunc) *router_web_node {
	ge.add(name, http.MethodGet, fen)
	ge.add(name, http.MethodPost, fen)
	ge.add(name, http.MethodPut, fen)
	ge.add(name, http.MethodPatch, fen)
	ge.add(name, http.MethodDelete, fen)
	ge.add(name, http.MethodHead, fen)
	ge.add(name, http.MethodTrace, fen)
	ge.add(name, http.MethodOptions, fen)
	return ge
}
func (ge *router_web_node) Post(name string, fen GHandlerFunc) *router_web_node {
	ge.add(name, http.MethodPost, fen)
	return ge
}
func (ge *router_web_node) Get(name string, fen GHandlerFunc) *router_web_node {
	ge.add(name, http.MethodGet, fen)
	return ge
}
func (ge *router_web_node) Delete(name string, fen GHandlerFunc) *router_web_node {
	ge.add(name, http.MethodDelete, fen)
	return ge
}
func (ge *router_web_node) Put(name string, fen GHandlerFunc) *router_web_node {
	ge.add(name, http.MethodPut, fen)
	return ge
}
func (ge *router_web_node) Options(name string, fen GHandlerFunc) *router_web_node {
	ge.add(name, http.MethodOptions, fen)
	return ge
}
func (ge *router_web_node) Trace(name string, fen GHandlerFunc) *router_web_node {
	ge.add(name, http.MethodTrace, fen)
	return ge
}
func (ge *router_web_node) Head(name string, fen GHandlerFunc) *router_web_node {
	ge.add(name, http.MethodHead, fen)
	return ge
}
func (ge *router_web_node) Patch(name string, fen GHandlerFunc) *router_web_node {
	ge.add(name, http.MethodPatch, fen)
	return ge
}
func (ge *router_web_node) Group(name string, fen ...GHandlerFunc) *router_web_node {
	return ge.addGroup(name, fen...)
}

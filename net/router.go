package net

import (
	"log"
	"strings"
	"sync"
)

type HandlerFunc func(req *WsMsgReq, rsp *WsMsgRsp)

type Router struct {
	group []*group
}

func NewRouter() *Router {
	return &Router{}
}

type MiddlewareFunc func(handlerFunc HandlerFunc) HandlerFunc

// 路由组
type group struct {
	mutex         sync.RWMutex
	prefix        string
	handlerMap    map[string]HandlerFunc //存储路由
	middlewareMap map[string][]MiddlewareFunc
	middlewares   []MiddlewareFunc
}

func (g *group) AddRouter(name string, HandlerFunc HandlerFunc, middlewares ...MiddlewareFunc) {
	g.mutex.Lock()
	defer g.mutex.Unlock()
	g.handlerMap[name] = HandlerFunc
	g.middlewareMap[name] = middlewares
}

func (g *group) Use(middlewares ...MiddlewareFunc) {
	g.middlewares = append(g.middlewares, middlewares...)
}

func (r *Router) Group(prefix string) *group {
	g := &group{
		prefix:        prefix,
		handlerMap:    make(map[string]HandlerFunc),
		middlewareMap: make(map[string][]MiddlewareFunc),
	}
	r.group = append(r.group, g)
	return g
}

func (g *group) exec(name string, req *WsMsgReq, rsp *WsMsgRsp) {
	h, ok := g.handlerMap[name]

	if !ok {
		h, ok = g.handlerMap["*"]
		if !ok {
			log.Println("路由未定义!")
		}
	}
	if ok {
		for i := 0; i < len(g.middlewares); i++ {
			h = g.middlewares[i](h)
		}
		mm, ok := g.middlewareMap[name]
		if ok {
			for i := 0; i < len(mm); i++ {
				h = mm[i](h)
			}
		}
		h(req, rsp)
	}
}

func (r *Router) Run(req *WsMsgReq, rsp *WsMsgRsp) {
	// account.login
	strs := strings.Split(req.Body.Name, ".")
	prefix := ""
	name := ""
	if len(strs) == 2 {
		prefix = strs[0]
		name = strs[1]
	}
	for _, g := range r.group {
		if g.prefix == prefix {
			g.exec(name, req, rsp)
		} else if g.prefix == "*" {
			g.exec(name, req, rsp)
		}
	}
}

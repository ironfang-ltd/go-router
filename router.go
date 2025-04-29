package router

import (
	"context"
	"net/http"
)

const (
	ErrPathMustStartWithSlash  = "path must start with '/'"
	ErrPathMustNotEndWithSlash = "path must not end with '/'"

	PathSep = '/'
)

type Middleware func(http.HandlerFunc) http.HandlerFunc

type Router interface {
	RouteGroup
	GetRoutes() []RouteDescriptor
	ServeHTTP(w http.ResponseWriter, r *http.Request)
}

type RouteGroup interface {
	Get(path string, handler http.HandlerFunc) Route
	Post(path string, handler http.HandlerFunc) Route
	Put(path string, handler http.HandlerFunc) Route
	Patch(path string, handler http.HandlerFunc) Route
	Delete(path string, handler http.HandlerFunc) Route
	Group(prefix string) RouteGroup
	Use(middleware ...Middleware)
}

type Route interface {
	// Use(middleware ...Middleware)
}

type RouteDescriptor struct {
	Method string
	Path   string
}

type router struct {
	config *Config
	node   *routeTreeNode
}

func New(opts ...Option) Router {

	config := &Config{
		NotFoundHandler: func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		},
		MethodNotAllowedHandler: func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusMethodNotAllowed)
		},
	}

	for _, opt := range opts {
		opt(config)
	}

	rtr := &router{
		config: config,
		node:   newRouteTreeNode(config),
	}

	return rtr
}

func (rtr *router) Get(path string, handler http.HandlerFunc) Route {
	return rtr.mapMethod(http.MethodGet, path, handler)
}

func (rtr *router) Post(path string, handler http.HandlerFunc) Route {
	return rtr.mapMethod(http.MethodPost, path, handler)
}

func (rtr *router) Put(path string, handler http.HandlerFunc) Route {
	return rtr.mapMethod(http.MethodPut, path, handler)
}

func (rtr *router) Patch(path string, handler http.HandlerFunc) Route {
	return rtr.mapMethod(http.MethodPatch, path, handler)
}

func (rtr *router) Delete(path string, handler http.HandlerFunc) Route {
	return rtr.mapMethod(http.MethodDelete, path, handler)
}

func (rtr *router) Group(prefix string) RouteGroup {

	node := rtr.node.GetOrCreateNode(prefix)

	group := &router{
		config: rtr.config,
		node:   node,
	}

	return group
}

func (rtr *router) Use(middleware ...Middleware) {
	rtr.node.Use(middleware...)
}

func (rtr *router) GetRoutes() []RouteDescriptor {

	var routes []RouteDescriptor

	q := []*routeTreeNode{rtr.node}

	for {
		if len(q) == 0 {
			break
		}

		node := q[0]
		q = q[1:]

		if node == nil {
			break
		}

		for i, handler := range node.handlers {
			if handler != nil {

				p := node.getPath()
				if len(p) == 0 {
					p = "/"
				}

				routes = append(routes, RouteDescriptor{
					Method: uint8ToMethod(uint8(i)),
					Path:   p,
				})
			}
		}

		q = append(q, node.children...)
	}

	return routes
}

func (rtr *router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	node, found := rtr.node.Find(r)

	if !found {
		ctx := context.WithValue(r.Context(), "FOUND", false)
		r = r.WithContext(ctx)
	}

	node.handler(w, r)
}

func (rtr *router) mapMethod(method, path string, handler http.HandlerFunc) *routeTreeNode {

	if len(path) == 0 || path[0] != PathSep {
		panic(ErrPathMustStartWithSlash)
	}

	if path == "/" {
		path = ""
	}

	if len(path) > 1 && path[len(path)-1] == PathSep {
		panic(ErrPathMustNotEndWithSlash)
	}

	node := rtr.node.GetOrCreateNode(path)
	node.SetHandler(method, handler)
	return node
}

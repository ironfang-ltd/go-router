package router

import (
	"fmt"
	"net/http"
	"sort"
	"strings"
)

type RouteMatch int

const (
	RouteMatchExact  RouteMatch = 0
	RouteMatchPrefix RouteMatch = 1
)

const (
	httpMethodGet uint8 = iota
	httpMethodHead
	httpMethodPost
	httpMethodPut
	httpMethodPatch
	httpMethodDelete
	httpMethodConnect
	httpMethodOptions
	httpMethodTrace
	httpMethodAny
	httpMethodCount
)

type routeTreeNode struct {
	config      *Config
	segment     string
	parent      *routeTreeNode
	children    []*routeTreeNode
	middlewares []Middleware
	handler     http.HandlerFunc
	middleware  http.HandlerFunc
	handlers    []http.HandlerFunc
	param       bool
	catchAll    bool
}

func newRouteTreeNode(config *Config) *routeTreeNode {
	node := &routeTreeNode{
		config:   config,
		segment:  "",
		parent:   nil,
		children: nil,
		handler:  nil,
		handlers: nil,
		param:    false,
		catchAll: false,
	}

	node.handler = node.final
	node.middleware = node.config.MethodNotAllowedHandler

	return node
}

func (r *routeTreeNode) GetOrCreateNode(path string) *routeTreeNode {

	node := r
	high := 0

	for {
		if len(path) == 0 {
			break
		}

		high = strings.IndexByte(path, PathSep)
		if high == -1 {
			high = len(path)
		}

		segment := path[:high]

		if segment == "" {
			node = r
			high++
			if high >= len(path) {
				break
			}
			path = path[high:]
			continue
		}

		found := false

		for _, child := range node.children {
			if child.segment == segment {
				// TODO: Check for conflicting param/catchAll
				node = child
				found = true
				break
			}
		}

		if !found {
			newNode := newRouteTreeNode(r.config)
			newNode.segment = segment
			newNode.parent = node
			newNode.param = segment[0] == ':'
			newNode.catchAll = segment[0] == '*'

			node.children = append(node.children, newNode)

			sort.Slice(node.children, func(i, j int) bool {
				// Sort Order: segment(static) > param > catchAll
				return nodePriority(node.children[i]) < nodePriority(node.children[j])
			})

			node = newNode
		}

		if segment == "*" {
			break
		}

		if high >= len(path) {
			break
		}

		high++
		path = path[high:]
	}

	return node
}

func (r *routeTreeNode) Find(req *http.Request) (*routeTreeNode, RouteMatch) {

	path := req.URL.Path

	if path == "" || path == "/" {
		return r, RouteMatchExact
	}

	node := r
	high := 0

	if path[0] == PathSep {
		path = path[1:]
	}

	for {
		if len(path) == 0 {
			break
		}

		high = strings.IndexByte(path, PathSep)
		if high == -1 {
			high = len(path)
		}

		segment := path[:high]
		high++
		found := false

		for _, child := range node.children {

			if child.param {

				req.SetPathValue(child.segment[1:], segment)

				if high >= len(path) {
					return child, RouteMatchExact
				}

				node = child
				found = true
				path = path[high:]
				break
			} else if child.segment == segment {
				if high >= len(path) {
					return child, RouteMatchExact
				}

				node = child
				found = true
				path = path[high:]
				break
			} else if child.catchAll {
				return child, RouteMatchExact
			}
		}

		if !found {
			break
		}
	}

	return node, RouteMatchPrefix
}

func (r *routeTreeNode) SetHandler(method string, handler http.HandlerFunc) {
	if r.handlers == nil {
		r.handlers = make([]http.HandlerFunc, httpMethodCount)
	}

	r.handlers[methodToUint8(method)] = handler
	r.handler = r.wrapMiddleware(r.final)
}

func (r *routeTreeNode) GetHandler(method string) http.HandlerFunc {

	if r.handlers == nil {
		return nil
	}

	if r.handlers[httpMethodAny] != nil {
		return r.handlers[httpMethodAny]
	}

	return r.handlers[methodToUint8(method)]
}

func (r *routeTreeNode) Use(middleware ...Middleware) {
	r.middlewares = append(r.middlewares, middleware...)
	r.handler = r.wrapMiddleware(r.final)
	r.middleware = r.wrapMiddleware(r.config.NotFoundHandler)
}

func (r *routeTreeNode) wrapMiddleware(final http.HandlerFunc) http.HandlerFunc {

	// collect all middlewares from parent nodes and current node
	middlewares := make([]Middleware, 0)
	node := r

	for {
		if node == nil {
			break
		}

		// reverse order of middlewares
		for i := len(node.middlewares) - 1; i >= 0; i-- {
			middlewares = append(middlewares, node.middlewares[i])
		}

		// prepend parent middlewares
		node = node.parent
	}

	for i := range middlewares {
		final = middlewares[i](final)
	}

	return final
}

func (r *routeTreeNode) final(w http.ResponseWriter, req *http.Request) {

	handler := r.GetHandler(req.Method)

	if handler == nil {
		if len(r.handlers) == 0 {
			r.config.NotFoundHandler(w, req)
		} else {
			w.Header().Add("Allow", strings.Join(r.getAllowedMethods(), ", "))
			r.config.MethodNotAllowedHandler(w, req)
		}
		return
	}

	handler(w, req)
}

func (r *routeTreeNode) getAllowedMethods() []string {
	var allowedMethods []string
	for method := range r.handlers {
		allowedMethods = append(allowedMethods, uint8ToMethod(uint8(method)))
	}
	return allowedMethods
}

func (r *routeTreeNode) getPath() string {

	if r.parent == nil {
		return r.segment
	}

	return r.parent.getPath() + "/" + r.segment
}

func nodePriority(node *routeTreeNode) int {

	if node.catchAll {
		return 3
	}

	if node.param {
		return 2
	}

	// static
	return 1
}

func methodToUint8(method string) uint8 {

	switch method {
	case http.MethodGet:
		return httpMethodGet
	case http.MethodHead:
		return httpMethodHead
	case http.MethodPost:
		return httpMethodPost
	case http.MethodPut:
		return httpMethodPut
	case http.MethodPatch:
		return httpMethodPatch
	case http.MethodDelete:
		return httpMethodDelete
	case http.MethodConnect:
		return httpMethodConnect
	case http.MethodOptions:
		return httpMethodOptions
	case http.MethodTrace:
		return httpMethodTrace
	case "*":
		return httpMethodAny
	}

	return httpMethodGet
}

func uint8ToMethod(method uint8) string {

	switch method {
	case httpMethodGet:
		return http.MethodGet
	case httpMethodHead:
		return http.MethodHead
	case httpMethodPost:
		return http.MethodPost
	case httpMethodPut:
		return http.MethodPut
	case httpMethodPatch:
		return http.MethodPatch
	case httpMethodDelete:
		return http.MethodDelete
	case httpMethodConnect:
		return http.MethodConnect
	case httpMethodOptions:
		return http.MethodOptions
	case httpMethodTrace:
		return http.MethodTrace
	case httpMethodAny:
		return "*"
	default:
		panic(fmt.Errorf("unknown method: %d", method))
	}
}

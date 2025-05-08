package router

import "net/http"

type Option func(*Config)

type Config struct {
	NotFoundHandler         http.HandlerFunc
	MethodNotAllowedHandler http.HandlerFunc
	MiddlewareMatch         RouteMatch
}

func WithMiddlewareMatch(match RouteMatch) Option {
	return func(c *Config) {
		c.MiddlewareMatch = match
	}
}

func WithNotFoundHandler(handler http.HandlerFunc) Option {

	if handler == nil {
		panic("not found handler must not be nil")
	}

	return func(c *Config) {
		c.NotFoundHandler = handler
	}
}

func WithMethodNotAllowedHandler(handler http.HandlerFunc) Option {

	if handler == nil {
		panic("method not allowed handler must not be nil")
	}

	return func(c *Config) {
		c.MethodNotAllowedHandler = handler
	}
}

package middleware

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/ironfang-ltd/go-router"
)

type CorsOption func(*CorsOptions)

type CorsOptions struct {
	AllowedOrigins   []string
	AllowedMethods   []string
	AllowedHeaders   []string
	AllowCredentials bool
	ExposedHeaders   []string
	MaxAge           int
}

func WithAllowedOrigins(origins ...string) func(*CorsOptions) {
	return func(opts *CorsOptions) {
		opts.AllowedOrigins = origins
	}
}

func WithAllowedMethods(methods ...string) func(*CorsOptions) {
	return func(opts *CorsOptions) {
		opts.AllowedMethods = methods
	}
}

func WithAllowedHeaders(headers ...string) func(*CorsOptions) {
	return func(opts *CorsOptions) {
		opts.AllowedHeaders = headers
	}
}

func WithAllowCredentials(allow bool) func(*CorsOptions) {
	return func(opts *CorsOptions) {
		opts.AllowCredentials = allow
	}
}

func Cors(options ...CorsOption) router.Middleware {

	opts := &CorsOptions{
		AllowedOrigins:   []string{},
		AllowedMethods:   []string{"HEAD", "OPTIONS", "GET", "POST", "PUT", "PATCH", "DELETE"},
		AllowedHeaders:   []string{},
		AllowCredentials: false,
		ExposedHeaders:   []string{},
		MaxAge:           600,
	}

	for _, option := range options {
		option(opts)
	}

	allowedHeaders := strings.Join(opts.AllowedHeaders, ", ")
	exposedHeaders := strings.Join(opts.ExposedHeaders, ", ")

	isPreflightRequest := func(r *http.Request) bool {
		return r.Header.Get("Origin") != "" && r.Method == http.MethodOptions && r.Header.Get("Access-Control-Request-Method") != ""
	}

	handlePreflightRequest := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Vary", "Access-Control-Request-Method")
		w.Header().Add("Vary", "Access-Control-Request-Headers")

		origin := r.Header.Get("Origin")

		// Check if the origin is allowed
		if !isOriginAllowed(opts.AllowedOrigins, origin) {
			slog.Info("CORS origin not allowed", "origin", origin)
			return
		}

		// Check if the method is allowed
		if !isMethodAllowed(opts.AllowedMethods, r.Header.Get("Access-Control-Request-Method")) {
			slog.Info("CORS method not allowed", "method", r.Header.Get("Access-Control-Request-Method"))
			return
		}

		// Check if the headers are allowed
		if !isHeadersAllowed(opts.AllowedHeaders, r.Header.Get("Access-Control-Request-Headers")) {
			slog.Info("CORS headers not allowed", "headers", r.Header.Get("Access-Control-Request-Headers"))
			return
		}

		// Set the allowed origin
		if len(opts.AllowedOrigins) > 0 {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
		}

		// Set the allowed method
		w.Header().Set("Access-Control-Allow-Methods", strings.ToUpper(r.Header.Get("Access-Control-Request-Method")))

		// Set the allowed headers if set
		if len(opts.AllowedHeaders) > 0 {
			w.Header().Set("Access-Control-Allow-Headers", allowedHeaders)
		}

		// Check if allow credentials is set
		if opts.AllowCredentials {
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		}

		// Set the max age header if set
		if opts.MaxAge > 0 {
			w.Header().Set("Access-Control-Max-Age", strconv.Itoa(opts.MaxAge))
		}
	}

	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {

			fmt.Printf("running cors middleware...\n")

			w.Header().Add("Vary", "Origin")

			if isPreflightRequest(r) {
				handlePreflightRequest(w, r)
				w.WriteHeader(http.StatusOK)
				return
			}

			origin := r.Header.Get("Origin")
			if origin == "" {
				next(w, r)
				return
			}

			// Check if the origin is allowed
			if !isOriginAllowed(opts.AllowedOrigins, origin) {
				slog.Info("CORS origin not allowed", "origin", origin)
				w.WriteHeader(http.StatusForbidden)
				return
			}

			// Check if the method is allowed
			if !isMethodAllowed(opts.AllowedMethods, r.Method) {
				slog.Info("CORS method not allowed", "method", r.Method)
				w.WriteHeader(http.StatusForbidden)
				return
			}

			// Set the allowed origin
			if len(opts.AllowedOrigins) > 0 {
				w.Header().Set("Access-Control-Allow-Origin", origin)
			} else {
				w.Header().Set("Access-Control-Allow-Origin", "*")
			}

			// Set the exposed headers
			if len(opts.ExposedHeaders) > 0 {
				w.Header().Add("Access-Control-Expose-Headers", exposedHeaders)
			}

			// Check if allow credentials is set and set it
			if opts.AllowCredentials {
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}

			next(w, r)
		}
	}
}

func isOriginAllowed(origins []string, requestedOrigin string) bool {
	if len(origins) == 0 {
		return true
	}

	if requestedOrigin == "" {
		return false
	}

	for _, origin := range origins {
		if strings.EqualFold(requestedOrigin, origin) {
			return true
		}
	}

	return false
}

func isHeadersAllowed(allowedHeaders []string, requestedHeaders string) bool {
	if len(allowedHeaders) == 0 {
		return true
	}

	requested := strings.Split(requestedHeaders, ", ")

	for _, requestedHeader := range requested {
		for _, allowedHeader := range allowedHeaders {
			if strings.EqualFold(requestedHeader, allowedHeader) {
				return true
			}
		}
	}

	return false
}

func isMethodAllowed(methods []string, requestedMethod string) bool {
	if len(methods) == 0 {
		return true
	}

	for _, method := range methods {
		if strings.EqualFold(requestedMethod, method) {
			return true
		}
	}

	return false
}

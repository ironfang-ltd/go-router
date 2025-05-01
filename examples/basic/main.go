package main

import (
	"log/slog"
	"net/http"

	"github.com/ironfang-ltd/go-router"
)

func main() {

	r := router.New()

	r.Use(func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			slog.Debug("middleware executing...")
			w.Header().Set("X-Test", "test")
			//next(w, r)
		}
	})

	/*r.Use(middleware.Cors(middleware.WithAllowedOrigins("http://localhost:3000")))

	// Simple
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, World!"))
	})

	// With Param
	r.Get("/:name", func(w http.ResponseWriter, r *http.Request) {
		name := r.PathValue("name")
		w.Write([]byte("Hello, " + name + "!"))
	})

	// With Group
	apiGroup := r.Group("/api")

	apiGroup.Use(middleware.Cors(middleware.WithAllowedMethods("OPTIONS", "POST")))

	apiGroup.Post("/hello", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, World from api!"))
	})*/

	err := http.ListenAndServe("127.0.0.1:8080", r)
	if err != nil {
		panic(err)
	}
}

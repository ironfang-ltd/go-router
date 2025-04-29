package main

import (
	"net/http"

	"github.com/ironfang-ltd/go-router"
	"github.com/ironfang-ltd/go-router/middleware"
)

func main() {

	r := router.New()

	r.Use(middleware.Cors(middleware.WithAllowedOrigins("http://localhost:3000")))

	// Simple
	r.Get("/asd", func(w http.ResponseWriter, r *http.Request) {
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
	})

	err := http.ListenAndServe("127.0.0.1:5000", r)
	if err != nil {
		panic(err)
	}
}

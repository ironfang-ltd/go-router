package router

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRouter_GetWithNoRoutes(t *testing.T) {

	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	r := New()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("response code is: %d, expected: %d", w.Code, http.StatusNotFound)
	}
}

func TestRouter_Get(t *testing.T) {

	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	r := New()

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	})

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("response code is: %d, expected: %d", w.Code, http.StatusOK)
	}
}

func TestRouter_GetMux(t *testing.T) {

	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	})

	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("response code is: %d, expected: %d", w.Code, http.StatusOK)
	}
}

func TestRouter_GetWithParam(t *testing.T) {

	req, _ := http.NewRequest("GET", "/test-value", nil)
	w := httptest.NewRecorder()

	r := New()

	r.Get("/:param", func(w http.ResponseWriter, r *http.Request) {

		param := r.PathValue("param")

		_, _ = w.Write([]byte(param))
	})

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("response code is: %d, expected: %d", w.Code, http.StatusOK)
	}

	if w.Body.String() != "test-value" {
		t.Errorf("response body is: %s, expected: %s", w.Body.String(), "test-value")
	}
}

func TestRouter_GetWithMiddleware(t *testing.T) {

	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	r := New()

	r.Use(func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Test", "test")
			next(w, r)
		}
	})

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("response code is: %d, expected: %d", w.Code, http.StatusOK)
	}

	if w.Header().Get("X-Test") != "test" {
		t.Errorf("response header X-Test is: %s, expected: %s", w.Header().Get("X-Test"), "test")
	}
}

func TestRouter_GetWithGroup(t *testing.T) {

	req, _ := http.NewRequest("GET", "/group/endpoint", nil)
	w := httptest.NewRecorder()

	r := New()

	g := r.Group("/group")

	g.Use(func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Test", "test")
			next(w, r)
		}
	})

	g.Get("/endpoint", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("group-endpoint"))
	})

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("response code is: %d, expected: %d", w.Code, http.StatusOK)
	}

	if w.Header().Get("X-Test") != "test" {
		t.Errorf("response header X-Test is: %s, expected: %s", w.Header().Get("X-Test"), "test")
	}
}

func TestRouter_OptionsWithMultipleMiddleware(t *testing.T) {

	req, _ := http.NewRequest("OPTIONS", "/group/endpoint", nil)
	w := httptest.NewRecorder()

	r := New()

	r.Use(func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Test", "test")
			next(w, r)
		}
	})

	g := r.Group("/group")

	g.Use(func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Test-2", "test-2")
			//next(w, r)
		}
	})

	g.Get("/endpoint", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("group-endpoint"))
	})

	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("response code is: %d, expected: %d", w.Code, http.StatusNotFound)
	}

	if w.Header().Get("X-Test") != "test" {
		t.Errorf("response header X-Test is: %s, expected: %s", w.Header().Get("X-Test"), "test")
	}

	if w.Header().Get("X-Test-2") != "test-2" {
		t.Errorf("response header X-Test-2 is: %s, expected: %s", w.Header().Get("X-Test-2"), "test-2")
	}
}

func TestRouter_NodeOrder(t *testing.T) {

	r := New()

	r.Get("/:param", func(w http.ResponseWriter, r *http.Request) {})
	r.Get("/*catchall", func(w http.ResponseWriter, r *http.Request) {})
	r.Get("/test", func(w http.ResponseWriter, r *http.Request) {})

	routes := r.GetRoutes()

	if routes[0].Path != "/test" {
		t.Errorf("first route '%s', expected: '%s'", routes[0].Path, "/test")
	}

	if routes[1].Path != "/:param" {
		t.Errorf("second route '%s', expected: '%s'", routes[1].Path, "/:param")
	}

	if routes[2].Path != "/*catchall" {
		t.Errorf("third route '%s', expected: '%s'", routes[2].Path, "/*catchall")
	}
}

func BenchmarkGet_SingleRoot(b *testing.B) {

	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	r := New()

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	})

	for n := 0; n < b.N; n++ {
		r.ServeHTTP(w, req)
	}
}

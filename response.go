package router

import "net/http"

type response struct {
	http.ResponseWriter
	written bool
}

func (rw *response) WriteHeader(statusCode int) {
	rw.written = true
	rw.ResponseWriter.WriteHeader(statusCode)
}

func (rw *response) Write(b []byte) (int, error) {
	return rw.ResponseWriter.Write(b)
}

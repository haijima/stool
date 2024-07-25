package main

import (
	"net/http"
)

type PathHandler struct{}

func (h *PathHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(r.URL.Path))
}

func main() {
	h := &PathHandler{}
	http.Handle("GET /foo/{id}", h)
	http.HandleFunc("POST /foo/{id}", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("POST /foo/{id} id=" + r.PathValue("id")))
	})

	mux := http.NewServeMux()
	mux.Handle("GET /foo/{foo_id}/bar/{bar_id}", h)
	mux.HandleFunc("GET /foo/{id...}", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("GET /foo/{id...} id...=" + r.PathValue("id")))
	})
	mux.HandleFunc("GET /foo/{$}", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("GET /foo/{$}"))
	})
	mux.HandleFunc("/foo/{id}", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("GET /foo/{id} id=" + r.PathValue("id")))
	})
	mux.HandleFunc("GET /foo/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("GET /foo/"))
	})
	mux.HandleFunc("GET /foo", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("GET /foo"))
	})
	mux.HandleFunc("GET /foo/bar", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("GET /foo/bar"))
	})
	http.ListenAndServe(":8080", mux)
}

package httphelp

import (
	"net/http"
)

type SimpleHandler map[string]http.Handler

func NewSimpleHandler() SimpleHandler {
	return SimpleHandler(make(map[string]http.Handler))
}

func (s SimpleHandler) Add(path string, h http.Handler) {
	s[path] = h
}

func (s SimpleHandler) Rem(path string) {
	delete(map[string]http.Handler(s), path)
}

func (s SimpleHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	lookup := r.URL.Path
	if h, ok := map[string]http.Handler(s)[lookup]; ok {
		h.ServeHTTP(w, r)
		return
	}
	http.NotFoundHandler().ServeHTTP(w, r)
}

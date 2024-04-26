package api

import (
	"net/http"
)

type Route struct {
	path    string
	handler func(w http.ResponseWriter, r *http.Request)
}

type Methods string

const (
	Post   Methods = "POST"
	Get    Methods = "GET"
	Put    Methods = "PUT"
	Delete Methods = "DELETE"
)

func SetRoute(path string, method Methods, handler func(w http.ResponseWriter, r *http.Request)) {
	http.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		if validator(w, r) {
			handler(w, r)
		} else {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte((`"error": "bad call"`)))
		}
	})
}

// Runs before each route
func validator(w http.ResponseWriter, r *http.Request) bool {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	return true
}

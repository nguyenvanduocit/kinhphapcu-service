package middleware

import "net/http"

func AddApiHeader(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/vnd.api+json")
		h.ServeHTTP(w, r)
	})
}
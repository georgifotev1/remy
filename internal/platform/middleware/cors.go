package middleware

import (
	"net/http"
	"slices"
	"strings"
)

func CORS(allowedOrigins []string) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if origin != "" && (slices.Contains(allowedOrigins, "*") || slices.Contains(allowedOrigins, origin)) {
				h := w.Header()
				h.Set("Access-Control-Allow-Origin", origin)
				h.Add("Vary", "Origin")
				h.Set("Access-Control-Allow-Methods", strings.Join([]string{
					http.MethodGet, http.MethodPost, http.MethodPut,
					http.MethodPatch, http.MethodDelete, http.MethodOptions,
				}, ", "))
				h.Set("Access-Control-Allow-Headers", "Content-Type, Authorization, "+RequestIDHeader)
				h.Set("Access-Control-Max-Age", "600")

				// Preflight
				if r.Method == http.MethodOptions {
					w.WriteHeader(http.StatusNoContent)
					return
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}

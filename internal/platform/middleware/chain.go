package middleware

import "net/http"

type Middleware func(http.Handler) http.Handler

// Chain wraps h with the given middlewares. The first middleware is the
// outermost, so Chain(h, A, B, C) runs A -> B -> C -> h.
func Chain(h http.Handler, mws ...Middleware) http.Handler {
	for i := len(mws) - 1; i >= 0; i-- {
		h = mws[i](h)
	}
	return h
}

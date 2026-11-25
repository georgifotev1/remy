package main

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"runtime/debug"
	"slices"
	"strings"
	"time"
)

type middleware func(http.Handler) http.Handler

// chain(h, A, B, C) runs A -> B -> C -> h.
func chain(h http.Handler, mws ...middleware) http.Handler {
	for i := range slices.Backward(mws) {
		h = mws[i](h)
	}
	return h
}

const requestIDHeader = "X-Request-ID"

func (app *application) requestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get(requestIDHeader)
		if id == "" || len(id) > 64 {
			b := make([]byte, 8)
			_, _ = rand.Read(b)
			id = hex.EncodeToString(b)
		}
		w.Header().Set(requestIDHeader, id)
		next.ServeHTTP(w, contextSetRequestID(r, id))
	})
}

type statusRecorder struct {
	http.ResponseWriter
	status      int
	bytes       int
	wroteHeader bool
}

func (r *statusRecorder) WriteHeader(code int) {
	if !r.wroteHeader {
		r.status = code
		r.wroteHeader = true
	}
	r.ResponseWriter.WriteHeader(code)
}

func (r *statusRecorder) Write(b []byte) (int, error) {
	if !r.wroteHeader {
		r.WriteHeader(http.StatusOK)
	}
	n, err := r.ResponseWriter.Write(b)
	r.bytes += n
	return n, err
}

// Unwrap lets http.ResponseController reach Flusher, Hijacker, etc.
func (r *statusRecorder) Unwrap() http.ResponseWriter { return r.ResponseWriter }

func (app *application) logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{
			ResponseWriter: w,
			status:         http.StatusOK,
		}

		next.ServeHTTP(rec, r)

		app.logger.Info("request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", rec.status,
			"bytes", rec.bytes,
			"duration", time.Since(start),
			"remote", r.RemoteAddr,
			"request_id", contextGetRequestID(r),
		)
	})
}

func (app *application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			pv := recover()
			if pv == nil {
				return
			}

			// Let net/http handle intentional aborts.
			if pv == http.ErrAbortHandler {
				panic(pv)
			}

			app.logger.Error("panic recovered",
				"panic", pv,
				"stack", string(debug.Stack()),
				"method", r.Method,
				"path", r.URL.Path,
				"request_id", contextGetRequestID(r),
			)
			w.Header().Set("Connection", "close")
			app.errorResponse(w, r, http.StatusInternalServerError, "the server encountered a problem and could not process your request")
		}()

		next.ServeHTTP(w, r)
	})
}

func (app *application) enableCORS(next http.Handler) http.Handler {
	allowed := app.config.cors.allowedOrigins
	methods := strings.Join([]string{
		http.MethodGet,
		http.MethodPost,
		http.MethodPut,
		http.MethodPatch,
		http.MethodDelete,
		http.MethodOptions,
	}, ", ")

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" && (slices.Contains(allowed, "*") || slices.Contains(allowed, origin)) {
			h := w.Header()
			h.Set("Access-Control-Allow-Origin", origin)
			h.Add("Vary", "Origin")
			h.Set("Access-Control-Allow-Methods", methods)
			h.Set("Access-Control-Allow-Headers", "Content-Type, Authorization, "+requestIDHeader)
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

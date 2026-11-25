package main

import (
	"context"
	"net/http"
)

type contextKey string

const requestIDContextKey = contextKey("request_id")

func contextSetRequestID(r *http.Request, id string) *http.Request {
	ctx := context.WithValue(r.Context(), requestIDContextKey, id)
	return r.WithContext(ctx)
}

func contextGetRequestID(r *http.Request) string {
	id, _ := r.Context().Value(requestIDContextKey).(string)
	return id
}

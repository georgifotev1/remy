package main

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

type contextKey string

const (
	requestIDContextKey = contextKey("request_id")
	userIDContextKey    = contextKey("user_id")
)

func contextSetRequestID(r *http.Request, id string) *http.Request {
	ctx := context.WithValue(r.Context(), requestIDContextKey, id)
	return r.WithContext(ctx)
}

func contextGetRequestID(r *http.Request) string {
	id, _ := r.Context().Value(requestIDContextKey).(string)
	return id
}

func contextSetUserID(r *http.Request, id uuid.UUID) *http.Request {
	ctx := context.WithValue(r.Context(), userIDContextKey, id)
	return r.WithContext(ctx)
}

// contextGetUserID panics if called on a route not wrapped in requireAuth.
// That's a programmer error, so failing loudly is intended.
func contextGetUserID(r *http.Request) uuid.UUID {
	id, ok := r.Context().Value(userIDContextKey).(uuid.UUID)
	if !ok {
		panic("missing user id in request context")
	}
	return id
}

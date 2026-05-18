package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

// RequestIDHeader is the HTTP header used to propagate a request id between
// clients, the gateway, and downstream services.
const RequestIDHeader = "X-Request-ID"

type ctxKey int

const requestIDKey ctxKey = iota

// RequestID is HTTP middleware that ensures every request has an X-Request-ID.
// If the client supplied one, it is preserved; otherwise a new UUID is
// generated. The id is exposed via the request context (see GetRequestID) and
// echoed back in the response headers.
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rid := r.Header.Get(RequestIDHeader)
		if rid == "" {
			rid = uuid.NewString()
		}
		w.Header().Set(RequestIDHeader, rid)
		ctx := context.WithValue(r.Context(), requestIDKey, rid)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetRequestID returns the request id stored in ctx by the RequestID
// middleware, or an empty string if no id has been set.
func GetRequestID(ctx context.Context) string {
	if v, ok := ctx.Value(requestIDKey).(string); ok {
		return v
	}
	return ""
}

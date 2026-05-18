package middleware

import (
	"log/slog"
	"net/http"
)

// Recovery is HTTP middleware that turns any panic from a downstream handler
// into a 500 response and logs the panic value. Without it, a panic would
// crash the server's goroutine for that connection.
func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rv := recover(); rv != nil {
				slog.ErrorContext(r.Context(), "panic recovered",
					"panic", rv,
					"method", r.Method,
					"path", r.URL.Path,
					"request_id", GetRequestID(r.Context()),
				)
				http.Error(w, "internal server error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

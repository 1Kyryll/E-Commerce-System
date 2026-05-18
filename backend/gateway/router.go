package gateway

import (
	"net/http"

	"github.com/1Kyryll/ecommerce-demo/backend/gateway/handlers"
	"github.com/1Kyryll/ecommerce-demo/backend/internal/middleware"
)

// NewRouter returns the gateway's root http.Handler with all routes registered
// and cross-cutting middleware applied. Middleware is composed outside-in:
// Recovery wraps everything (so it catches panics inside RequestID too), and
// RequestID runs first per request so the recovered panic log can reference it.
func NewRouter() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", handlers.Healthz)

	return middleware.Recovery(middleware.RequestID(mux))
}

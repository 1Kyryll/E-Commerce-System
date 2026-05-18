package gateway

import (
	"net/http"
	"time"

	"github.com/1Kyryll/ecommerce-demo/backend/gateway/auth"
	"github.com/1Kyryll/ecommerce-demo/backend/gateway/handlers"
	"github.com/1Kyryll/ecommerce-demo/backend/internal/middleware"
)

// Deps bundles the wiring NewRouter needs. main constructs this; the router
// itself stays unaware of how Service was built.
type Deps struct {
	AuthSvc       *auth.Service
	SessionTTL    time.Duration
	SecureCookies bool
}

func NewRouter(d Deps) http.Handler {
	mux := http.NewServeMux()

	// Public routes.
	mux.HandleFunc("GET /healthz", handlers.Healthz)

	authH := handlers.NewAuthHandlers(d.AuthSvc, d.SecureCookies)
	mux.HandleFunc("POST /auth/signup", authH.Signup)
	mux.HandleFunc("POST /auth/login", authH.Login)
	mux.HandleFunc("POST /auth/logout", authH.Logout)

	// Protected routes. mux.Handle takes an http.Handler so we wrap each
	// protected HandlerFunc in Auth.
	meH := handlers.NewMeHandlers(d.AuthSvc)
	protect := middleware.Auth(d.AuthSvc)
	mux.Handle("GET /me", protect(http.HandlerFunc(meH.Get)))

	return middleware.Recovery(middleware.RequestID(mux))
}

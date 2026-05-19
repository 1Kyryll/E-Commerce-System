package gateway

import (
	"net/http"
	"time"

	"github.com/1Kyryll/ecommerce-demo/backend/gateway/auth"
	"github.com/1Kyryll/ecommerce-demo/backend/gateway/handlers"
	"github.com/1Kyryll/ecommerce-demo/backend/internal/middleware"
)

// Deps bundles the wiring NewRouter needs. main constructs this; the router
// itself stays unaware of how each dependency was built.
type Deps struct {
	AuthSvc       *auth.Service
	ProductsH     *handlers.ProductHandlers
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

	// Public product browsing.
	mux.HandleFunc("GET /products", d.ProductsH.List)
	mux.HandleFunc("GET /products/{id}", d.ProductsH.Get)

	// Protected routes.
	meH := handlers.NewMeHandlers(d.AuthSvc)
	protect := middleware.Auth(d.AuthSvc)
	mux.Handle("GET /me", protect(http.HandlerFunc(meH.Get)))

	return middleware.Recovery(middleware.RequestID(mux))
}

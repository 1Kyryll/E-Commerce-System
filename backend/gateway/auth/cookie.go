package auth

import (
	"net/http"
	"time"

	"github.com/1Kyryll/ecommerce-demo/backend/internal/middleware"
)

// SessionCookie returns the cookie that carries token. secure must be true in
// production to ensure the cookie is only sent over HTTPS.
func SessionCookie(token string, ttl time.Duration, secure bool) *http.Cookie {
	return &http.Cookie{
		Name:     middleware.SessionCookieName,
		Value:    token,
		Path:     "/",
		MaxAge:   int(ttl.Seconds()),
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	}
}

// ClearSessionCookie returns a cookie whose Set-Cookie response tells the
// browser to discard any existing session.
func ClearSessionCookie(secure bool) *http.Cookie {
	return &http.Cookie{
		Name:     middleware.SessionCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	}
}

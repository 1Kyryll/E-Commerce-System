package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

// SessionCookieName is the canonical cookie name carrying the JWT session
// token. Defined here (not in gateway/auth) so the middleware can stay
// independent of the auth implementation.
const SessionCookieName = "session"

// Verifier validates a session token and returns the user id it represents.
// Returning any non-nil error causes the request to be rejected with 401.
type Verifier interface {
	VerifySession(token string) (uuid.UUID, error)
}

type userCtxKey int

const userIDKey userCtxKey = iota

// Auth produces middleware that requires a valid session cookie. The verifier
// abstracts JWT validation so this package doesn't depend on the auth
// implementation.
func Auth(v Verifier) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c, err := r.Cookie(SessionCookieName)
			if err != nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			uid, err := v.VerifySession(c.Value)
			if err != nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			ctx := context.WithValue(r.Context(), userIDKey, uid)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetUserID returns the authenticated user id put on ctx by Auth, or (zero, false)
// if Auth did not run.
func GetUserID(ctx context.Context) (uuid.UUID, bool) {
	v, ok := ctx.Value(userIDKey).(uuid.UUID)
	return v, ok
}

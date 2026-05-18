package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/1Kyryll/ecommerce-demo/backend/gateway/auth"
)

// authService is the slice of *auth.Service handlers actually use. Declared
// here so tests can inject a fake.
type authService interface {
	Signup(ctx context.Context, name, email, password string) (auth.User, string, error)
	Login(ctx context.Context, email, password string) (auth.User, string, error)
	TTL() time.Duration
}

type AuthHandlers struct {
	svc          authService
	secureCookie bool
}

func NewAuthHandlers(svc authService, secureCookie bool) *AuthHandlers {
	return &AuthHandlers{svc: svc, secureCookie: secureCookie}
}

type signupRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// userResponse is the shape returned by signup, login, and /me. Defined in
// this file because signup/login are its primary writers; me.go reads it too.
type userResponse struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

func (h *AuthHandlers) Signup(w http.ResponseWriter, r *http.Request) {
	var req signupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	user, token, err := h.svc.Signup(r.Context(), req.Name, req.Email, req.Password)
	if err != nil {
		switch {
		case errors.Is(err, auth.ErrEmailExists):
			http.Error(w, "email already exists", http.StatusConflict)
		default:
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		return
	}

	http.SetCookie(w, auth.SessionCookie(token, h.svc.TTL(), h.secureCookie))
	writeJSON(w, http.StatusCreated, userResponse{ID: user.ID.String(), Name: user.Name, Email: user.Email})
}

func (h *AuthHandlers) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	user, token, err := h.svc.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, auth.ErrInvalidCredentials) {
			http.Error(w, "invalid credentials", http.StatusUnauthorized)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, auth.SessionCookie(token, h.svc.TTL(), h.secureCookie))
	writeJSON(w, http.StatusOK, userResponse{ID: user.ID.String(), Name: user.Name, Email: user.Email})
}

func (h *AuthHandlers) Logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, auth.ClearSessionCookie(h.secureCookie))
	w.WriteHeader(http.StatusNoContent)
}

// writeJSON is a tiny helper shared by handler files in this package.
func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

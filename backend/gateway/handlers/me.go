package handlers

import (
	"context"
	"net/http"

	"github.com/google/uuid"

	"github.com/1Kyryll/ecommerce-demo/backend/gateway/auth"
	"github.com/1Kyryll/ecommerce-demo/backend/internal/middleware"
)

// meService is the slice of *auth.Service the /me handler uses.
type meService interface {
	GetUser(ctx context.Context, id uuid.UUID) (auth.User, error)
}

type MeHandlers struct {
	svc meService
}

func NewMeHandlers(svc meService) *MeHandlers {
	return &MeHandlers{svc: svc}
}

func (h *MeHandlers) Get(w http.ResponseWriter, r *http.Request) {
	uid, ok := middleware.GetUserID(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	user, err := h.svc.GetUser(r.Context(), uid)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	writeJSON(w, http.StatusOK, userResponse{ID: user.ID.String(), Name: user.Name, Email: user.Email})
}

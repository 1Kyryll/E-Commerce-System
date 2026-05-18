package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"

	"github.com/1Kyryll/ecommerce-demo/backend/gateway/auth"
	"github.com/1Kyryll/ecommerce-demo/backend/internal/middleware"
)

type fakeMeSvc struct {
	user auth.User
	err  error
}

func (f *fakeMeSvc) GetUser(_ context.Context, _ uuid.UUID) (auth.User, error) {
	return f.user, f.err
}

func TestMe_OK(t *testing.T) {
	uid := uuid.New()
	h := NewMeHandlers(&fakeMeSvc{user: auth.User{ID: uid, Name: "Ada", Email: "ada@example.com"}})

	req := httptest.NewRequest(http.MethodGet, "/me", nil)
	ctx := context.WithValue(req.Context(), middleware.UserIDContextKey(), uid)
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()
	h.Get(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
	var body userResponse
	_ = json.NewDecoder(rec.Body).Decode(&body)
	if body.Email != "ada@example.com" {
		t.Errorf("email = %q", body.Email)
	}
}

func TestMe_NoCtx_Unauthorized(t *testing.T) {
	h := NewMeHandlers(&fakeMeSvc{})
	req := httptest.NewRequest(http.MethodGet, "/me", nil)
	rec := httptest.NewRecorder()
	h.Get(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", rec.Code)
	}
}

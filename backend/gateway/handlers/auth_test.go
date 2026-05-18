package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/1Kyryll/ecommerce-demo/backend/gateway/auth"
)

type fakeAuthSvc struct {
	signup func(ctx context.Context, name, email, password string) (auth.User, string, error)
	login  func(ctx context.Context, email, password string) (auth.User, string, error)
	ttl    time.Duration
}

func (f *fakeAuthSvc) Signup(ctx context.Context, name, email, password string) (auth.User, string, error) {
	return f.signup(ctx, name, email, password)
}
func (f *fakeAuthSvc) Login(ctx context.Context, email, password string) (auth.User, string, error) {
	return f.login(ctx, email, password)
}
func (f *fakeAuthSvc) TTL() time.Duration { return f.ttl }

func TestSignup_Created(t *testing.T) {
	want := auth.User{ID: uuid.New(), Name: "Ada", Email: "ada@example.com"}
	h := NewAuthHandlers(&fakeAuthSvc{
		signup: func(_ context.Context, name, email, password string) (auth.User, string, error) {
			return want, "tok", nil
		},
		ttl: time.Hour,
	}, false)

	body, _ := json.Marshal(map[string]string{"name": "Ada", "email": "ada@example.com", "password": "hunter22"})
	req := httptest.NewRequest(http.MethodPost, "/auth/signup", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	h.Signup(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("status = %d, want 201", rec.Code)
	}
	cookies := rec.Result().Cookies()
	if len(cookies) != 1 || cookies[0].Name != "session" || cookies[0].Value != "tok" {
		t.Errorf("session cookie missing or wrong: %+v", cookies)
	}
	if !cookies[0].HttpOnly {
		t.Error("cookie not HttpOnly")
	}
}

func TestSignup_EmailExists_Conflict(t *testing.T) {
	h := NewAuthHandlers(&fakeAuthSvc{
		signup: func(context.Context, string, string, string) (auth.User, string, error) {
			return auth.User{}, "", auth.ErrEmailExists
		},
		ttl: time.Hour,
	}, false)

	body, _ := json.Marshal(map[string]string{"name": "Ada", "email": "ada@example.com", "password": "hunter22"})
	req := httptest.NewRequest(http.MethodPost, "/auth/signup", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	h.Signup(rec, req)

	if rec.Code != http.StatusConflict {
		t.Errorf("status = %d, want 409", rec.Code)
	}
}

func TestSignup_BadJSON_BadRequest(t *testing.T) {
	h := NewAuthHandlers(&fakeAuthSvc{ttl: time.Hour}, false)
	req := httptest.NewRequest(http.MethodPost, "/auth/signup", bytes.NewReader([]byte("{not json")))
	rec := httptest.NewRecorder()
	h.Signup(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rec.Code)
	}
}

func TestLogin_OK(t *testing.T) {
	want := auth.User{ID: uuid.New(), Name: "Ada", Email: "ada@example.com"}
	h := NewAuthHandlers(&fakeAuthSvc{
		login: func(context.Context, string, string) (auth.User, string, error) {
			return want, "tok", nil
		},
		ttl: time.Hour,
	}, false)

	body, _ := json.Marshal(map[string]string{"email": "ada@example.com", "password": "hunter22"})
	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	h.Login(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
	cookies := rec.Result().Cookies()
	if len(cookies) != 1 || cookies[0].Value != "tok" {
		t.Errorf("session cookie wrong: %+v", cookies)
	}
}

func TestLogin_BadCredentials_Unauthorized(t *testing.T) {
	h := NewAuthHandlers(&fakeAuthSvc{
		login: func(context.Context, string, string) (auth.User, string, error) {
			return auth.User{}, "", auth.ErrInvalidCredentials
		},
		ttl: time.Hour,
	}, false)

	body, _ := json.Marshal(map[string]string{"email": "ada@example.com", "password": "wrong"})
	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	h.Login(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", rec.Code)
	}
}

func TestLogout_ClearsCookie(t *testing.T) {
	h := NewAuthHandlers(&fakeAuthSvc{ttl: time.Hour}, false)
	req := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
	rec := httptest.NewRecorder()
	h.Logout(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("status = %d, want 204", rec.Code)
	}
	cookies := rec.Result().Cookies()
	if len(cookies) != 1 || cookies[0].Name != "session" || cookies[0].MaxAge >= 0 {
		t.Errorf("expected cleared session cookie, got %+v", cookies)
	}
}

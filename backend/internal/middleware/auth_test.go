package middleware

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
)

type fakeVerifier struct {
	uid uuid.UUID
	err error
}

func (f fakeVerifier) VerifySession(token string) (uuid.UUID, error) {
	return f.uid, f.err
}

func TestAuth_NoCookie_Unauthorized(t *testing.T) {
	handler := Auth(fakeVerifier{})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("inner handler should not run")
	}))
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", rec.Code)
	}
}

func TestAuth_BadToken_Unauthorized(t *testing.T) {
	handler := Auth(fakeVerifier{err: errors.New("bad")})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("inner handler should not run")
	}))
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.AddCookie(&http.Cookie{Name: SessionCookieName, Value: "bad"})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", rec.Code)
	}
}

func TestAuth_GoodToken_PassesThroughWithCtx(t *testing.T) {
	want := uuid.New()
	var seen uuid.UUID
	handler := Auth(fakeVerifier{uid: want})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got, ok := GetUserID(r.Context())
		if !ok {
			t.Error("user id missing from context")
		}
		seen = got
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.AddCookie(&http.Cookie{Name: SessionCookieName, Value: "good"})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
	if seen != want {
		t.Errorf("seen uid = %s, want %s", seen, want)
	}
}

func TestGetUserID_AbsentReturnsFalse(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	if _, ok := GetUserID(req.Context()); ok {
		t.Error("expected ok=false when no uid in ctx")
	}
}

package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRequestID_GeneratesWhenMissing(t *testing.T) {
	var seenInHandler string
	handler := RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seenInHandler = GetRequestID(r.Context())
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if seenInHandler == "" {
		t.Error("expected non-empty request id in handler context")
	}
	if got := rec.Header().Get(RequestIDHeader); got != seenInHandler {
		t.Errorf("response header %s = %q, want %q", RequestIDHeader, got, seenInHandler)
	}
}

func TestRequestID_PropagatesExisting(t *testing.T) {
	const want = "client-supplied-id"
	var seenInHandler string
	handler := RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seenInHandler = GetRequestID(r.Context())
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(RequestIDHeader, want)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if seenInHandler != want {
		t.Errorf("handler ctx id = %q, want %q", seenInHandler, want)
	}
	if got := rec.Header().Get(RequestIDHeader); got != want {
		t.Errorf("response header %s = %q, want %q", RequestIDHeader, got, want)
	}
}

func TestGetRequestID_EmptyWhenAbsent(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	if got := GetRequestID(req.Context()); got != "" {
		t.Errorf("expected empty string for missing id, got %q", got)
	}
}

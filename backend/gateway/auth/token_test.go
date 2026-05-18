package auth

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

var testSecret = []byte("a-test-secret-at-least-32-chars-long")

func TestIssueAndVerify_RoundTrip(t *testing.T) {
	uid := uuid.New()
	token, err := IssueToken(testSecret, uid, time.Hour)
	if err != nil {
		t.Fatalf("IssueToken: %v", err)
	}
	got, err := VerifyToken(testSecret, token)
	if err != nil {
		t.Fatalf("VerifyToken: %v", err)
	}
	if got != uid {
		t.Errorf("uid = %s, want %s", got, uid)
	}
}

func TestVerifyToken_BadSecret(t *testing.T) {
	token, _ := IssueToken(testSecret, uuid.New(), time.Hour)
	_, err := VerifyToken([]byte("different-secret-also-32-chars-min!!"), token)
	if !errors.Is(err, ErrInvalidToken) {
		t.Errorf("err = %v, want ErrInvalidToken", err)
	}
}

func TestVerifyToken_Expired(t *testing.T) {
	token, _ := IssueToken(testSecret, uuid.New(), -time.Hour)
	_, err := VerifyToken(testSecret, token)
	if !errors.Is(err, ErrInvalidToken) {
		t.Errorf("err = %v, want ErrInvalidToken", err)
	}
}

func TestVerifyToken_GarbageString(t *testing.T) {
	_, err := VerifyToken(testSecret, "definitely-not-a-jwt")
	if !errors.Is(err, ErrInvalidToken) {
		t.Errorf("err = %v, want ErrInvalidToken", err)
	}
}

package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

type fakeRepo struct {
	byEmail   map[string]User
	byID      map[uuid.UUID]User
	insertErr error
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{byEmail: map[string]User{}, byID: map[uuid.UUID]User{}}
}

func (f *fakeRepo) Insert(_ context.Context, u User) (User, error) {
	if f.insertErr != nil {
		return User{}, f.insertErr
	}
	if _, ok := f.byEmail[u.Email]; ok {
		return User{}, ErrEmailExists
	}
	u.CreatedAt = time.Now()
	f.byEmail[u.Email] = u
	f.byID[u.ID] = u
	return u, nil
}

func (f *fakeRepo) GetByEmail(_ context.Context, email string) (User, error) {
	u, ok := f.byEmail[email]
	if !ok {
		return User{}, ErrInvalidCredentials
	}
	return u, nil
}

func (f *fakeRepo) GetByID(_ context.Context, id uuid.UUID) (User, error) {
	u, ok := f.byID[id]
	if !ok {
		return User{}, ErrInvalidCredentials
	}
	return u, nil
}

func newTestService(repo userRepo) *Service {
	return NewService(repo, testSecret, time.Hour)
}

func TestService_SignupHappyPath(t *testing.T) {
	svc := newTestService(newFakeRepo())
	user, token, err := svc.Signup(context.Background(), "Ada", "ada@example.com", "hunter22")
	if err != nil {
		t.Fatalf("Signup: %v", err)
	}
	if user.Email != "ada@example.com" {
		t.Errorf("email = %q", user.Email)
	}
	if user.PasswordHash == "hunter22" {
		t.Error("password stored in plaintext")
	}
	if token == "" {
		t.Error("empty token")
	}
	uid, err := VerifyToken(testSecret, token)
	if err != nil {
		t.Fatalf("VerifyToken: %v", err)
	}
	if uid != user.ID {
		t.Errorf("token subject = %s, want %s", uid, user.ID)
	}
}

func TestService_SignupNormalizesEmail(t *testing.T) {
	repo := newFakeRepo()
	svc := newTestService(repo)
	_, _, err := svc.Signup(context.Background(), "Ada", "  ADA@Example.COM  ", "hunter22")
	if err != nil {
		t.Fatalf("Signup: %v", err)
	}
	if _, ok := repo.byEmail["ada@example.com"]; !ok {
		t.Error("email was not normalized to lowercase + trimmed")
	}
}

func TestService_SignupDuplicateEmail(t *testing.T) {
	svc := newTestService(newFakeRepo())
	_, _, err := svc.Signup(context.Background(), "Ada", "ada@example.com", "hunter22")
	if err != nil {
		t.Fatal(err)
	}
	_, _, err = svc.Signup(context.Background(), "Other", "ada@example.com", "hunter22")
	if !errors.Is(err, ErrEmailExists) {
		t.Errorf("err = %v, want ErrEmailExists", err)
	}
}

func TestService_LoginHappyPath(t *testing.T) {
	svc := newTestService(newFakeRepo())
	signupUser, _, err := svc.Signup(context.Background(), "Ada", "ada@example.com", "hunter22")
	if err != nil {
		t.Fatal(err)
	}
	loginUser, token, err := svc.Login(context.Background(), "ada@example.com", "hunter22")
	if err != nil {
		t.Fatalf("Login: %v", err)
	}
	if loginUser.ID != signupUser.ID {
		t.Errorf("id mismatch")
	}
	if token == "" {
		t.Error("empty token")
	}
}

func TestService_LoginWrongPassword(t *testing.T) {
	svc := newTestService(newFakeRepo())
	_, _, _ = svc.Signup(context.Background(), "Ada", "ada@example.com", "hunter22")
	_, _, err := svc.Login(context.Background(), "ada@example.com", "wrong-password")
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Errorf("err = %v, want ErrInvalidCredentials", err)
	}
}

func TestService_LoginUnknownEmail(t *testing.T) {
	svc := newTestService(newFakeRepo())
	_, _, err := svc.Login(context.Background(), "noone@example.com", "anything")
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Errorf("err = %v, want ErrInvalidCredentials", err)
	}
}

func TestService_VerifySessionRoundTrip(t *testing.T) {
	svc := newTestService(newFakeRepo())
	user, token, _ := svc.Signup(context.Background(), "Ada", "ada@example.com", "hunter22")
	uid, err := svc.VerifySession(token)
	if err != nil {
		t.Fatalf("VerifySession: %v", err)
	}
	if uid != user.ID {
		t.Errorf("uid = %s, want %s", uid, user.ID)
	}
}

func TestService_GetUserHappyPath(t *testing.T) {
	svc := newTestService(newFakeRepo())
	user, _, _ := svc.Signup(context.Background(), "Ada", "ada@example.com", "hunter22")
	got, err := svc.GetUser(context.Background(), user.ID)
	if err != nil {
		t.Fatalf("GetUser: %v", err)
	}
	if got.Email != "ada@example.com" {
		t.Errorf("email = %q", got.Email)
	}
}

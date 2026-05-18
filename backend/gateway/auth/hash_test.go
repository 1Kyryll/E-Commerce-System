package auth

import "testing"

func TestHashAndVerify_RoundTrip(t *testing.T) {
	hash, err := HashPassword("hunter2")
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}
	if hash == "hunter2" {
		t.Fatal("hash equals plaintext")
	}
	if err := VerifyPassword(hash, "hunter2"); err != nil {
		t.Errorf("VerifyPassword same: %v", err)
	}
}

func TestVerifyPassword_WrongPassword(t *testing.T) {
	hash, _ := HashPassword("hunter2")
	if err := VerifyPassword(hash, "hunter3"); err == nil {
		t.Error("expected error for wrong password")
	}
}

func TestHashPassword_DistinctSalts(t *testing.T) {
	a, _ := HashPassword("same")
	b, _ := HashPassword("same")
	if a == b {
		t.Error("hashes are identical — bcrypt should salt each one")
	}
}

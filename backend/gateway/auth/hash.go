package auth

import "golang.org/x/crypto/bcrypt"

// HashPassword runs bcrypt at the default cost. The returned string is safe
// to store in the database; VerifyPassword consumes the same format.
func HashPassword(plaintext string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(plaintext), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// VerifyPassword returns nil iff plaintext matches the bcrypt-encoded hash.
func VerifyPassword(hash, plaintext string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plaintext))
}

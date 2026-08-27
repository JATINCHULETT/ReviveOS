package auth

import (
	"testing"
	"time"
)

func TestPasswordHashing(t *testing.T) {
	password := "SecretPass123"
	hash := HashPassword(password)

	if !CheckPasswordHash(password, hash) {
		t.Errorf("expected CheckPasswordHash to return true for correct password")
	}

	if CheckPasswordHash("WrongPass", hash) {
		t.Errorf("expected CheckPasswordHash to return false for incorrect password")
	}
}

func TestJWTTokenGenerationAndVerification(t *testing.T) {
	secret := "test_secret_key_12345"
	userID := "usr_1111-2222"
	merchantID := "merch_3333-4444"
	email := "admin@reviveos.io"
	role := "ADMIN"

	token, err := GenerateToken(userID, merchantID, email, role, secret, 2*time.Hour)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	claims, err := VerifyToken(token, secret)
	if err != nil {
		t.Fatalf("failed to verify token: %v", err)
	}

	if claims.UserID != userID || claims.Email != email || claims.Role != role || claims.MerchantID != merchantID {
		t.Errorf("claims mismatch: %+v", claims)
	}

	// Tampered token test
	tamperedToken := token + "tampered"
	_, err = VerifyToken(tamperedToken, secret)
	if err == nil {
		t.Errorf("expected tampered token to fail verification")
	}

	// Wrong secret test
	_, err = VerifyToken(token, "wrong_secret")
	if err == nil {
		t.Errorf("expected wrong secret to fail verification")
	}
}

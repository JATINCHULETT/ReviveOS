package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

// Claims contains JWT payload claims.
type Claims struct {
	UserID     string `json:"user_id"`
	MerchantID string `json:"merchant_id,omitempty"`
	Email      string `json:"email"`
	Role       string `json:"role"` // ADMIN or MERCHANT
	IssuedAt   int64  `json:"iat"`
	ExpiresAt  int64  `json:"exp"`
}

// HashPassword hashes a password using SHA-256 with a secure random salt.
func HashPassword(password string) string {
	salt := make([]byte, 16)
	_, _ = rand.Read(salt)
	saltHex := hex.EncodeToString(salt)

	mac := hmac.New(sha256.New, []byte(saltHex))
	mac.Write([]byte(password))
	hashHex := hex.EncodeToString(mac.Sum(nil))

	return fmt.Sprintf("%s:%s", saltHex, hashHex)
}

// CheckPasswordHash verifies a password against a salt:hash string.
func CheckPasswordHash(password, storedHash string) bool {
	parts := strings.Split(storedHash, ":")
	if len(parts) != 2 {
		// Fallback check for simple match or plain hashing
		mac := sha256.Sum256([]byte(password))
		return hex.EncodeToString(mac[:]) == storedHash
	}

	saltHex := parts[0]
	expectedHash := parts[1]

	mac := hmac.New(sha256.New, []byte(saltHex))
	mac.Write([]byte(password))
	actualHash := hex.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(actualHash), []byte(expectedHash))
}

// GenerateToken creates an HMAC-SHA256 signed JWT token.
func GenerateToken(userID, merchantID, email, role, secret string, duration time.Duration) (string, error) {
	if secret == "" {
		secret = "reviveos_jwt_default_secret_key_2026"
	}

	header := map[string]string{
		"alg": "HS256",
		"typ": "JWT",
	}
	headerJSON, err := json.Marshal(header)
	if err != nil {
		return "", err
	}
	headerB64 := base64.RawURLEncoding.EncodeToString(headerJSON)

	now := time.Now().UTC()
	claims := Claims{
		UserID:     userID,
		MerchantID: merchantID,
		Email:      email,
		Role:       role,
		IssuedAt:   now.Unix(),
		ExpiresAt:  now.Add(duration).Unix(),
	}
	claimsJSON, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}
	claimsB64 := base64.RawURLEncoding.EncodeToString(claimsJSON)

	unsignedToken := fmt.Sprintf("%s.%s", headerB64, claimsB64)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(unsignedToken))
	sigB64 := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))

	return fmt.Sprintf("%s.%s", unsignedToken, sigB64), nil
}

// VerifyToken verifies the JWT signature and expiration, returning the Claims.
func VerifyToken(tokenString, secret string) (*Claims, error) {
	if secret == "" {
		secret = "reviveos_jwt_default_secret_key_2026"
	}

	parts := strings.Split(tokenString, ".")
	if len(parts) != 3 {
		return nil, errors.New("invalid token format")
	}

	headerB64, claimsB64, sigB64 := parts[0], parts[1], parts[2]
	unsignedToken := fmt.Sprintf("%s.%s", headerB64, claimsB64)

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(unsignedToken))
	expectedSig := mac.Sum(nil)
	actualSig, err := base64.RawURLEncoding.DecodeString(sigB64)
	if err != nil || !hmac.Equal(expectedSig, actualSig) {
		return nil, errors.New("invalid token signature")
	}

	claimsJSON, err := base64.RawURLEncoding.DecodeString(claimsB64)
	if err != nil {
		return nil, errors.New("invalid token payload")
	}

	var claims Claims
	if err := json.Unmarshal(claimsJSON, &claims); err != nil {
		return nil, errors.New("malformed claims")
	}

	if claims.ExpiresAt < time.Now().UTC().Unix() {
		return nil, errors.New("token expired")
	}

	return &claims, nil
}

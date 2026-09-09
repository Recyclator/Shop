package utils

import (
	"testing"
	"time"
)

func TestRevokeAndIsRevoked(t *testing.T) {
	jti := "test-jti-uuid-123456"

	// Before revocation
	if IsRevoked(jti) {
		t.Fatal("el token no revocado no debería estar revocado")
	}

	// Revoke for 5 minutes
	RevokeToken(jti, time.Now().Add(5*time.Minute))

	// After revocation
	if !IsRevoked(jti) {
		t.Fatal("el token revocado debería estar marcado como revocado")
	}

	// Empty JTI should not report as revoked
	if IsRevoked("") {
		t.Fatal("un JTI vacío no debería reportarse como revocado")
	}
}

func TestRevokeExpiredEntry(t *testing.T) {
	jti := "test-jti-expired-999"

	// Revoke with expiration already in the past
	RevokeToken(jti, time.Now().Add(-1*time.Second))

	// Should not be active in blacklist
	if IsRevoked(jti) {
		t.Fatal("un token con expiración en el pasado no debería ser revocado")
	}
}

func TestRevokeTokenString(t *testing.T) {
	SetJWTSecret("test-secret-key-32-characters-long-12345")

	tokenString, err := GenerateJWT(123, "revoke@nexora.com", "vendedor", 10*time.Minute)
	if err != nil {
		t.Fatalf("error generando token: %v", err)
	}

	claims, err := ValidateJWT(tokenString)
	if err != nil {
		t.Fatalf("token recién generado debe ser válido: %v", err)
	}

	if IsRevoked(claims.ID) {
		t.Fatal("token recién generado no debe estar en lista negra")
	}

	// Revoke using full token string
	RevokeTokenString(tokenString)

	if !IsRevoked(claims.ID) {
		t.Fatal("token revocado mediante RevokeTokenString debería figurar como revocado")
	}
}

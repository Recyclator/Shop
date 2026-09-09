package utils

import (
	"testing"
	"time"
)

func init() {
	SetJWTSecret("test-secret-key-32-characters-long-12345")
}

func TestGenerateAndValidateJWT(t *testing.T) {
	userID := uint(42)
	email := "user@nexora.com"
	role := "admin"
	duration := 15 * time.Minute

	tokenString, err := GenerateJWT(userID, email, role, duration)
	if err != nil {
		t.Fatalf("error generando token: %v", err)
	}

	if tokenString == "" {
		t.Fatal("el token generado no debe estar vacío")
	}

	claims, err := ValidateJWT(tokenString)
	if err != nil {
		t.Fatalf("error validando token: %v", err)
	}

	if claims.UserID != userID {
		t.Fatalf("esperaba UserID %d, pero obtuvo %d", userID, claims.UserID)
	}

	if claims.Email != email {
		t.Fatalf("esperaba Email %s, pero obtuvo %s", email, claims.Email)
	}

	if claims.Role != role {
		t.Fatalf("esperaba Role %s, pero obtuvo %s", role, claims.Role)
	}

	if claims.ID == "" {
		t.Fatal("esperaba que el token tuviera un JTI generado con UUID")
	}
}

func TestExpiredJWT(t *testing.T) {
	// Generate an already expired token (-1 second)
	tokenString, err := GenerateJWT(10, "exp@nexora.com", "user", -1*time.Second)
	if err != nil {
		t.Fatalf("error generando token expirado: %v", err)
	}

	_, err = ValidateJWT(tokenString)
	if err == nil {
		t.Fatal("esperaba que un token expirado fuera rechazado por ValidateJWT")
	}
}

func TestInvalidJWT(t *testing.T) {
	invalidTokens := []string{
		"",
		"not.a.jwt",
		"header.payload",
		"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.invalidpayload.invalidsig",
	}

	for _, tok := range invalidTokens {
		_, err := ValidateJWT(tok)
		if err == nil {
			t.Fatalf("esperaba error para token inválido '%s', pero pasó", tok)
		}
	}
}

func TestGenerateTokenPair(t *testing.T) {
	userID := uint(99)
	email := "pair@nexora.com"
	role := "cajero"

	pair, err := GenerateTokenPair(userID, email, role, 15*time.Minute, 7*24*time.Hour)
	if err != nil {
		t.Fatalf("error generando token pair: %v", err)
	}

	if pair == nil || pair.AccessToken == "" || pair.RefreshToken == "" {
		t.Fatal("el par de tokens no debe ser nulo ni contener strings vacíos")
	}

	accessClaims, err := ValidateJWT(pair.AccessToken)
	if err != nil {
		t.Fatalf("access token inválido: %v", err)
	}
	if accessClaims.UserID != userID {
		t.Fatalf("esperaba UserID %d, obtuvo %d", userID, accessClaims.UserID)
	}

	refreshClaims, err := ValidateJWT(pair.RefreshToken)
	if err != nil {
		t.Fatalf("refresh token inválido: %v", err)
	}
	if refreshClaims.UserID != userID {
		t.Fatalf("esperaba UserID %d, obtuvo %d", userID, refreshClaims.UserID)
	}
}

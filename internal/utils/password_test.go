package utils

import (
	"testing"
)

func TestHashAndCheckPassword(t *testing.T) {
	rawPassword := "SecurePass123!#"
	hash, err := HashPassword(rawPassword)
	if err != nil {
		t.Fatalf("error hashing password: %v", err)
	}

	if hash == "" || hash == rawPassword {
		t.Fatal("el hash no debe estar vacío ni ser igual a la contraseña en texto plano")
	}

	if !CheckPassword(rawPassword, hash) {
		t.Fatal("la contraseña correcta debería coincidir con el hash")
	}

	if CheckPassword("WrongPassword123!#", hash) {
		t.Fatal("una contraseña incorrecta no debería coincidir con el hash")
	}
}

func TestValidatePassword(t *testing.T) {
	tests := []struct {
		name        string
		password    string
		expectError bool
	}{
		{"Too short", "Short1!a", true},
		{"No uppercase", "lowercase123!#$", true},
		{"No lowercase", "UPPERCASE123!#$", true},
		{"No number", "NoNumbersHere!#$", true},
		{"No special char", "NoSpecialChars123", true},
		{"Valid strong password", "Str0ng#P@ssword2026", false},
		{"Valid minimum 12 chars", "Abcdefgh123!#", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidatePassword(tc.password)
			if tc.expectError && err == nil {
				t.Fatalf("esperaba error para '%s', pero pasó", tc.password)
			}
			if !tc.expectError && err != nil {
				t.Fatalf("esperaba que '%s' fuera válida, pero falló: %v", tc.password, err)
			}
		})
	}
}

func TestGenerateRandomPassword(t *testing.T) {
	lengths := []int{12, 16, 24, 32}
	for _, l := range lengths {
		p := GenerateRandomPassword(l)
		if len(p) != l {
			t.Fatalf("esperaba longitud %d, pero obtuvo %d", l, len(p))
		}
	}
}

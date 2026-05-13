package utils

import (
	"errors"
	"unicode"

	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func ValidatePassword(password string) error {
	if len(password) < 12 {
		return errors.New("la contraseña debe tener al menos 12 caracteres")
	}

	var hasUpper, hasLower, hasDigit, hasSpecial bool
	for _, r := range password {
		switch {
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsDigit(r):
			hasDigit = true
		case unicode.IsPunct(r) || unicode.IsSymbol(r):
			hasSpecial = true
		}
	}

	if !hasUpper {
		return errors.New("la contraseña debe contener al menos una letra mayúscula")
	}
	if !hasLower {
		return errors.New("la contraseña debe contener al menos una letra minúscula")
	}
	if !hasDigit {
		return errors.New("la contraseña debe contener al menos un número")
	}
	if !hasSpecial {
		return errors.New("la contraseña debe contener al menos un símbolo especial")
	}

	return nil
}

func GenerateRandomPassword(length int) string {
	// Implementar si se necesita generación automática de contraseñas
	return ""
}

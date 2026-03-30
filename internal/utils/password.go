package utils

import (
	"errors"

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
	if len(password) < 6 {
		return errors.New("la contraseña debe tener al menos 6 caracteres")
	}
	return nil
}

func GenerateRandomPassword(length int) string {
	// Implementar si se necesita generación automática de contraseñas
	return ""
}

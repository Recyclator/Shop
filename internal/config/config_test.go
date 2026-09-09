package config

import (
	"testing"
)

func TestConfigValidateDevelopment(t *testing.T) {
	cfg := &Config{
		ServerPort: "3000",
		Env:        "development",
		DBHost:     "localhost",
		DBPort:     "5432",
		DBUser:     "postgres",
		DBPassword: "password",
		DBName:     "nexora",
		JWTSecret:  "change-this-secret-in-production",
	}

	if err := cfg.Validate(); err != nil {
		t.Fatalf("esperaba que la configuración de desarrollo fuera válida, pero falló: %v", err)
	}
}

func TestConfigValidateProductionInsecureSecret(t *testing.T) {
	cfg := &Config{
		ServerPort: "3000",
		Env:        "production",
		DBHost:     "localhost",
		DBPort:     "5432",
		DBUser:     "postgres",
		DBPassword: "strong_production_password_123!",
		DBName:     "nexora",
		JWTSecret:  "change-this-secret-in-production",
	}

	if err := cfg.Validate(); err == nil {
		t.Fatal("esperaba que la configuración de producción fallara con JWT_SECRET por defecto, pero pasó")
	}
}

func TestConfigValidateProductionShortSecret(t *testing.T) {
	cfg := &Config{
		ServerPort: "3000",
		Env:        "production",
		DBHost:     "localhost",
		DBPort:     "5432",
		DBUser:     "postgres",
		DBPassword: "strong_production_password_123!",
		DBName:     "nexora",
		JWTSecret:  "too_short_secret", // < 32 chars
	}

	if err := cfg.Validate(); err == nil {
		t.Fatal("esperaba que la configuración de producción fallara con JWT_SECRET corto (<32 chars), pero pasó")
	}
}

func TestConfigValidateProductionValid(t *testing.T) {
	cfg := &Config{
		ServerPort: "3000",
		Env:        "production",
		DBHost:     "localhost",
		DBPort:     "5432",
		DBUser:     "postgres",
		DBPassword: "strong_production_password_123!",
		DBName:     "nexora",
		JWTSecret:  "a_very_secure_random_production_secret_key_with_sufficient_entropy_12345",
	}

	if err := cfg.Validate(); err != nil {
		t.Fatalf("esperaba que la configuración de producción válida pasara, pero falló: %v", err)
	}
}

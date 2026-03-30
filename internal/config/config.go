package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	// Server
	ServerPort string
	Env        string

	// Database
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string

	// JWT
	JWTSecret      string
	JWTExpire      time.Duration
	JWTExpireHours int

	// Casbin
	CasbinModelPath  string
	CasbinPolicyPath string

	// Application
	AppURL      string
	FrontendURL string

	// QR
	QRBaseURL string

	// Company (for DIAN)
	CompanyName            string
	CompanyNIT             string
	CompanyDV              string
	CompanyAddress         string
	CompanyCity            string
	CompanyCityCode        string
	CompanyDepartment      string
	CompanyPhone           string
	CompanyEmail           string
	CompanyRegimen         string
	CompanyResponsabilidad string

	// DIAN
	DIANAPIURL           string
	DIANAPIKey           string
	DIANSoftwareID       string
	DIANSoftwarePassword string
	DIANCertificatePath  string
	DIANTestMode         bool
}

func Load() *Config {
	jwtExpireHours := 168 // Default 7 days
	if expire := os.Getenv("JWT_EXPIRE_HOURS"); expire != "" {
		if hours, err := strconv.Atoi(expire); err == nil {
			jwtExpireHours = hours
		}
	}

	return &Config{
		ServerPort:             getEnv("SERVER_PORT", "3000"),
		Env:                    getEnv("ENV", "development"),
		DBHost:                 getEnv("DB_HOST", "localhost"),
		DBPort:                 getEnv("DB_PORT", "5432"),
		DBUser:                 getEnv("DB_USER", "postgres"),
		DBPassword:             getEnv("DB_PASSWORD", "password"),
		DBName:                 getEnv("DB_NAME", "nexora"),
		DBSSLMode:              getEnv("DB_SSL_MODE", "disable"),
		JWTSecret:              getEnv("JWT_SECRET", "change-this-secret-in-production"),
		JWTExpireHours:         jwtExpireHours,
		JWTExpire:              time.Duration(jwtExpireHours) * time.Hour,
		CasbinModelPath:        getEnv("CASBIN_MODEL_PATH", "pkg/casbin/model.conf"),
		CasbinPolicyPath:       getEnv("CASBIN_POLICY_PATH", "pkg/casbin/policy.csv"),
		AppURL:                 getEnv("APP_URL", "http://localhost:3000"),
		FrontendURL:            getEnv("FRONTEND_URL", "http://localhost:5173"),
		QRBaseURL:              getEnv("QR_BASE_URL", "https://nexora.co/qr"),
		CompanyName:            getEnv("COMPANY_NAME", "Nexora SAS"),
		CompanyNIT:             getEnv("COMPANY_NIT", "1234567890"),
		CompanyDV:              getEnv("COMPANY_DV", "1"),
		CompanyAddress:         getEnv("COMPANY_ADDRESS", "Carrera 10 # 20-30"),
		CompanyCity:            getEnv("COMPANY_CITY", "Bogotá"),
		CompanyCityCode:        getEnv("COMPANY_CITY_CODE", "11001"),
		CompanyDepartment:      getEnv("COMPANY_DEPARTMENT", "Bogotá D.C."),
		CompanyPhone:           getEnv("COMPANY_PHONE", "3001234567"),
		CompanyEmail:           getEnv("COMPANY_EMAIL", "info@nexora.co"),
		CompanyRegimen:         getEnv("COMPANY_REGIMEN", "responsable_iva"),
		CompanyResponsabilidad: getEnv("COMPANY_RESPONSABILIDAD", "R-99-PN"),
		DIANAPIURL:             getEnv("DIAN_API_URL", "https://facturaelectronica.dian.gov.co"),
		DIANAPIKey:             getEnv("DIAN_API_KEY", ""),
		DIANSoftwareID:         getEnv("DIAN_SOFTWARE_ID", ""),
		DIANSoftwarePassword:   getEnv("DIAN_SOFTWARE_PASSWORD", ""),
		DIANCertificatePath:    getEnv("DIAN_CERT_PATH", ""),
		DIANTestMode:           getEnv("DIAN_TEST_MODE", "true") == "true",
	}
}

func (c *Config) GetDSN() string {
	return fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=America/Bogota",
		c.DBHost, c.DBUser, c.DBPassword, c.DBName, c.DBPort, c.DBSSLMode,
	)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

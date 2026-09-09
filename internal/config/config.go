package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	// Server
	ServerPort string
	Env        string

	// Database
	DBHost         string
	DBPort         string
	DBUser         string
	DBPassword     string
	DBName         string
	DBSSLMode      string
	DBMaxOpenConns int
	DBMaxIdleConns int

	// Redis
	RedisHost     string
	RedisPort     string
	RedisPassword string
	RedisDB       int

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

	redisDB := 0
	if dbStr := os.Getenv("REDIS_DB"); dbStr != "" {
		if db, err := strconv.Atoi(dbStr); err == nil {
			redisDB = db
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
		DBMaxOpenConns:         getIntEnv("DB_MAX_OPEN_CONNS", 50),
		DBMaxIdleConns:         getIntEnv("DB_MAX_IDLE_CONNS", 10),
		RedisHost:              getEnv("REDIS_HOST", "localhost"),
		RedisPort:              getEnv("REDIS_PORT", "6379"),
		RedisPassword:          getEnv("REDIS_PASSWORD", ""),
		RedisDB:                redisDB,
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

// Validate valida que la configuración sea coherente y cumpla los requisitos de seguridad
func (c *Config) Validate() error {
	var errs []string

	if c.ServerPort == "" {
		errs = append(errs, "SERVER_PORT no puede estar vacío")
	}
	if c.DBHost == "" {
		errs = append(errs, "DB_HOST no puede estar vacío")
	}
	if c.DBName == "" {
		errs = append(errs, "DB_NAME no puede estar vacío")
	}
	if c.DBUser == "" {
		errs = append(errs, "DB_USER no puede estar vacío")
	}

	isProduction := c.Env == "production" || c.Env == "prod"

	if isProduction {
		insecureSecrets := map[string]bool{
			"change-this-secret-in-production": true,
			"secret":                           true,
			"123456":                           true,
			"password":                         true,
			"":                                 true,
		}

		if insecureSecrets[c.JWTSecret] {
			errs = append(errs, "JWT_SECRET inseguro: en producción debe definirse un secreto criptográfico único (no el valor por defecto)")
		} else if len(c.JWTSecret) < 32 {
			errs = append(errs, fmt.Sprintf("JWT_SECRET débil: longitud actual %d caracteres, se requiere mínimo 32 caracteres", len(c.JWTSecret)))
		}

		insecurePasswords := map[string]bool{
			"password": true,
			"postgres": true,
			"admin":    true,
			"123456":   true,
			"":         true,
		}
		if insecurePasswords[c.DBPassword] {
			errs = append(errs, "DB_PASSWORD inseguro: en producción no se permite usar contraseñas por defecto o vacías")
		}

		if c.DBSSLMode == "disable" && c.DBHost != "localhost" && c.DBHost != "127.0.0.1" && c.DBHost != "db" && c.DBHost != "postgres" {
			errs = append(errs, "DB_SSL_MODE inseguro: conexiones a bases de datos remotas en producción deben usar SSL (require o verify-full)")
		}
	} else {
		if c.JWTSecret == "change-this-secret-in-production" {
			log.Println("⚠️  ADVERTENCIA DE SEGURIDAD: Usando JWT_SECRET por defecto en desarrollo. Cámbielo antes de desplegar a producción.")
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errores de configuración:\n - %s", strings.Join(errs, "\n - "))
	}

	return nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getIntEnv(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil && intVal > 0 {
			return intVal
		}
	}
	return defaultValue
}

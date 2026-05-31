package utils

import (
	"log"
	"os"
)

var jwtSecret string

func GetJWTSecret() string {
	if jwtSecret == "" {
		jwtSecret = os.Getenv("JWT_SECRET")
		if jwtSecret == "" {
			log.Fatal("FATAL: JWT_SECRET no esté configurado. Defina la variable de entorno JWT_SECRET antes de iniciar la aplicación.")
		}
	}
	return jwtSecret
}

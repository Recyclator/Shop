package utils

import "os"

var jwtSecret string

func GetJWTSecret() string {
	if jwtSecret == "" {
		jwtSecret = os.Getenv("JWT_SECRET")
		if jwtSecret == "" {
			jwtSecret = "default-secret-change-in-production"
		}
	}
	return jwtSecret
}

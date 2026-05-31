package middleware

import (
	"github.com/gofiber/fiber/v2"
)

// SecurityHeadersMiddleware agrega headers de seguridad HTTP
func SecurityHeadersMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Prevenir clickjacking
		c.Set("X-Frame-Options", "DENY")

		// Prevenir MIME-sniffing
		c.Set("X-Content-Type-Options", "nosniff")

		// Política de Referrer
		c.Set("Referrer-Policy", "strict-origin-when-cross-origin")

		// Política de permisos de características
		c.Set("Permissions-Policy", "camera=(), microphone=(), geolocation=(), payment=()")

		// X-XSS-Protection (legacy, pero aún defensivo)
		c.Set("X-XSS-Protection", "1; mode=block")

		// Strict-Transport-Security (HSTS) - solo en producción
		// Nota: Descomentar cuando se tenga HTTPS
		// c.Set("Strict-Transport-Security", "max-age=63072000; includeSubDomains; preload")

		// Content-Security-Policy
		csp := "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; font-src 'self'; connect-src 'self'; frame-ancestors 'none'; base-uri 'self'; form-action 'self';"
		c.Set("Content-Security-Policy", csp)

		return c.Next()
	}
}

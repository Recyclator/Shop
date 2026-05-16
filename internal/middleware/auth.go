package middleware

import (
	"strings"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/nexora/backend/internal/database"
	"github.com/nexora/backend/internal/models"
	"github.com/nexora/backend/internal/utils"
)

// UserKey es la clave para guardar el usuario en el contexto de Fiber
const UserKey = "user"

// AuthMiddleware valida el token JWT y agrega el usuario al contexto
func AuthMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")

		if authHeader == "" {
			return utils.ErrorWithCode(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "encabezado de autorización faltante")
		}

		// Extraer token del formato "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return utils.ErrorWithCode(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "formato de token inválido")
		}

		tokenString := parts[1]

		// Validar token
		claims, err := utils.ValidateJWT(tokenString)
		if err != nil {
			return utils.ErrorWithCode(c, fiber.StatusUnauthorized, "TOKEN_INVALID", "token inválido o expirado")
		}

		// Cargar usuario desde la base de datos
		var user models.User
		if result := database.DB.Preload("Roles.Permisos").First(&user, claims.UserID); result.Error != nil {
			return utils.ErrorWithCode(c, fiber.StatusUnauthorized, "USER_NOT_FOUND", "usuario no encontrado")
		}

		// Verificar si el usuario está activo
		if !user.Activo {
			return utils.ErrorWithCode(c, fiber.StatusForbidden, "USER_INACTIVE", "usuario inactivo")
		}

		// Guardar usuario en el contexto
		c.Locals(UserKey, &user)
		return c.Next()
	}
}

// GetUser obtiene el usuario del contexto
func GetUser(c *fiber.Ctx) *models.User {
	user, ok := c.Locals(UserKey).(*models.User)
	if !ok {
		return nil
	}
	return user
}

// OptionalAuthMiddleware intenta autenticar pero permite acceso si no hay token
func OptionalAuthMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")

		if authHeader == "" {
			// Sin autenticación, continuar como guest
			return c.Next()
		}

		// Intentar extraer y validar token
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return c.Next()
		}

		tokenString := parts[1]
		claims, err := utils.ValidateJWT(tokenString)
		if err != nil {
			return c.Next()
		}

		// Cargar usuario
		var user models.User
		if result := database.DB.Preload("Roles.Permisos").First(&user, claims.UserID); result.Error != nil {
			return c.Next()
		}

		if !user.Activo {
			return c.Next()
		}

		c.Locals(UserKey, &user)
		return c.Next()
	}
}

// RateLimiterMiddleware implementa rate limiting básico
func RateLimiterMiddleware(maxRequests int, window time.Duration) fiber.Handler {
	requests := make(map[string][]time.Time)
	var mu sync.Mutex

	return func(c *fiber.Ctx) error {
		ip := c.IP()
		now := time.Now()

		mu.Lock()
		validRequests := requests[ip][:0]
		for _, t := range requests[ip] {
			if now.Sub(t) < window {
				validRequests = append(validRequests, t)
			}
		}
		requests[ip] = validRequests

		if len(requests[ip]) >= maxRequests {
			mu.Unlock()
			return utils.ErrorWithCode(c, fiber.StatusTooManyRequests, "RATE_LIMIT", "demasiadas solicitudes, intenta más tarde")
		}

		requests[ip] = append(requests[ip], now)
		mu.Unlock()
		return c.Next()
	}
}

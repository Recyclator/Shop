package middleware

import (
	"strings"

	"github.com/casbin/casbin/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/nexora/backend/internal/utils"
)

var enforcer *casbin.Enforcer

// InitCasbin inicializa el adaptador de Casbin
func InitCasbin(modelPath, policyPath string) error {
	var err error
	enforcer, err = casbin.NewEnforcer(modelPath, policyPath)
	if err != nil {
		return err
	}

	// Habilitar logger
	enforcer.EnableLog(true)

	return nil
}

// GetEnforcer retorna la instancia de Casbin
func GetEnforcer() *casbin.Enforcer {
	return enforcer
}

// RequirePermission crea un middleware que requiere un permiso específico
func RequirePermission(permission string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		user := GetUser(c)
		if user == nil {
			return utils.ErrorWithCode(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "usuario no autenticado")
		}

		// Obtener el primer rol del usuario
		roles := user.GetUserRoleNames()
		if len(roles) == 0 {
			roles = []string{"guest"} // Rol por defecto
		}

		// Obtener el recurso y acción de la ruta
		obj := c.Params("*") // recurso
		if obj == "" {
			obj = c.Path() // usar la ruta como recurso
		}
		act := c.Method() // método HTTP como acción

		// Verificar con Casbin
		allowed, err := enforcer.Enforce(roles[0], obj, act)
		if err != nil {
			return utils.ErrorWithCode(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", "error al verificar permisos")
		}

		if !allowed {
			return utils.ErrorWithCode(c, fiber.StatusForbidden, "FORBIDDEN", "no tienes permiso para realizar esta acción")
		}

		return c.Next()
	}
}

// RequireRole crea un middleware que requiere un rol específico
func RequireRole(roles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		user := GetUser(c)
		if user == nil {
			return utils.ErrorWithCode(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "usuario no autenticado")
		}

		for _, role := range roles {
			if user.HasRole(role) {
				return c.Next()
			}
		}

		return utils.ErrorWithCode(c, fiber.StatusForbidden, "FORBIDDEN", "no tienes el rol requerido")
	}
}

// RequireLevel crea un middleware que requiere un nivel mínimo de acceso
func RequireLevel(minLevel int) fiber.Handler {
	return func(c *fiber.Ctx) error {
		user := GetUser(c)
		if user == nil {
			return utils.ErrorWithCode(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "usuario no autenticado")
		}

		level := user.GetHighestRoleLevel()
		if level > minLevel {
			return utils.ErrorWithCode(c, fiber.StatusForbidden, "FORBIDDEN", "nivel de acceso insuficiente")
		}

		return c.Next()
	}
}

// RequirePermissionAny crea un middleware que requiere AL MENOS UNO de los permisos especificados
func RequirePermissionAny(permissions ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		user := GetUser(c)
		if user == nil {
			return utils.ErrorWithCode(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "usuario no autenticado")
		}

		for _, perm := range permissions {
			if user.HasPermission(perm) {
				return c.Next()
			}
		}

		return utils.ErrorWithCode(c, fiber.StatusForbidden, "FORBIDDEN", "no tienes ninguno de los permisos requeridos")
	}
}

// RequirePermissionAll crea un middleware que requiere TODOS los permisos especificados
func RequirePermissionAll(permissions ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		user := GetUser(c)
		if user == nil {
			return utils.ErrorWithCode(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "usuario no autenticado")
		}

		for _, perm := range permissions {
			if !user.HasPermission(perm) {
				return utils.ErrorWithCode(c, fiber.StatusForbidden, "FORBIDDEN", "no tienes todos los permisos requeridos")
			}
		}

		return c.Next()
	}
}

// Helper para obtener el objeto de la ruta
func getResourceFromPath(path string) string {
	// Remover prefijos comunes
	path = strings.TrimPrefix(path, "/api/")

	// Obtener el primer segmento como recurso
	parts := strings.Split(path, "/")
	if len(parts) > 0 && parts[0] != "" {
		return parts[0]
	}

	return path
}

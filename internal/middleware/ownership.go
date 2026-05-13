package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/nexora/backend/internal/database"
	"github.com/nexora/backend/internal/models"
	"github.com/nexora/backend/internal/utils"
)

// RequireResourceOwnership verifica que el usuario autenticado sea el dueño del recurso
// o tenga permisos administrativos. Es una protección contra IDOR.
// resourceType: "user", "order", "customer", "layaway", etc.
// idParam: nombre del parámetro de ruta que contiene el ID del recurso (ej: "id")
func RequireResourceOwnership(resourceType string, idParam string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		user := GetUser(c)
		if user == nil {
			return utils.ErrorWithCode(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "usuario no autenticado")
		}

		// Superadmin siempre tiene acceso
		if user.HasRole("superadmin") {
			return c.Next()
		}

		// Admin tiene acceso a todo excepto su propia cuenta
		if user.HasRole("admin") && resourceType != "user" {
			return c.Next()
		}

		id, err := c.ParamsInt(idParam)
		if err != nil {
			return utils.ErrorWithCode(c, fiber.StatusBadRequest, "INVALID_ID", "ID inválido")
		}

		// Verificar propiedad según el tipo de recurso
		isOwner := false
		switch resourceType {
		case "user":
			isOwner = (user.ID == uint(id))
			// Admin puede modificar usuarios no admin
			if !isOwner && user.HasRole("admin") {
				var targetUser models.User
				if result := database.DB.First(&targetUser, id); result.Error == nil {
					// Admin no puede modificar a otro admin o superadmin
					if !targetUser.HasRole("admin") && !targetUser.HasRole("superadmin") {
						return c.Next()
					}
				}
			}
		case "order":
			var order models.Order
			if result := database.DB.First(&order, id); result.Error == nil {
				isOwner = (order.VendedorID != nil && *order.VendedorID == user.ID) || order.ClienteID == user.ID
			}
		case "customer":
			// Los customers pueden ser visto por vendedores y admins
			if user.HasPermission("customers.read") {
				return c.Next()
			}
		case "layaway":
			var layaway models.Layaway
			if result := database.DB.First(&layaway, id); result.Error == nil {
				// Para layaway, solo verificar permisos ya que no hay campo VendedorID
				isOwner = true
			}
		}

		if !isOwner {
			return utils.ErrorWithCode(c, fiber.StatusForbidden, "FORBIDDEN", "no tienes permiso para acceder a este recurso")
		}

		return c.Next()
	}
}

package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/nexora/backend/internal/database"
	"github.com/nexora/backend/internal/middleware"
	"github.com/nexora/backend/internal/models"
	"github.com/nexora/backend/internal/utils"
)

// RefreshPermissions actualiza los permisos de los roles
// POST /api/system/refresh-permissions
func RefreshPermissions(c *fiber.Ctx) error {
	currentUser := middleware.GetUser(c)
	if currentUser == nil {
		return utils.ErrorWithCode(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "no autorizado")
	}

	// Solo superadmin puede hacer esto
	if currentUser.GetHighestRoleLevel() > 1 {
		return utils.ErrorWithCode(c, fiber.StatusForbidden, "FORBIDDEN", "solo el superadmin puede refrescar permisos")
	}

	if err := models.AssignDefaultPermissions(database.DB); err != nil {
		return utils.ErrorWithCode(c, fiber.StatusInternalServerError, "PERMISSION_ERROR", "error al actualizar permisos: "+err.Error())
	}

	return utils.SuccessMessage(c, fiber.StatusOK, "permisos actualizados exitosamente")
}

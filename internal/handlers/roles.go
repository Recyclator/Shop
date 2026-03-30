package handlers

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/nexora/backend/internal/database"
	"github.com/nexora/backend/internal/middleware"
	"github.com/nexora/backend/internal/models"
	"github.com/nexora/backend/internal/utils"
)

// GetRoles obtiene todos los roles
// GET /api/roles
func GetRoles(c *fiber.Ctx) error {
	var roles []models.Role

	if err := database.DB.Preload("Permisos").Find(&roles).Error; err != nil {
		return utils.ErrorWithCode(c, fiber.StatusInternalServerError, "DB_ERROR", "error al obtener roles")
	}

	return utils.SuccessData(c, fiber.StatusOK, roles)
}

// GetRoleByID obtiene un rol por ID
// GET /api/roles/:id
func GetRoleByID(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return utils.ErrorWithCode(c, fiber.StatusBadRequest, "INVALID_ID", "ID de rol inválido")
	}

	var role models.Role
	if result := database.DB.Preload("Permisos").Preload("Usuarios").First(&role, id); result.Error != nil {
		return utils.ErrorWithCode(c, fiber.StatusNotFound, "NOT_FOUND", "rol no encontrado")
	}

	return utils.SuccessData(c, fiber.StatusOK, role)
}

// CreateRole crea un nuevo rol
// POST /api/roles
func CreateRole(c *fiber.Ctx) error {
	currentUser := middleware.GetUser(c)
	if currentUser == nil {
		return utils.ErrorWithCode(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "no autorizado")
	}

	if !currentUser.HasPermission("roles.create") {
		return utils.ErrorWithCode(c, fiber.StatusForbidden, "FORBIDDEN", "no tienes permiso para crear roles")
	}

	var input struct {
		Nombre      string `json:"nombre" validate:"required"`
		DisplayName string `json:"display_name" validate:"required"`
		Descripcion string `json:"descripcion"`
		Nivel       int    `json:"nivel"`
		PermisoIDs  []uint `json:"permiso_ids"`
	}

	if err := c.BodyParser(&input); err != nil {
		return utils.ErrorWithCode(c, fiber.StatusBadRequest, "INVALID_BODY", "formato de datos inválido")
	}

	// Validar nombre único
	var existing models.Role
	if result := database.DB.Where("nombre = ?", input.Nombre).First(&existing); result.RowsAffected > 0 {
		return utils.ErrorWithCode(c, fiber.StatusConflict, "ROLE_EXISTS", "el nombre del rol ya existe")
	}

	role := models.Role{
		Nombre:      input.Nombre,
		DisplayName: input.DisplayName,
		Descripcion: input.Descripcion,
		Nivel:       input.Nivel,
		Activo:      true,
	}

	// Asignar permisos
	if len(input.PermisoIDs) > 0 {
		var permisos []models.Permission
		database.DB.Find(&permisos, input.PermisoIDs)
		role.Permisos = permisos
	}

	if result := database.DB.Create(&role); result.Error != nil {
		return utils.ErrorWithCode(c, fiber.StatusInternalServerError, "CREATE_ERROR", "error al crear rol")
	}

	// Cargar permisos
	database.DB.Preload("Permisos").First(&role, role.ID)

	return utils.Success(c, fiber.StatusCreated, "rol creado exitosamente", role)
}

// UpdateRole actualiza un rol
// PUT /api/roles/:id
func UpdateRole(c *fiber.Ctx) error {
	currentUser := middleware.GetUser(c)
	if currentUser == nil {
		return utils.ErrorWithCode(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "no autorizado")
	}

	if !currentUser.HasPermission("roles.update") {
		return utils.ErrorWithCode(c, fiber.StatusForbidden, "FORBIDDEN", "no tienes permiso para actualizar roles")
	}

	id, err := c.ParamsInt("id")
	if err != nil {
		return utils.ErrorWithCode(c, fiber.StatusBadRequest, "INVALID_ID", "ID de rol inválido")
	}

	var role models.Role
	if result := database.DB.Preload("Permisos").First(&role, id); result.Error != nil {
		return utils.ErrorWithCode(c, fiber.StatusNotFound, "NOT_FOUND", "rol no encontrado")
	}

	// No permitir modificar roles protegidos
	protected := map[string]bool{"superadmin": true, "admin": true, "vendedor": true, "contador": true, "cliente": true}
	if protected[role.Nombre] {
		return utils.ErrorWithCode(c, fiber.StatusForbidden, "PROTECTED_ROLE", "no se puede modificar un rol protegido")
	}

	var input struct {
		DisplayName string `json:"display_name"`
		Descripcion string `json:"descripcion"`
		Nivel       *int   `json:"nivel"`
		Activo      *bool  `json:"activo"`
		PermisoIDs  []uint `json:"permiso_ids"`
	}

	if err := c.BodyParser(&input); err != nil {
		return utils.ErrorWithCode(c, fiber.StatusBadRequest, "INVALID_BODY", "formato de datos inválido")
	}

	updates := make(map[string]interface{})

	if input.DisplayName != "" {
		updates["display_name"] = input.DisplayName
	}
	if input.Descripcion != "" {
		updates["descripcion"] = input.Descripcion
	}
	if input.Nivel != nil {
		updates["nivel"] = *input.Nivel
	}
	if input.Activo != nil {
		updates["activo"] = *input.Activo
	}

	if len(updates) > 0 {
		database.DB.Model(&role).Updates(updates)
	}

	// Actualizar permisos si se proporcionan
	if input.PermisoIDs != nil {
		var permisos []models.Permission
		database.DB.Find(&permisos, input.PermisoIDs)
		database.DB.Model(&role).Association("Permisos").Replace(permisos)
	}

	// Recargar
	database.DB.Preload("Permisos").First(&role, role.ID)

	return utils.Success(c, fiber.StatusOK, "rol actualizado exitosamente", role)
}

// DeleteRole elimina un rol
// DELETE /api/roles/:id
func DeleteRole(c *fiber.Ctx) error {
	currentUser := middleware.GetUser(c)
	if currentUser == nil {
		return utils.ErrorWithCode(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "no autorizado")
	}

	if !currentUser.HasPermission("roles.delete") {
		return utils.ErrorWithCode(c, fiber.StatusForbidden, "FORBIDDEN", "no tienes permiso para eliminar roles")
	}

	id, err := c.ParamsInt("id")
	if err != nil {
		return utils.ErrorWithCode(c, fiber.StatusBadRequest, "INVALID_ID", "ID de rol inválido")
	}

	var role models.Role
	if result := database.DB.First(&role, id); result.Error != nil {
		return utils.ErrorWithCode(c, fiber.StatusNotFound, "NOT_FOUND", "rol no encontrado")
	}

	// No permitir eliminar roles protegidos
	protected := map[string]bool{"superadmin": true, "admin": true, "vendedor": true, "contador": true, "cliente": true}
	if protected[role.Nombre] {
		return utils.ErrorWithCode(c, fiber.StatusForbidden, "PROTECTED_ROLE", "no se puede eliminar un rol protegido")
	}

	// Verificar que no tenga usuarios asignados
	count := database.DB.Model(&role).Association("Usuarios").Count()
	if count > 0 {
		return utils.ErrorWithCode(c, fiber.StatusConflict, "ROLE_IN_USE", "el rol tiene usuarios asignados")
	}

	if result := database.DB.Delete(&role); result.Error != nil {
		return utils.ErrorWithCode(c, fiber.StatusInternalServerError, "DELETE_ERROR", "error al eliminar rol")
	}

	return utils.SuccessMessage(c, fiber.StatusOK, "rol eliminado exitosamente")
}

// GetPermissions obtiene todos los permisos
// GET /api/permissions
func GetPermissions(c *fiber.Ctx) error {
	modulo := c.Query("modulo", "")
	accion := c.Query("accion", "")

	query := database.DB.Model(&models.Permission{})

	if modulo != "" {
		query = query.Where("modulo = ?", modulo)
	}
	if accion != "" {
		query = query.Where("accion = ?", accion)
	}

	var permissions []models.Permission
	if err := query.Order("modulo, accion").Find(&permissions).Error; err != nil {
		return utils.ErrorWithCode(c, fiber.StatusInternalServerError, "DB_ERROR", "error al obtener permisos")
	}

	return utils.SuccessData(c, fiber.StatusOK, permissions)
}

// GetPermissionByID obtiene un permiso por ID
// GET /api/permissions/:id
func GetPermissionByID(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return utils.ErrorWithCode(c, fiber.StatusBadRequest, "INVALID_ID", "ID de permiso inválido")
	}

	var permission models.Permission
	if result := database.DB.First(&permission, id); result.Error != nil {
		return utils.ErrorWithCode(c, fiber.StatusNotFound, "NOT_FOUND", "permiso no encontrado")
	}

	return utils.SuccessData(c, fiber.StatusOK, permission)
}

// SearchPermissions busca permisos
// GET /api/permissions/search?q=...
func SearchPermissions(c *fiber.Ctx) error {
	q := strings.TrimSpace(c.Query("q", ""))

	var permissions []models.Permission
	query := database.DB.Model(&models.Permission{})

	if q != "" {
		searchPattern := "%" + q + "%"
		query = query.Where("nombre ILIKE ? OR codigo ILIKE ? OR modulo ILIKE ?", searchPattern, searchPattern, searchPattern)
	}

	if err := query.Order("modulo, accion").Limit(20).Find(&permissions).Error; err != nil {
		return utils.ErrorWithCode(c, fiber.StatusInternalServerError, "DB_ERROR", "error al buscar permisos")
	}

	return utils.SuccessData(c, fiber.StatusOK, permissions)
}

// GetMyPermissions obtiene los permisos del usuario actual
// GET /api/auth/permissions
func GetMyPermissions(c *fiber.Ctx) error {
	user := middleware.GetUser(c)
	if user == nil {
		return utils.ErrorWithCode(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "no autorizado")
	}

	// Recargar con permisos
	database.DB.Preload("Roles.Permisos").First(&user, user.ID)

	// Recolectar permisos únicos
	permisosMap := make(map[string]models.Permission)
	for _, role := range user.Roles {
		for _, perm := range role.Permisos {
			permisosMap[perm.Codigo] = perm
		}
	}

	var permissions []models.Permission
	for _, perm := range permisosMap {
		permissions = append(permissions, perm)
	}

	return utils.SuccessData(c, fiber.StatusOK, permissions)
}

// AssignPermissionsToRole asigna permisos a un rol
// POST /api/roles/:id/permissions
func AssignPermissionsToRole(c *fiber.Ctx) error {
	currentUser := middleware.GetUser(c)
	if currentUser == nil {
		return utils.ErrorWithCode(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "no autorizado")
	}

	if !currentUser.HasPermission("roles.update") {
		return utils.ErrorWithCode(c, fiber.StatusForbidden, "FORBIDDEN", "no tienes permiso para modificar permisos")
	}

	id, err := c.ParamsInt("id")
	if err != nil {
		return utils.ErrorWithCode(c, fiber.StatusBadRequest, "INVALID_ID", "ID de rol inválido")
	}

	var input struct {
		PermisoIDs []uint `json:"permiso_ids"`
	}

	if err := c.BodyParser(&input); err != nil {
		return utils.ErrorWithCode(c, fiber.StatusBadRequest, "INVALID_BODY", "formato de datos inválido")
	}

	var role models.Role
	if result := database.DB.First(&role, id); result.Error != nil {
		return utils.ErrorWithCode(c, fiber.StatusNotFound, "NOT_FOUND", "rol no encontrado")
	}

	var permisos []models.Permission
	database.DB.Find(&permisos, input.PermisoIDs)

	if result := database.DB.Model(&role).Association("Permisos").Replace(permisos); result != nil {
		return utils.ErrorWithCode(c, fiber.StatusInternalServerError, "ASSOCIATION_ERROR", "error al asignar permisos")
	}

	database.DB.Preload("Permisos").First(&role, role.ID)

	return utils.Success(c, fiber.StatusOK, "permisos asignados exitosamente", role)
}

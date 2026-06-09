package handlers

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/nexora/backend/internal/database"
	"github.com/nexora/backend/internal/middleware"
	"github.com/nexora/backend/internal/models"
	"github.com/nexora/backend/internal/utils"
)

// GetUsers obtiene la lista de usuarios con paginación
// GET /api/users
func GetUsers(c *fiber.Ctx) error {
	// Parámetros de paginación
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 20)
	search := c.Query("search", "")
	role := c.Query("role", "")
	activo := c.Query("activo", "")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit

	// Construir query
	query := database.DB.Model(&models.User{}).Preload("Roles")

	// Filtros
	if search != "" {
		searchPattern := "%" + search + "%"
		query = query.Where("nombre ILIKE ? OR email ILIKE ? OR apellido ILIKE ?", searchPattern, searchPattern, searchPattern)
	}

	if role != "" {
		query = query.Joins("JOIN user_roles ON user_roles.user_id = users.id JOIN roles ON roles.id = user_roles.role_id").Where("roles.nombre = ?", role)
	}

	if activo != "" {
		query = query.Where("activo = ?", activo == "true")
	}

	// Contar total
	var total int64
	query.Count(&total)

	// Obtener página
	var users []models.User
	if err := query.Offset(offset).Limit(limit).Order("users.created_at DESC").Find(&users).Error; err != nil {
		return utils.ErrorWithCode(c, fiber.StatusInternalServerError, "DB_ERROR", "error al obtener usuarios")
	}

	// Convertir a respuesta
	userResponses := make([]models.UserResponse, len(users))
	for i, user := range users {
		userResponses[i] = user.ToResponse()
	}

	return utils.Paginated(c, fiber.StatusOK, userResponses, page, limit, total)
}

// GetUserByID obtiene un usuario por ID
// GET /api/users/:id
func GetUserByID(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return utils.ErrorWithCode(c, fiber.StatusBadRequest, "INVALID_ID", "ID de usuario inválido")
	}

	var user models.User
	if result := database.DB.Preload("Roles.Permisos").First(&user, id); result.Error != nil {
		return utils.ErrorWithCode(c, fiber.StatusNotFound, "NOT_FOUND", "usuario no encontrado")
	}

	return utils.SuccessData(c, fiber.StatusOK, user.ToResponse())
}

// CreateUser crea un nuevo usuario (admin)
// POST /api/users
func CreateUser(c *fiber.Ctx) error {
	currentUser := middleware.GetUser(c)
	if currentUser == nil {
		return utils.ErrorWithCode(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "no autorizado")
	}

	// Verificar permiso
	if !currentUser.HasPermission("users.create") && currentUser.GetHighestRoleLevel() > 1 {
		return utils.ErrorWithCode(c, fiber.StatusForbidden, "FORBIDDEN", "no tienes permiso para crear usuarios")
	}

	var input struct {
		models.RegisterInput
		RoleIDs []uint `json:"role_ids"`
	}

	if err := c.BodyParser(&input); err != nil {
		return utils.ErrorWithCode(c, fiber.StatusBadRequest, "INVALID_BODY", "formato de datos inválido")
	}

	// Validaciones
	if input.Email == "" || input.Password == "" || input.Nombre == "" {
		return utils.ErrorWithCode(c, fiber.StatusBadRequest, "VALIDATION_ERROR", "email, password y nombre son requeridos")
	}

	// Verificar email único
	var existing models.User
	if result := database.DB.Where("email = ?", input.Email).First(&existing); result.RowsAffected > 0 {
		return utils.ErrorWithCode(c, fiber.StatusConflict, "EMAIL_EXISTS", "el email ya está registrado")
	}

	// Hash contraseña
	hashedPassword, err := utils.HashPassword(input.Password)
	if err != nil {
		return utils.ErrorWithCode(c, fiber.StatusInternalServerError, "HASH_ERROR", "error al procesar contraseña")
	}

	user := models.User{
		Email:           input.Email,
		Password:        hashedPassword,
		Nombre:          input.Nombre,
		Apellido:        input.Apellido,
		Telefono:        input.Telefono,
		TipoDocumento:   input.TipoDocumento,
		NumeroDocumento: input.NumeroDocumento,
		Activo:          true,
	}

	// Asignar roles
	if len(input.RoleIDs) > 0 {
		var roles []models.Role
		database.DB.Find(&roles, input.RoleIDs)
		user.Roles = roles
	}

	if result := database.DB.Create(&user); result.Error != nil {
		return utils.ErrorWithCode(c, fiber.StatusInternalServerError, "CREATE_ERROR", "error al crear usuario")
	}

	// Cargar roles
	database.DB.Preload("Roles").First(&user, user.ID)

	return utils.Success(c, fiber.StatusCreated, "usuario creado exitosamente", user.ToResponse())
}

// UpdateUser actualiza un usuario
// PUT /api/users/:id
func UpdateUser(c *fiber.Ctx) error {
	currentUser := middleware.GetUser(c)
	if currentUser == nil {
		return utils.ErrorWithCode(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "no autorizado")
	}

	id, err := c.ParamsInt("id")
	if err != nil {
		return utils.ErrorWithCode(c, fiber.StatusBadRequest, "INVALID_ID", "ID de usuario inválido")
	}

	// Verificar permisos (solo admins o el mismo usuario)
	if currentUser.ID != uint(id) && !currentUser.HasPermission("users.update") {
		return utils.ErrorWithCode(c, fiber.StatusForbidden, "FORBIDDEN", "no tienes permiso para actualizar este usuario")
	}

	var user models.User
	if result := database.DB.First(&user, id); result.Error != nil {
		return utils.ErrorWithCode(c, fiber.StatusNotFound, "NOT_FOUND", "usuario no encontrado")
	}

	var input models.UpdateUserInput
	if err := c.BodyParser(&input); err != nil {
		return utils.ErrorWithCode(c, fiber.StatusBadRequest, "INVALID_BODY", "formato de datos inválido")
	}

	// Actualizar campos
	updates := make(map[string]interface{})

	if input.Nombre != "" {
		updates["nombre"] = input.Nombre
	}
	if input.Apellido != "" {
		updates["apellido"] = input.Apellido
	}
	if input.Telefono != "" {
		updates["telefono"] = input.Telefono
	}
	if input.Direccion != "" {
		updates["direccion"] = input.Direccion
	}
	if input.Ciudad != "" {
		updates["ciudad"] = input.Ciudad
	}
	if input.Departamento != "" {
		updates["departamento"] = input.Departamento
	}
	if input.CodigoPostal != "" {
		updates["codigo_postal"] = input.CodigoPostal
	}

	// Solo admins pueden cambiar el estado activo
	if input.Activo != nil && currentUser.GetHighestRoleLevel() <= 1 {
		updates["activo"] = *input.Activo
	}

	if len(updates) > 0 {
		if result := database.DB.Model(&user).Updates(updates); result.Error != nil {
			return utils.ErrorWithCode(c, fiber.StatusInternalServerError, "UPDATE_ERROR", "error al actualizar usuario")
		}
	}

	// Recargar usuario
	database.DB.Preload("Roles").First(&user, user.ID)

	database.InvalidateUserCache(user.ID)

	return utils.Success(c, fiber.StatusOK, "usuario actualizado exitosamente", user.ToResponse())
}

// DeleteUser elimina (desactiva) un usuario
// DELETE /api/users/:id
func DeleteUser(c *fiber.Ctx) error {
	currentUser := middleware.GetUser(c)
	if currentUser == nil {
		return utils.ErrorWithCode(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "no autorizado")
	}

	// Solo admins pueden eliminar
	if !currentUser.HasPermission("users.delete") && currentUser.GetHighestRoleLevel() > 1 {
		return utils.ErrorWithCode(c, fiber.StatusForbidden, "FORBIDDEN", "no tienes permiso para eliminar usuarios")
	}

	id, err := c.ParamsInt("id")
	if err != nil {
		return utils.ErrorWithCode(c, fiber.StatusBadRequest, "INVALID_ID", "ID de usuario inválido")
	}

	// No puedes eliminarte a ti mismo
	if currentUser.ID == uint(id) {
		return utils.ErrorWithCode(c, fiber.StatusBadRequest, "SELF_DELETE", "no puedes eliminar tu propia cuenta")
	}

	var user models.User
	if result := database.DB.First(&user, id); result.Error != nil {
		return utils.ErrorWithCode(c, fiber.StatusNotFound, "NOT_FOUND", "usuario no encontrado")
	}

	// Desactivar en lugar de eliminar
	user.Activo = false
	if result := database.DB.Save(&user); result.Error != nil {
		return utils.ErrorWithCode(c, fiber.StatusInternalServerError, "DELETE_ERROR", "error al eliminar usuario")
	}

	return utils.SuccessMessage(c, fiber.StatusOK, "usuario eliminado exitosamente")
}

// AssignRoles asigna roles a un usuario
// POST /api/users/:id/roles
func AssignRoles(c *fiber.Ctx) error {
	currentUser := middleware.GetUser(c)
	if currentUser == nil {
		return utils.ErrorWithCode(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "no autorizado")
	}

	if !currentUser.HasPermission("roles.assign") {
		return utils.ErrorWithCode(c, fiber.StatusForbidden, "FORBIDDEN", "no tienes permiso para asignar roles")
	}

	id, err := c.ParamsInt("id")
	if err != nil {
		return utils.ErrorWithCode(c, fiber.StatusBadRequest, "INVALID_ID", "ID de usuario inválido")
	}

	var input struct {
		RoleIDs []uint `json:"role_ids"`
	}

	if err := c.BodyParser(&input); err != nil {
		return utils.ErrorWithCode(c, fiber.StatusBadRequest, "INVALID_BODY", "formato de datos inválido")
	}

	var user models.User
	if result := database.DB.Preload("Roles").First(&user, id); result.Error != nil {
		return utils.ErrorWithCode(c, fiber.StatusNotFound, "NOT_FOUND", "usuario no encontrado")
	}

	// Obtener roles
	var roles []models.Role
	database.DB.Find(&roles, input.RoleIDs)

	// Reemplazar roles
	if result := database.DB.Model(&user).Association("Roles").Replace(roles); result != nil {
		return utils.ErrorWithCode(c, fiber.StatusInternalServerError, "ASSOCIATION_ERROR", "error al asignar roles")
	}

	// Recargar
	database.DB.Preload("Roles.Permisos").First(&user, user.ID)

	database.InvalidateUserCache(user.ID)

	return utils.Success(c, fiber.StatusOK, "roles asignados exitosamente", user.ToResponse())
}

// SearchUsers busca usuarios por nombre o email
// GET /api/users/search?q=...
func SearchUsers(c *fiber.Ctx) error {
	q := strings.TrimSpace(c.Query("q", ""))
	if q == "" {
		return utils.ErrorWithCode(c, fiber.StatusBadRequest, "EMPTY_QUERY", "parámetro de búsqueda requerido")
	}

	searchPattern := "%" + q + "%"
	var users []models.User

	if err := database.DB.
		Where("nombre ILIKE ? OR email ILIKE ? OR apellido ILIKE ?", searchPattern, searchPattern, searchPattern).
		Preload("Roles").
		Limit(10).
		Find(&users).Error; err != nil {
		return utils.ErrorWithCode(c, fiber.StatusInternalServerError, "DB_ERROR", "error al buscar usuarios")
	}

	responses := make([]models.UserResponse, len(users))
	for i, user := range users {
		responses[i] = user.ToResponse()
	}

	return utils.SuccessData(c, fiber.StatusOK, responses)
}

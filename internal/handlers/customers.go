package handlers

import (
	"regexp"

	"github.com/gofiber/fiber/v2"
	"github.com/nexora/backend/internal/database"
	"github.com/nexora/backend/internal/middleware"
	"github.com/nexora/backend/internal/models"
	"github.com/nexora/backend/internal/utils"
)

// GetCustomers obtiene la lista de clientes con paginación y búsqueda
// GET /api/customers
func GetCustomers(c *fiber.Ctx) error {
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 20)
	search := c.Query("search", "")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit

	query := database.DB.Model(&models.Customer{})

	// Búsqueda por nombre, cédula, email o teléfono
	if search != "" {
		searchPattern := "%" + search + "%"
		query = query.Where(
			"nombre ILIKE ? OR cedula ILIKE ? OR email ILIKE ? OR telefono ILIKE ?",
			searchPattern, searchPattern, searchPattern, searchPattern,
		)
	}

	// Contar total
	var total int64
	query.Count(&total)

	// Obtener página
	var customers []models.Customer
	if err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&customers).Error; err != nil {
		return utils.ErrorWithCode(c, fiber.StatusInternalServerError, "DB_ERROR", "error al obtener clientes")
	}

	// Convertir a respuestas
	responses := make([]models.CustomerResponse, len(customers))
	for i, customer := range customers {
		responses[i] = customer.ToResponse()
	}

	return utils.Paginated(c, fiber.StatusOK, responses, page, limit, total)
}

// GetCustomerByID obtiene un cliente por ID
// GET /api/customers/:id
func GetCustomerByID(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return utils.ErrorWithCode(c, fiber.StatusBadRequest, "INVALID_ID", "ID de cliente inválido")
	}

	var customer models.Customer
	if result := database.DB.First(&customer, id); result.Error != nil {
		return utils.ErrorWithCode(c, fiber.StatusNotFound, "NOT_FOUND", "cliente no encontrado")
	}

	return utils.SuccessData(c, fiber.StatusOK, customer.ToResponse())
}

// GetCustomerByCedula obtiene un cliente por cédula
// GET /api/customers/cedula/:cedula
func GetCustomerByCedula(c *fiber.Ctx) error {
	cedula := c.Params("cedula")

	var customer models.Customer
	if result := database.DB.Where("cedula = ?", cedula).First(&customer); result.Error != nil {
		return utils.ErrorWithCode(c, fiber.StatusNotFound, "NOT_FOUND", "cliente no encontrado")
	}

	return utils.SuccessData(c, fiber.StatusOK, customer.ToResponse())
}

// CreateCustomer crea un nuevo cliente
// POST /api/customers
func CreateCustomer(c *fiber.Ctx) error {
	currentUser := middleware.GetUser(c)
	if currentUser == nil {
		return utils.ErrorWithCode(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "no autorizado")
	}

	if !currentUser.HasPermission("customers.create") {
		return utils.ErrorWithCode(c, fiber.StatusForbidden, "FORBIDDEN", "no tienes permiso para crear clientes")
	}

	var input models.CreateCustomerInput
	if err := c.BodyParser(&input); err != nil {
		return utils.ErrorWithCode(c, fiber.StatusBadRequest, "INVALID_BODY", "formato de datos inválido")
	}

	// Validaciones
	if input.Cedula == "" || input.Nombre == "" || input.Email == "" || input.Telefono == "" {
		return utils.ErrorWithCode(c, fiber.StatusBadRequest, "VALIDATION_ERROR", "todos los campos son requeridos")
	}

	// Validar formato de email
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(input.Email) {
		return utils.ErrorWithCode(c, fiber.StatusBadRequest, "INVALID_EMAIL", "el formato del email es inválido")
	}

	// Verificar cédula única
	var existing models.Customer
	if result := database.DB.Where("cedula = ?", input.Cedula).First(&existing); result.RowsAffected > 0 {
		return utils.ErrorWithCode(c, fiber.StatusConflict, "CEDULA_EXISTS", "ya existe un cliente con esa cédula")
	}

	customer := models.Customer{
		Cedula:   input.Cedula,
		Nombre:   input.Nombre,
		Email:    input.Email,
		Telefono: input.Telefono,
	}

	if result := database.DB.Create(&customer); result.Error != nil {
		return utils.ErrorWithCode(c, fiber.StatusInternalServerError, "CREATE_ERROR", "error al crear cliente")
	}

	return utils.Success(c, fiber.StatusCreated, "cliente creado exitosamente", customer.ToResponse())
}

// UpdateCustomer actualiza un cliente
// PUT /api/customers/:id
func UpdateCustomer(c *fiber.Ctx) error {
	currentUser := middleware.GetUser(c)
	if currentUser == nil {
		return utils.ErrorWithCode(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "no autorizado")
	}

	if !currentUser.HasPermission("customers.update") {
		return utils.ErrorWithCode(c, fiber.StatusForbidden, "FORBIDDEN", "no tienes permiso para actualizar clientes")
	}

	id, err := c.ParamsInt("id")
	if err != nil {
		return utils.ErrorWithCode(c, fiber.StatusBadRequest, "INVALID_ID", "ID de cliente inválido")
	}

	var customer models.Customer
	if result := database.DB.First(&customer, id); result.Error != nil {
		return utils.ErrorWithCode(c, fiber.StatusNotFound, "NOT_FOUND", "cliente no encontrado")
	}

	var input models.UpdateCustomerInput
	if err := c.BodyParser(&input); err != nil {
		return utils.ErrorWithCode(c, fiber.StatusBadRequest, "INVALID_BODY", "formato de datos inválido")
	}

	updates := make(map[string]interface{})

	if input.Cedula != "" {
		// Verificar que la cédula no esté en uso por otro cliente
		var existing models.Customer
		if result := database.DB.Where("cedula = ? AND id != ?", input.Cedula, customer.ID).First(&existing); result.RowsAffected > 0 {
			return utils.ErrorWithCode(c, fiber.StatusConflict, "CEDULA_EXISTS", "ya existe un cliente con esa cédula")
		}
		updates["cedula"] = input.Cedula
	}

	if input.Nombre != "" {
		updates["nombre"] = input.Nombre
	}

	if input.Email != "" {
		// Validar formato de email
		emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
		if !emailRegex.MatchString(input.Email) {
			return utils.ErrorWithCode(c, fiber.StatusBadRequest, "INVALID_EMAIL", "el formato del email es inválido")
		}
		updates["email"] = input.Email
	}

	if input.Telefono != "" {
		updates["telefono"] = input.Telefono
	}

	if len(updates) > 0 {
		if result := database.DB.Model(&customer).Updates(updates); result.Error != nil {
			return utils.ErrorWithCode(c, fiber.StatusInternalServerError, "UPDATE_ERROR", "error al actualizar cliente")
		}
	}

	return utils.Success(c, fiber.StatusOK, "cliente actualizado exitosamente", customer.ToResponse())
}

// DeleteCustomer elimina un cliente
// DELETE /api/customers/:id
func DeleteCustomer(c *fiber.Ctx) error {
	currentUser := middleware.GetUser(c)
	if currentUser == nil {
		return utils.ErrorWithCode(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "no autorizado")
	}

	if !currentUser.HasPermission("customers.delete") {
		return utils.ErrorWithCode(c, fiber.StatusForbidden, "FORBIDDEN", "no tienes permiso para eliminar clientes")
	}

	id, err := c.ParamsInt("id")
	if err != nil {
		return utils.ErrorWithCode(c, fiber.StatusBadRequest, "INVALID_ID", "ID de cliente inválido")
	}

	var customer models.Customer
	if result := database.DB.First(&customer, id); result.Error != nil {
		return utils.ErrorWithCode(c, fiber.StatusNotFound, "NOT_FOUND", "cliente no encontrado")
	}

	// Verificar si tiene separados activos
	var layawayCount int64
	database.DB.Model(&models.Layaway{}).Where("customer_id = ? AND estado = ?", customer.ID, "activo").Count(&layawayCount)
	if layawayCount > 0 {
		return utils.ErrorWithCode(c, fiber.StatusBadRequest, "HAS_ACTIVE_LAYAWAYS", "el cliente tiene separados activos, no se puede eliminar")
	}

	// Soft delete
	if result := database.DB.Delete(&customer); result.Error != nil {
		return utils.ErrorWithCode(c, fiber.StatusInternalServerError, "DELETE_ERROR", "error al eliminar cliente")
	}

	return utils.SuccessMessage(c, fiber.StatusOK, "cliente eliminado exitosamente")
}

package handlers

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/nexora/backend/internal/database"
	"github.com/nexora/backend/internal/middleware"
	"github.com/nexora/backend/internal/models"
	"github.com/nexora/backend/internal/utils"
)

// GetLayaways obtiene la lista de separados con paginación y filtros
// GET /api/layaways
func GetLayaways(c *fiber.Ctx) error {
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 20)
	search := c.Query("search", "")
	estado := c.Query("estado", "")
	customerID := c.QueryInt("customer_id", 0)

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit

	query := database.DB.Model(&models.Layaway{}).
		Preload("Customer").
		Preload("Product")

	// Filtros
	if search != "" {
		searchPattern := "%" + search + "%"
		query = query.Where(
			"(EXISTS (SELECT 1 FROM customers WHERE customers.id = layaways.customer_id AND (customers.nombre ILIKE ? OR customers.cedula ILIKE ?)) OR "+
				"EXISTS (SELECT 1 FROM products WHERE products.id = layaways.product_id AND products.nombre ILIKE ?))",
			searchPattern, searchPattern, searchPattern,
		)
	}

	if estado != "" {
		query = query.Where("estado = ?", estado)
	}

	if customerID > 0 {
		query = query.Where("customer_id = ?", customerID)
	}

	// Contar total
	var total int64
	query.Count(&total)

	// Obtener página
	var layaways []models.Layaway
	if err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&layaways).Error; err != nil {
		return utils.ErrorWithCode(c, fiber.StatusInternalServerError, "DB_ERROR", "error al obtener separados")
	}

	// Convertir a respuestas
	responses := make([]models.LayawayResponse, len(layaways))
	for i, layaway := range layaways {
		responses[i] = layaway.ToResponse()
	}

	return utils.Paginated(c, fiber.StatusOK, responses, page, limit, total)
}

// GetLayawayByID obtiene un separado por ID
// GET /api/layaways/:id
func GetLayawayByID(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return utils.ErrorWithCode(c, fiber.StatusBadRequest, "INVALID_ID", "ID de separado inválido")
	}

	var layaway models.Layaway
	if result := database.DB.
		Preload("Customer").
		Preload("Product").
		Preload("Abonos").
		First(&layaway, id); result.Error != nil {
		return utils.ErrorWithCode(c, fiber.StatusNotFound, "NOT_FOUND", "separado no encontrado")
	}

	return utils.SuccessData(c, fiber.StatusOK, layaway.ToResponse())
}

// CreateLayaway crea un nuevo separado
// POST /api/layaways
func CreateLayaway(c *fiber.Ctx) error {
	currentUser := middleware.GetUser(c)
	if currentUser == nil {
		return utils.ErrorWithCode(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "no autorizado")
	}

	if !currentUser.HasPermission("layaways.create") {
		return utils.ErrorWithCode(c, fiber.StatusForbidden, "FORBIDDEN", "no tienes permiso para crear separados")
	}

	var input models.CreateLayawayInput
	if err := c.BodyParser(&input); err != nil {
		return utils.ErrorWithCode(c, fiber.StatusBadRequest, "INVALID_BODY", "formato de datos inválido")
	}

	// Validaciones
	if input.CustomerID == 0 || input.ProductID == 0 || input.Cantidad < 1 {
		return utils.ErrorWithCode(c, fiber.StatusBadRequest, "VALIDATION_ERROR", "cliente, producto y cantidad son requeridos")
	}

	// Verificar que el cliente existe
	var customer models.Customer
	if result := database.DB.First(&customer, input.CustomerID); result.Error != nil {
		return utils.ErrorWithCode(c, fiber.StatusNotFound, "CUSTOMER_NOT_FOUND", "cliente no encontrado")
	}

	// Verificar que el producto existe y obtener su precio
	var product models.Product
	if result := database.DB.First(&product, input.ProductID); result.Error != nil {
		return utils.ErrorWithCode(c, fiber.StatusNotFound, "PRODUCT_NOT_FOUND", "producto no encontrado")
	}

	// Calcular precio total
	precioTotal := product.Precio * float64(input.Cantidad)

	// Validar que el abono inicial no sea mayor al total
	if input.AbonoInicial > precioTotal {
		return utils.ErrorWithCode(c, fiber.StatusBadRequest, "INVALID_INITIAL_PAYMENT", "el abono inicial no puede ser mayor al precio total")
	}

	// Calcular saldo pendiente
	saldoPendiente := precioTotal - input.AbonoInicial

	// Calcular fecha de vencimiento (3 meses desde ahora)
	fechaVencimiento := time.Now().AddDate(0, 3, 0)

	layaway := models.Layaway{
		CustomerID:       input.CustomerID,
		ProductID:        input.ProductID,
		Cantidad:         input.Cantidad,
		PrecioUnitario:   product.Precio,
		PrecioTotal:      precioTotal,
		AbonoInicial:     input.AbonoInicial,
		FechaVencimiento: fechaVencimiento,
		Estado:           "activo",
		SaldoPendiente:   saldoPendiente,
	}

	// Crear el registrado el abono inicial como primer pago si es mayor a 0
	tx := database.DB.Begin()

	if result := tx.Create(&layaway); result.Error != nil {
		tx.Rollback()
		return utils.ErrorWithCode(c, fiber.StatusInternalServerError, "CREATE_ERROR", "error al crear separado")
	}

	// Registrar el abono inicial como primer pago
	if input.AbonoInicial > 0 {
		payment := models.Payment{
			LayawayID: layaway.ID,
			Monto:     input.AbonoInicial,
			FechaPago: time.Now(),
		}
		if result := tx.Create(&payment); result.Error != nil {
			tx.Rollback()
			return utils.ErrorWithCode(c, fiber.StatusInternalServerError, "CREATE_ERROR", "error al registrar el abono inicial")
		}
	}

	if err := tx.Commit().Error; err != nil {
		return utils.ErrorWithCode(c, fiber.StatusInternalServerError, "DB_ERROR", "error al guardar separado")
	}

	// Recargar relaciones
	database.DB.Preload("Customer").Preload("Product").Preload("Abonos").First(&layaway, layaway.ID)

	return utils.Success(c, fiber.StatusCreated, "separado creado exitosamente", layaway.ToResponse())
}

// AddPayment registra un nuevo abono a un separado
// POST /api/layaways/:id/payments
func AddPayment(c *fiber.Ctx) error {
	currentUser := middleware.GetUser(c)
	if currentUser == nil {
		return utils.ErrorWithCode(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "no autorizado")
	}

	if !currentUser.HasPermission("layaways.update") {
		return utils.ErrorWithCode(c, fiber.StatusForbidden, "FORBIDDEN", "no tienes permiso para registrar abonos")
	}

	id, err := c.ParamsInt("id")
	if err != nil {
		return utils.ErrorWithCode(c, fiber.StatusBadRequest, "INVALID_ID", "ID de separado inválido")
	}

	var layaway models.Layaway
	if result := database.DB.First(&layaway, id); result.Error != nil {
		return utils.ErrorWithCode(c, fiber.StatusNotFound, "NOT_FOUND", "separado no encontrado")
	}

	// Verificar que el separado esté activo
	if layaway.Estado != "activo" {
		return utils.ErrorWithCode(c, fiber.StatusBadRequest, "INVALID_STATE", "el separado no está activo")
	}

	// Verificar que no esté vencido
	if time.Now().After(layaway.FechaVencimiento) {
		return utils.ErrorWithCode(c, fiber.StatusBadRequest, "EXPIRED", "el separado está vencido")
	}

	var input models.CreatePaymentInput
	if err := c.BodyParser(&input); err != nil {
		return utils.ErrorWithCode(c, fiber.StatusBadRequest, "INVALID_BODY", "formato de datos inválido")
	}

	// Validar monto
	if input.Monto <= 0 {
		return utils.ErrorWithCode(c, fiber.StatusBadRequest, "INVALID_AMOUNT", "el monto debe ser mayor a 0")
	}

	// Validar que el monto no sea mayor al saldo pendiente
	if input.Monto > layaway.SaldoPendiente {
		return utils.ErrorWithCode(c, fiber.StatusBadRequest, "EXCEEDS_BALANCE", "el monto no puede ser mayor al saldo pendiente")
	}

	tx := database.DB.Begin()

	// Crear el pago
	payment := models.Payment{
		LayawayID: layaway.ID,
		Monto:     input.Monto,
		FechaPago: time.Now(),
	}

	if result := tx.Create(&payment); result.Error != nil {
		tx.Rollback()
		return utils.ErrorWithCode(c, fiber.StatusInternalServerError, "CREATE_ERROR", "error al registrar el abono")
	}

	// Actualizar saldo pendiente
	nuevoSaldo := layaway.SaldoPendiente - input.Monto
	updates := map[string]interface{}{
		"saldo_pendiente": nuevoSaldo,
	}

	// Si el saldo es 0, marcar como pagado
	if nuevoSaldo <= 0 {
		updates["estado"] = "pagado"
	}

	if result := tx.Model(&layaway).Updates(updates); result.Error != nil {
		tx.Rollback()
		return utils.ErrorWithCode(c, fiber.StatusInternalServerError, "UPDATE_ERROR", "error al actualizar el separado")
	}

	if err := tx.Commit().Error; err != nil {
		return utils.ErrorWithCode(c, fiber.StatusInternalServerError, "DB_ERROR", "error al guardar el abono")
	}

	// Recargar
	database.DB.Preload("Customer").Preload("Product").Preload("Abonos").First(&layaway, layaway.ID)

	message := "abono registrado exitosamente"
	if nuevoSaldo <= 0 {
		message = "separado pagado completamente"
	}

	return utils.Success(c, fiber.StatusCreated, message, fiber.Map{
		"separado":       layaway.ToResponse(),
		"abonorealizado": payment.ToResponse(),
		"saldo_anterior": layaway.SaldoPendiente + input.Monto,
		"monto_abonado":  input.Monto,
		"saldo_nuevo":    nuevoSaldo,
	})
}

// GetLayawayPayments obtiene los abonos de un separado
// GET /api/layaways/:id/payments
func GetLayawayPayments(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return utils.ErrorWithCode(c, fiber.StatusBadRequest, "INVALID_ID", "ID de separado inválido")
	}

	var layaway models.Layaway
	if result := database.DB.First(&layaway, id); result.Error != nil {
		return utils.ErrorWithCode(c, fiber.StatusNotFound, "NOT_FOUND", "separado no encontrado")
	}

	var payments []models.Payment
	if result := database.DB.Where("layaway_id = ?", layaway.ID).Order("fecha_pago ASC").Find(&payments); result.Error != nil {
		return utils.ErrorWithCode(c, fiber.StatusInternalServerError, "DB_ERROR", "error al obtener abonos")
	}

	responses := make([]models.PaymentResponse, len(payments))
	for i, payment := range payments {
		responses[i] = payment.ToResponse()
	}

	return utils.SuccessData(c, fiber.StatusOK, fiber.Map{
		"separado":       layaway.ToResponse(),
		"abonos":         responses,
		"saldo_pendiente": layaway.SaldoPendiente,
	})
}

// UpdateLayawayStatus actualiza el estado de un separado
// PUT /api/layaways/:id
func UpdateLayawayStatus(c *fiber.Ctx) error {
	currentUser := middleware.GetUser(c)
	if currentUser == nil {
		return utils.ErrorWithCode(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "no autorizado")
	}

	if !currentUser.HasPermission("layaways.update") {
		return utils.ErrorWithCode(c, fiber.StatusForbidden, "FORBIDDEN", "no tienes permiso para actualizar separados")
	}

	id, err := c.ParamsInt("id")
	if err != nil {
		return utils.ErrorWithCode(c, fiber.StatusBadRequest, "INVALID_ID", "ID de separado inválido")
	}

	var layaway models.Layaway
	if result := database.DB.First(&layaway, id); result.Error != nil {
		return utils.ErrorWithCode(c, fiber.StatusNotFound, "NOT_FOUND", "separado no encontrado")
	}

	var input models.UpdateLayawayInput
	if err := c.BodyParser(&input); err != nil {
		return utils.ErrorWithCode(c, fiber.StatusBadRequest, "INVALID_BODY", "formato de datos inválido")
	}

	if input.Estado == "" {
		return utils.ErrorWithCode(c, fiber.StatusBadRequest, "VALIDATION_ERROR", "el estado es requerido")
	}

	// Validar estado válido
	validStates := map[string]bool{
		"activo":    true,
		"pagado":    true,
		"vencido":   true,
		"cancelado": true,
	}

	if !validStates[input.Estado] {
		return utils.ErrorWithCode(c, fiber.StatusBadRequest, "INVALID_STATE", "estado inválido")
	}

	// No permitir cambiar estado si ya está pagado
	if layaway.Estado == "pagado" {
		return utils.ErrorWithCode(c, fiber.StatusBadRequest, "ALREADY_PAID", "el separado ya está pagado")
	}

	if result := database.DB.Model(&layaway).Update("estado", input.Estado); result.Error != nil {
		return utils.ErrorWithCode(c, fiber.StatusInternalServerError, "UPDATE_ERROR", "error al actualizar el estado del separado")
	}

	return utils.Success(c, fiber.StatusOK, "estado del separado actualizado", layaway.ToResponse())
}

// CancelLayaway cancela un separado
// POST /api/layaways/:id/cancel
func CancelLayaway(c *fiber.Ctx) error {
	currentUser := middleware.GetUser(c)
	if currentUser == nil {
		return utils.ErrorWithCode(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "no autorizado")
	}

	if !currentUser.HasPermission("layaways.delete") {
		return utils.ErrorWithCode(c, fiber.StatusForbidden, "FORBIDDEN", "no tienes permiso para cancelar separados")
	}

	id, err := c.ParamsInt("id")
	if err != nil {
		return utils.ErrorWithCode(c, fiber.StatusBadRequest, "INVALID_ID", "ID de separado inválido")
	}

	var layaway models.Layaway
	if result := database.DB.First(&layaway, id); result.Error != nil {
		return utils.ErrorWithCode(c, fiber.StatusNotFound, "NOT_FOUND", "separado no encontrado")
	}

	if layaway.Estado == "pagado" {
		return utils.ErrorWithCode(c, fiber.StatusBadRequest, "ALREADY_PAID", "el separado ya está pagado")
	}

	if layaway.Estado == "cancelado" {
		return utils.ErrorWithCode(c, fiber.StatusBadRequest, "ALREADY_CANCELLED", "el separado ya está cancelado")
	}

	if result := database.DB.Model(&layaway).Update("estado", "cancelado"); result.Error != nil {
		return utils.ErrorWithCode(c, fiber.StatusInternalServerError, "UPDATE_ERROR", "error al cancelar el separado")
	}

	database.DB.Preload("Customer").Preload("Product").First(&layaway, layaway.ID)

	return utils.Success(c, fiber.StatusOK, "separado cancelado exitosamente", layaway.ToResponse())
}

// ExpiredLayaways marca los separados vencidos
// POST /api/layaways/check-expired
func ExpiredLayaways(c *fiber.Ctx) error {
	currentUser := middleware.GetUser(c)
	if currentUser == nil {
		return utils.ErrorWithCode(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "no autorizado")
	}

	if !currentUser.HasPermission("layaways.update") {
		return utils.ErrorWithCode(c, fiber.StatusForbidden, "FORBIDDEN", "no tienes permiso para gestionar separados")
	}

	now := time.Now()

	result := database.DB.Model(&models.Layaway{}).
		Where("estado = ? AND fecha_vencimiento < ?", "activo", now).
		Update("estado", "vencido")

	if result.Error != nil {
		return utils.ErrorWithCode(c, fiber.StatusInternalServerError, "UPDATE_ERROR", "error al verificar separados vencidos")
	}

	return utils.Success(c, fiber.StatusOK, "verificación completada", fiber.Map{
		"separados_vencidos": result.RowsAffected,
	})
}

// DeleteLayaway elimina un separado
// DELETE /api/layaways/:id
func DeleteLayaway(c *fiber.Ctx) error {
	currentUser := middleware.GetUser(c)
	if currentUser == nil {
		return utils.ErrorWithCode(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "no autorizado")
	}

	if !currentUser.HasPermission("layaways.delete") {
		return utils.ErrorWithCode(c, fiber.StatusForbidden, "FORBIDDEN", "no tienes permiso para eliminar separados")
	}

	id, err := c.ParamsInt("id")
	if err != nil {
		return utils.ErrorWithCode(c, fiber.StatusBadRequest, "INVALID_ID", "ID de separado inválido")
	}

	var layaway models.Layaway
	if result := database.DB.First(&layaway, id); result.Error != nil {
		return utils.ErrorWithCode(c, fiber.StatusNotFound, "NOT_FOUND", "separado no encontrado")
	}

	// Soft delete
	if result := database.DB.Delete(&layaway); result.Error != nil {
		return utils.ErrorWithCode(c, fiber.StatusInternalServerError, "DELETE_ERROR", "error al eliminar separado")
	}

	return utils.SuccessMessage(c, fiber.StatusOK, "separado eliminado exitosamente")
}

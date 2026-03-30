package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/nexora/backend/internal/database"
	"github.com/nexora/backend/internal/middleware"
	"github.com/nexora/backend/internal/models"
	"github.com/nexora/backend/internal/services"
	"github.com/nexora/backend/internal/utils"
)

// GetInvoices obtiene la lista de facturas
// GET /api/invoices
func GetInvoices(c *fiber.Ctx) error {
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 20)
	estado := c.Query("estado", "")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit

	query := database.DB.Model(&models.Invoice{}).Preload("Orden")

	if estado != "" {
		query = query.Where("estado = ?", estado)
	}

	var total int64
	query.Count(&total)

	var invoices []models.Invoice
	if err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&invoices).Error; err != nil {
		return utils.ErrorWithCode(c, fiber.StatusInternalServerError, "DB_ERROR", "error al obtener facturas")
	}

	return utils.Paginated(c, fiber.StatusOK, invoices, page, limit, total)
}

// GetInvoiceByID obtiene una factura por ID
// GET /api/invoices/:id
func GetInvoiceByID(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return utils.ErrorWithCode(c, fiber.StatusBadRequest, "INVALID_ID", "ID de factura inválido")
	}

	var invoice models.Invoice
	if result := database.DB.
		Preload("Orden").
		Preload("Orden.Items").
		First(&invoice, id); result.Error != nil {
		return utils.ErrorWithCode(c, fiber.StatusNotFound, "NOT_FOUND", "factura no encontrada")
	}

	return utils.SuccessData(c, fiber.StatusOK, invoice)
}

// CreateInvoiceFromOrder crea una factura electrónica desde un pedido
// POST /api/invoices
func CreateInvoiceFromOrder(c *fiber.Ctx) error {
	currentUser := middleware.GetUser(c)
	if currentUser == nil {
		return utils.ErrorWithCode(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "no autorizado")
	}

	if !currentUser.HasPermission("invoices.create") {
		return utils.ErrorWithCode(c, fiber.StatusForbidden, "FORBIDDEN", "no tienes permiso para crear facturas")
	}

	var input struct {
		OrdenID uint `json:"orden_id" validate:"required"`
	}
	if err := c.BodyParser(&input); err != nil {
		return utils.ErrorWithCode(c, fiber.StatusBadRequest, "INVALID_BODY", "formato de datos inválido")
	}

	var order models.Order
	if result := database.DB.Preload("Items").First(&order, input.OrdenID); result.Error != nil {
		return utils.ErrorWithCode(c, fiber.StatusNotFound, "NOT_FOUND", "pedido no encontrado")
	}

	// Verificar si ya tiene factura
	var existing models.Invoice
	if result := database.DB.Where("orden_id = ?", input.OrdenID).First(&existing); result.RowsAffected > 0 {
		return utils.ErrorWithCode(c, fiber.StatusConflict, "INVOICE_EXISTS", "el pedido ya tiene una factura")
	}

	dian := services.NewDIANService()
	invoice, err := dian.CreateInvoiceFromOrder(&order)
	if err != nil {
		return utils.ErrorWithCode(c, fiber.StatusInternalServerError, "CREATE_ERROR", err.Error())
	}

	return utils.Success(c, fiber.StatusCreated, "factura creada exitosamente", invoice)
}

// SendInvoiceToDIAN envía una factura a la DIAN
// POST /api/invoices/:id/send
func SendInvoiceToDIAN(c *fiber.Ctx) error {
	currentUser := middleware.GetUser(c)
	if currentUser == nil {
		return utils.ErrorWithCode(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "no autorizado")
	}

	if !currentUser.HasPermission("invoices.send") {
		return utils.ErrorWithCode(c, fiber.StatusForbidden, "FORBIDDEN", "no tienes permiso para enviar facturas")
	}

	id, err := c.ParamsInt("id")
	if err != nil {
		return utils.ErrorWithCode(c, fiber.StatusBadRequest, "INVALID_ID", "ID de factura inválido")
	}

	var invoice models.Invoice
	if result := database.DB.First(&invoice, id); result.Error != nil {
		return utils.ErrorWithCode(c, fiber.StatusNotFound, "NOT_FOUND", "factura no encontrada")
	}

	if invoice.EstadoDIAN == "aceptada" {
		return utils.ErrorWithCode(c, fiber.StatusBadRequest, "ALREADY_SENT", "la factura ya fue aceptada por DIAN")
	}

	dian := services.NewDIANService()
	if err := dian.SendToDIAN(&invoice); err != nil {
		return utils.ErrorWithCode(c, fiber.StatusInternalServerError, "DIAN_ERROR", err.Error())
	}

	return utils.Success(c, fiber.StatusOK, "factura enviada a DIAN exitosamente", invoice)
}

// GetInvoiceXML obtiene el XML de una factura
// GET /api/invoices/:id/xml
func GetInvoiceXML(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return utils.ErrorWithCode(c, fiber.StatusBadRequest, "INVALID_ID", "ID de factura inválido")
	}

	var invoice models.Invoice
	if result := database.DB.First(&invoice, id); result.Error != nil {
		return utils.ErrorWithCode(c, fiber.StatusNotFound, "NOT_FOUND", "factura no encontrada")
	}

	if invoice.XMLContent == "" {
		dian := services.NewDIANService()
		xml, err := dian.GenerateInvoiceXML(&invoice)
		if err != nil {
			return utils.ErrorWithCode(c, fiber.StatusInternalServerError, "XML_ERROR", err.Error())
		}
		invoice.XMLContent = xml
		database.DB.Save(&invoice)
	}

	c.Set("Content-Type", "application/xml")
	return c.SendString(invoice.XMLContent)
}

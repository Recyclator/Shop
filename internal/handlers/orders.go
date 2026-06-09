package handlers

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/nexora/backend/internal/database"
	"github.com/nexora/backend/internal/middleware"
	"github.com/nexora/backend/internal/models"
	"github.com/nexora/backend/internal/utils"
	"gorm.io/gorm/clause"
)

// GetOrders obtiene la lista de pedidos
// GET /api/orders
func GetOrders(c *fiber.Ctx) error {
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 20)
	estado := c.Query("estado", "")
	vendedorID := c.QueryInt("vendedor", 0)
	fechaDesde := c.Query("fecha_desde", "")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit

	query := database.DB.Model(&models.Order{}).Preload("Cliente").Preload("Vendedor").Preload("Items")

	if estado != "" {
		query = query.Where("estado = ?", estado)
	}

	if vendedorID > 0 {
		query = query.Where("vendedor_id = ?", vendedorID)
	}

	if fechaDesde != "" {
		query = query.Where("date(created_at) >= ?", fechaDesde)
	}

	var total int64
	query.Count(&total)

	var orders []models.Order
	if err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&orders).Error; err != nil {
		return utils.ErrorWithCode(c, fiber.StatusInternalServerError, "DB_ERROR", "error al obtener pedidos")
	}

	responses := make([]models.OrderResponse, len(orders))
	for i, o := range orders {
		responses[i] = o.ToResponse()
	}

	return utils.Paginated(c, fiber.StatusOK, responses, page, limit, total)
}

// GetOrderByID obtiene un pedido por ID
// GET /api/orders/:id
func GetOrderByID(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return utils.ErrorWithCode(c, fiber.StatusBadRequest, "INVALID_ID", "ID de pedido inválido")
	}

	var order models.Order
	if result := database.DB.
		Preload("Cliente").
		Preload("Vendedor").
		Preload("Items").
		Preload("Items.Producto").
		First(&order, id); result.Error != nil {
		return utils.ErrorWithCode(c, fiber.StatusNotFound, "NOT_FOUND", "pedido no encontrado")
	}

	return utils.SuccessData(c, fiber.StatusOK, order.ToResponse())
}

// CreateOrder crea un nuevo pedido
// POST /api/orders
func CreateOrder(c *fiber.Ctx) error {
	currentUser := middleware.GetUser(c)
	if currentUser == nil {
		return utils.ErrorWithCode(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "no autorizado")
	}

	if !currentUser.HasPermission("orders.create") {
		return utils.ErrorWithCode(c, fiber.StatusForbidden, "FORBIDDEN", "no tienes permiso para crear pedidos")
	}

	var input models.CreateOrderInput
	if err := c.BodyParser(&input); err != nil {
		return utils.ErrorWithCode(c, fiber.StatusBadRequest, "INVALID_BODY", "formato de datos inválido")
	}

	if len(input.Items) == 0 {
		return utils.ErrorWithCode(c, fiber.StatusBadRequest, "NO_ITEMS", "el pedido debe tener al menos un producto")
	}

	// Generar número de orden
	numeroOrden := generateOrderNumber()

	// Calcular totales
	subtotal := 0.0
	impuesto := 0.0
	impuestoPorcentaje := 19.0 // IVA Colombia

	items := make([]models.OrderItem, len(input.Items))

	tx := database.DB.Begin()

	for i, item := range input.Items {
		var product models.Product
		if result := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&product, item.ProductoID); result.Error != nil {
			tx.Rollback()
			return utils.ErrorWithCode(c, fiber.StatusBadRequest, "INVALID_PRODUCT", fmt.Sprintf("producto %d no encontrado", item.ProductoID))
		}

		if product.ControlaInventario && product.Stock < item.Cantidad && !product.PermiteStockNegativo {
			tx.Rollback()
			return utils.ErrorWithCode(c, fiber.StatusBadRequest, "INSUFFICIENT_STOCK", fmt.Sprintf("stock insuficiente para %s", product.Nombre))
		}

		itemTotal := item.PrecioUnitario * float64(item.Cantidad)
		itemImpuesto := itemTotal * (impuestoPorcentaje / 100)

		items[i] = models.OrderItem{
			ProductoID:     item.ProductoID,
			NombreProducto: product.Nombre,
			SKU:            product.SKU,
			VarianteID:     item.VarianteID,
			Cantidad:       item.Cantidad,
			PrecioUnitario: item.PrecioUnitario,
			Impuesto:       itemImpuesto,
			Total:          itemTotal + itemImpuesto,
		}

		subtotal += itemTotal
		impuesto += itemImpuesto

		if product.ControlaInventario {
			if err := tx.Model(&product).Update("stock", product.Stock-item.Cantidad).Error; err != nil {
				tx.Rollback()
				return utils.ErrorWithCode(c, fiber.StatusInternalServerError, "STOCK_ERROR", "error al actualizar stock")
			}
		}
	}

	// Calcular descuento
	descuento := 0.0
	if input.CodigoDescuento != "" {
		// Por simplicidad, descuento del 10% si hay código
		descuento = subtotal * 0.10
	}

	total := subtotal + impuesto - descuento

	var clienteID uint
	if input.ClienteID != nil {
		clienteID = *input.ClienteID
	}

	order := models.Order{
		NumeroOrden:        numeroOrden,
		ClienteID:          clienteID,
		ClienteNombre:      "Cliente Mostrador",
		Estado:             "completado",
		MetodoPago:         input.MetodoPago,
		ReferenciaPago:     input.ReferenciaPago,
		FechaPago:          nil,
		EstadoPago:         "pagado",
		Subtotal:           subtotal,
		Descuento:          descuento,
		Impuesto:           impuesto,
		ImpuestoPorcentaje: impuestoPorcentaje,
		Total:              total,
		VendedorID:         &currentUser.ID,
		Notas:              input.Notas,
		Items:              items,
	}

	if result := tx.Create(&order); result.Error != nil {
		tx.Rollback()
		return utils.ErrorWithCode(c, fiber.StatusInternalServerError, "CREATE_ERROR", "error al crear pedido")
	}

	if err := tx.Commit().Error; err != nil {
		return utils.ErrorWithCode(c, fiber.StatusInternalServerError, "COMMIT_ERROR", "error al confirmar la transacción")
	}

	// Guardar cliente si se proporciona
	if input.ClienteID != nil {
		var cliente models.User
		if result := database.DB.First(&cliente, *input.ClienteID); result.Error == nil {
			order.ClienteNombre = cliente.Nombre
			order.ClienteEmail = cliente.Email
		}
	}

	if result := database.DB.Create(&order); result.Error != nil {
		return utils.ErrorWithCode(c, fiber.StatusInternalServerError, "CREATE_ERROR", "error al crear pedido")
	}

	// Cargar relaciones
	database.DB.Preload("Cliente").Preload("Vendedor").Preload("Items").First(&order, order.ID)

	return utils.Success(c, fiber.StatusCreated, "pedido creado exitosamente", order.ToResponse())
}

// QuickPOSSale crea una venta rápida desde el POS
// POST /api/orders/pos
func QuickPOSSale(c *fiber.Ctx) error {
	currentUser := middleware.GetUser(c)
	if currentUser == nil {
		return utils.ErrorWithCode(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "no autorizado")
	}

	if !currentUser.HasPermission("orders.create") {
		return utils.ErrorWithCode(c, fiber.StatusForbidden, "FORBIDDEN", "no tienes permiso para ventas")
	}

	var input struct {
		Items []struct {
			ProductoID uint  `json:"producto_id"`
			VariantID  *uint `json:"variant_id"`
			Cantidad   int   `json:"cantidad"`
		} `json:"items" validate:"required,min=1"`
		MetodoPago string  `json:"metodo_pago"`
		Recibido   float64 `json:"recibido"`
	}

	if err := c.BodyParser(&input); err != nil {
		return utils.ErrorWithCode(c, fiber.StatusBadRequest, "INVALID_BODY", "formato de datos inválido")
	}

	if len(input.Items) == 0 {
		return utils.ErrorWithCode(c, fiber.StatusBadRequest, "NO_ITEMS", "debe agregar al menos un producto")
	}

	numeroOrden := generateOrderNumber()

	subtotal := 0.0
	impuesto := 0.0
	impuestoPorcentaje := 19.0
	items := make([]models.OrderItem, len(input.Items))

	tx := database.DB.Begin()

	for i, item := range input.Items {
		var product models.Product
		if result := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&product, item.ProductoID); result.Error != nil {
			tx.Rollback()
			return utils.ErrorWithCode(c, fiber.StatusBadRequest, "INVALID_PRODUCT", fmt.Sprintf("producto %d no encontrado", item.ProductoID))
		}

		var variant *models.ProductVariant
		var precioUnitario float64 = product.Precio
		var sku string = product.SKU
		var variantInfo string

		if item.VariantID != nil {
			variant = new(models.ProductVariant)
			if result := tx.Preload("Atributos").Clauses(clause.Locking{Strength: "UPDATE"}).First(variant, *item.VariantID); result.Error != nil {
				tx.Rollback()
				return utils.ErrorWithCode(c, fiber.StatusBadRequest, "INVALID_VARIANT", fmt.Sprintf("variante %d no encontrada", *item.VariantID))
			}
			if variant.ProductoID != item.ProductoID {
				tx.Rollback()
				return utils.ErrorWithCode(c, fiber.StatusBadRequest, "VARIANT_MISMATCH", "la variante no pertenece al producto")
			}
			if variant.PrecioOverride > 0 {
				precioUnitario = variant.PrecioOverride
			}
			sku = variant.SKU
			for _, attr := range variant.Atributos {
				if variantInfo != "" {
					variantInfo += ", "
				}
				variantInfo += attr.Value
			}
		}

		stockActual := product.Stock
		if variant != nil {
			stockActual = variant.Stock
		}

		if product.ControlaInventario && stockActual < item.Cantidad && !product.PermiteStockNegativo {
			tx.Rollback()
			productoNombre := product.Nombre
			if variant != nil {
				productoNombre = variant.Nombre
			}
			return utils.ErrorWithCode(c, fiber.StatusBadRequest, "INSUFFICIENT_STOCK", fmt.Sprintf("stock insuficiente para %s (disponible: %d)", productoNombre, stockActual))
		}

		itemTotal := precioUnitario * float64(item.Cantidad)
		itemImpuesto := itemTotal * (impuestoPorcentaje / 100)

		items[i] = models.OrderItem{
			ProductoID:     item.ProductoID,
			VarianteID:     item.VariantID,
			NombreProducto: product.Nombre,
			SKU:            sku,
			VarianteInfo:   variantInfo,
			Cantidad:       item.Cantidad,
			PrecioUnitario: precioUnitario,
			Impuesto:       itemImpuesto,
			Total:          itemTotal + itemImpuesto,
		}

		subtotal += itemTotal
		impuesto += itemImpuesto

		if product.ControlaInventario {
			if variant != nil {
				stockAnterior := variant.Stock
				variant.Stock -= item.Cantidad
				if err := tx.Save(variant).Error; err != nil {
					tx.Rollback()
					return utils.ErrorWithCode(c, fiber.StatusInternalServerError, "STOCK_ERROR", "error al actualizar stock de variante")
				}
				registrarMovimientoStock(item.ProductoID, variant.ID, currentUser.ID, models.StockVenta, -item.Cantidad, stockAnterior, variant.Stock, "Venta POS")
			} else {
				stockAnterior := product.Stock
				product.Stock -= item.Cantidad
				if err := tx.Save(&product).Error; err != nil {
					tx.Rollback()
					return utils.ErrorWithCode(c, fiber.StatusInternalServerError, "STOCK_ERROR", "error al actualizar stock de producto")
				}
				registrarMovimientoStock(item.ProductoID, 0, currentUser.ID, models.StockVenta, -item.Cantidad, stockAnterior, product.Stock, "Venta POS")
			}
		}
	}

	total := subtotal + impuesto

	now := time.Now()
	order := models.Order{
		NumeroOrden:        numeroOrden,
		ClienteNombre:      "Cliente Mostrador",
		Estado:             "completado",
		MetodoPago:         input.MetodoPago,
		FechaPago:          &now,
		EstadoPago:         "pagado",
		Subtotal:           subtotal,
		Impuesto:           impuesto,
		ImpuestoPorcentaje: impuestoPorcentaje,
		Total:              total,
		VendedorID:         &currentUser.ID,
		Items:              items,
	}

	if result := tx.Create(&order); result.Error != nil {
		tx.Rollback()
		return utils.ErrorWithCode(c, fiber.StatusInternalServerError, "CREATE_ERROR", "error al crear venta")
	}

	if err := tx.Commit().Error; err != nil {
		return utils.ErrorWithCode(c, fiber.StatusInternalServerError, "DB_ERROR", "error al guardar venta")
	}

	cambio := input.Recibido - total
	if cambio < 0 {
		cambio = 0
	}

	database.DB.Preload("Cliente").Preload("Vendedor").Preload("Items").First(&order, order.ID)

	return utils.SuccessData(c, fiber.StatusCreated, fiber.Map{
		"order":  order.ToResponse(),
		"cambio": cambio,
		"total":  total,
	})
}

// UpdateOrder actualiza un pedido
// PUT /api/orders/:id
func UpdateOrder(c *fiber.Ctx) error {
	currentUser := middleware.GetUser(c)
	if currentUser == nil {
		return utils.ErrorWithCode(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "no autorizado")
	}

	if !currentUser.HasPermission("orders.update") {
		return utils.ErrorWithCode(c, fiber.StatusForbidden, "FORBIDDEN", "no tienes permiso para actualizar pedidos")
	}

	id, err := c.ParamsInt("id")
	if err != nil {
		return utils.ErrorWithCode(c, fiber.StatusBadRequest, "INVALID_ID", "ID de pedido inválido")
	}

	var order models.Order
	if result := database.DB.First(&order, id); result.Error != nil {
		return utils.ErrorWithCode(c, fiber.StatusNotFound, "NOT_FOUND", "pedido no encontrado")
	}

	var input struct {
		Estado string `json:"estado"`
		Notas  string `json:"notas"`
	}
	if err := c.BodyParser(&input); err != nil {
		return utils.ErrorWithCode(c, fiber.StatusBadRequest, "INVALID_BODY", "formato de datos inválido")
	}

	updates := make(map[string]interface{})
	if input.Estado != "" {
		updates["estado"] = input.Estado
	}
	if input.Notas != "" {
		updates["notas"] = input.Notas
	}

	if len(updates) > 0 {
		if result := database.DB.Model(&order).Updates(updates); result.Error != nil {
			return utils.ErrorWithCode(c, fiber.StatusInternalServerError, "UPDATE_ERROR", "error al actualizar pedido")
		}
	}

	database.DB.Preload("Cliente").Preload("Vendedor").Preload("Items").First(&order, order.ID)

	return utils.Success(c, fiber.StatusOK, "pedido actualizado exitosamente", order.ToResponse())
}

// CancelOrder cancela un pedido
// POST /api/orders/:id/cancel
func CancelOrder(c *fiber.Ctx) error {
	currentUser := middleware.GetUser(c)
	if currentUser == nil {
		return utils.ErrorWithCode(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "no autorizado")
	}

	if !currentUser.HasPermission("orders.delete") {
		return utils.ErrorWithCode(c, fiber.StatusForbidden, "FORBIDDEN", "no tienes permiso para cancelar pedidos")
	}

	id, err := c.ParamsInt("id")
	if err != nil {
		return utils.ErrorWithCode(c, fiber.StatusBadRequest, "INVALID_ID", "ID de pedido inválido")
	}

	var order models.Order
	if result := database.DB.Preload("Items").First(&order, id); result.Error != nil {
		return utils.ErrorWithCode(c, fiber.StatusNotFound, "NOT_FOUND", "pedido no encontrado")
	}

	if order.Estado == "cancelado" {
		return utils.ErrorWithCode(c, fiber.StatusBadRequest, "ALREADY_CANCELLED", "el pedido ya está cancelado")
	}

	// Restaurar stock
	tx := database.DB.Begin()
	for _, item := range order.Items {
		var product models.Product
		if result := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&product, item.ProductoID); result.Error == nil && product.ControlaInventario {
			if item.VarianteID != nil && *item.VarianteID > 0 {
				var variant models.ProductVariant
				if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&variant, *item.VarianteID).Error; err == nil {
					if err := tx.Model(&variant).Update("stock", variant.Stock+item.Cantidad).Error; err != nil {
						tx.Rollback()
						return utils.ErrorWithCode(c, fiber.StatusInternalServerError, "STOCK_ERROR", "error al restaurar stock de variante")
					}
				}
			} else {
				if err := tx.Model(&product).Update("stock", product.Stock+item.Cantidad).Error; err != nil {
					tx.Rollback()
					return utils.ErrorWithCode(c, fiber.StatusInternalServerError, "STOCK_ERROR", "error al restaurar stock")
				}
			}
		}
	}

	order.Estado = "cancelado"
	if err := tx.Save(&order).Error; err != nil {
		tx.Rollback()
		return utils.ErrorWithCode(c, fiber.StatusInternalServerError, "UPDATE_ERROR", "error al cancelar el pedido")
	}
	tx.Commit()

	return utils.Success(c, fiber.StatusOK, "pedido cancelado exitosamente", order.ToResponse())
}

func generateOrderNumber() string {
	now := time.Now()
	return fmt.Sprintf("NX-%d%02d%02d-%04d", now.Year(), now.Month(), now.Day(), now.Unix()%10000)
}

func registrarMovimientoStock(productoID uint, variantID uint, usuarioID uint, tipo string, cantidad int, stockAnterior int, stockNuevo int, motivo string) {
	movimiento := models.StockMovement{
		ProductoID:    productoID,
		VariantID:     nil,
		UsuarioID:     usuarioID,
		Tipo:          tipo,
		Cantidad:      cantidad,
		StockAnterior: stockAnterior,
		StockNuevo:    stockNuevo,
		Motivo:        motivo,
	}
	if variantID > 0 {
		movimiento.VariantID = &variantID
	}
	database.DB.Create(&movimiento)
}

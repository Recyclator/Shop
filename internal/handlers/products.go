package handlers

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/nexora/backend/internal/database"
	"github.com/nexora/backend/internal/middleware"
	"github.com/nexora/backend/internal/models"
	"github.com/nexora/backend/internal/utils"
	"gorm.io/gorm/clause"
)

// GetProducts obtiene la lista de productos con paginación y filtros
// GET /api/products
func GetProducts(c *fiber.Ctx) error {
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 20)
	search := c.Query("search", "")
	categoria := c.QueryInt("categoria", 0)
	estado := c.Query("estado", "")
	destacado := c.QueryBool("destacado", false)
	orden := c.Query("orden", "recientes") // recientes, precio_asc, precio_desc, nombre

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit

	query := database.DB.Model(&models.Product{}).Preload("Categoria").Preload("Tags").Preload("Variantes").Preload("Imagenes")

	// Filtros
	if search != "" {
		searchPattern := "%" + search + "%"
		query = query.Where("nombre ILIKE ? OR sku ILIKE ? OR descripcion ILIKE ?", searchPattern, searchPattern, searchPattern)
	}

	if categoria > 0 {
		query = query.Where("categoria_id = ?", categoria)
	}

	if estado != "" {
		query = query.Where("estado = ?", estado)
	}

	if destacado {
		query = query.Where("destacado = ?", true)
	}

	// Ordenamiento
	switch orden {
	case "precio_asc":
		query = query.Order("precio ASC")
	case "precio_desc":
		query = query.Order("precio DESC")
	case "nombre":
		query = query.Order("nombre ASC")
	default:
		query = query.Order("created_at DESC")
	}

	// Contar total
	var total int64
	query.Count(&total)

	// Obtener página
	var products []models.Product
	if err := query.Offset(offset).Limit(limit).Find(&products).Error; err != nil {
		return utils.ErrorWithCode(c, fiber.StatusInternalServerError, "DB_ERROR", "error al obtener productos")
	}

	// Convertir a respuestas
	responses := make([]models.ProductResponse, len(products))
	for i, p := range products {
		responses[i] = p.ToResponse()
	}

	return utils.Paginated(c, fiber.StatusOK, responses, page, limit, total)
}

// GetProductByID obtiene un producto por ID
// GET /api/products/:id
func GetProductByID(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return utils.ErrorWithCode(c, fiber.StatusBadRequest, "INVALID_ID", "ID de producto inválido")
	}

	cacheKey := fmt.Sprintf("nexora:catalog:product:%d", id)
	if database.RDB != nil {
		if val, err := database.RDB.Get(c.Context(), cacheKey).Result(); err == nil && val != "" {
			var cached models.ProductResponse
			if err := json.Unmarshal([]byte(val), &cached); err == nil {
				return utils.SuccessData(c, fiber.StatusOK, cached)
			}
		}
	}

	var product models.Product
	if result := database.DB.
		Preload("Categoria").
		Preload("Tags").
		Preload("Variantes").
		Preload("Imagenes").
		First(&product, id); result.Error != nil {
		return utils.ErrorWithCode(c, fiber.StatusNotFound, "NOT_FOUND", "producto no encontrado")
	}

	resp := product.ToResponse()
	if database.RDB != nil {
		if data, err := json.Marshal(resp); err == nil {
			database.RDB.Set(c.Context(), cacheKey, string(data), 10*time.Minute)
		}
	}

	return utils.SuccessData(c, fiber.StatusOK, resp)
}

// GetProductBySKU obtiene un producto por SKU
// GET /api/products/sku/:sku
func GetProductBySKU(c *fiber.Ctx) error {
	sku := c.Params("sku")

	var product models.Product
	if result := database.DB.
		Preload("Categoria").
		Preload("Tags").
		Preload("Variantes").
		Where("sku = ?", sku).First(&product); result.Error != nil {
		return utils.ErrorWithCode(c, fiber.StatusNotFound, "NOT_FOUND", "producto no encontrado")
	}

	return utils.SuccessData(c, fiber.StatusOK, product.ToResponse())
}

// GetProductByBarcode obtiene un producto por código de barras
// GET /api/products/barcode/:code
func GetProductByBarcode(c *fiber.Ctx) error {
	code := c.Params("code")

	var product models.Product
	if result := database.DB.
		Preload("Categoria").
		Preload("Tags").
		Preload("Variantes").
		Where("barcode = ? OR sku = ?", code, code).First(&product); result.Error != nil {
		return utils.ErrorWithCode(c, fiber.StatusNotFound, "NOT_FOUND", "producto no encontrado")
	}

	return utils.SuccessData(c, fiber.StatusOK, product.ToResponse())
}

// CreateProduct crea un nuevo producto
// POST /api/products
func CreateProduct(c *fiber.Ctx) error {
	currentUser := middleware.GetUser(c)
	if currentUser == nil {
		return utils.ErrorWithCode(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "no autorizado")
	}

	if !currentUser.HasPermission("products.create") {
		return utils.ErrorWithCode(c, fiber.StatusForbidden, "FORBIDDEN", "no tienes permiso para crear productos")
	}

	var input models.CreateProductInput
	if err := c.BodyParser(&input); err != nil {
		return utils.ErrorWithCode(c, fiber.StatusBadRequest, "INVALID_BODY", "formato de datos inválido")
	}

	// Validaciones
	if input.SKU == "" || input.Nombre == "" || input.Precio <= 0 {
		return utils.ErrorWithCode(c, fiber.StatusBadRequest, "VALIDATION_ERROR", "SKU, nombre y precio son requeridos")
	}

	// Verificar SKU único
	var existing models.Product
	if result := database.DB.Where("sku = ?", input.SKU).First(&existing); result.RowsAffected > 0 {
		return utils.ErrorWithCode(c, fiber.StatusConflict, "SKU_EXISTS", "el SKU ya existe")
	}

	estado := input.Estado
	if estado == "" {
		estado = "activo"
	}

	// Generar slug
	product := models.Product{
		SKU:                  input.SKU,
		Nombre:               input.Nombre,
		Slug:                 generateSlug(input.Nombre),
		Descripcion:          input.Descripcion,
		DescripcionCorta:     input.DescripcionCorta,
		Precio:               input.Precio,
		PrecioAnterior:       input.PrecioAnterior,
		Costo:                input.Costo,
		Stock:                input.Stock,
		StockMinimo:          input.StockMinimo,
		PermiteStockNegativo: input.PermiteStockNegativo,
		ControlaInventario:   input.ControlaInventario,
		CategoriaID:          input.CategoriaID,
		ImagenPrincipal:      input.ImagenPrincipal,
		Destacado:            input.Destacado,
		Nuevo:                input.Nuevo,
		TiempoEntregaDias:    input.TiempoEntregaDias,
		MetaTitulo:           input.MetaTitulo,
		MetaDescripcion:      input.MetaDescripcion,
		PalabrasClave:        input.PalabrasClave,
		TieneVariantes:       input.TieneVariantes || len(input.Variantes) > 0,
		Estado:               estado,
	}

	// Calcular porcentaje de descuento
	if product.PrecioAnterior > 0 && product.PrecioAnterior > product.Precio {
		product.PorcentajeDescuento = int(((product.PrecioAnterior - product.Precio) / product.PrecioAnterior) * 100)
	}

	// Guardar primero para obtener ID
	if result := database.DB.Create(&product); result.Error != nil {
		return utils.ErrorWithCode(c, fiber.StatusInternalServerError, "CREATE_ERROR", "error al crear producto")
	}

	// Generar código de barras usando el SKU
	product.Barcode = fmt.Sprintf("https://barcode.tec-it.com/barcode.ashx?data=%s&code=Code128&translate-esc=on", product.SKU)
	database.DB.Save(&product)

	// Crear variantes si se proporcionaron
	if len(input.Variantes) > 0 {
		totalVariantStock := 0
		for i, vi := range input.Variantes {
			varSKU := vi.SKU
			if varSKU == "" {
				varSKU = fmt.Sprintf("%s-V%d", product.SKU, i+1)
			}
			varName := vi.Nombre
			if varName == "" {
				varName = fmt.Sprintf("%s - Variante %d", product.Nombre, i+1)
			}
			stockMin := vi.StockMinimo
			if stockMin <= 0 {
				stockMin = product.StockMinimo
			}
			variant := models.ProductVariant{
				ProductoID:     product.ID,
				SKU:            varSKU,
				Barcode:        vi.Barcode,
				Nombre:         varName,
				PrecioOverride: vi.PrecioOverride,
				Stock:          vi.Stock,
				StockMinimo:    stockMin,
				Imagen:         vi.Imagen,
				Activa:         true,
				Orden:          i,
			}
			if err := database.DB.Create(&variant).Error; err == nil {
				for _, a := range vi.Atributos {
					attrID := a.AttributeID
					if attrID == 0 && a.AttributeName != "" {
						var existingAttr models.ProductAttribute
						if err := database.DB.Where("LOWER(nombre) = LOWER(?)", a.AttributeName).First(&existingAttr).Error; err == nil {
							attrID = existingAttr.ID
						} else {
							newAttr := models.ProductAttribute{
								CategoriaID: product.CategoriaID,
								Nombre:      a.AttributeName,
								Tipo:        "select",
								Activo:      true,
							}
							newAttr.SetValores([]string{a.Value})
							if err := database.DB.Create(&newAttr).Error; err == nil {
								attrID = newAttr.ID
							}
						}
					}
					if attrID > 0 {
						attrValue := models.VariantAttributeValue{
							VariantID:   variant.ID,
							AttributeID: attrID,
							Value:       a.Value,
						}
						database.DB.Create(&attrValue)
					}
				}
				totalVariantStock += vi.Stock
			}
		}
		product.TieneVariantes = true
		product.Stock = totalVariantStock
		database.DB.Save(&product)
	}

	// Asignar tags
	if len(input.Tags) > 0 {
		var tags []models.Tag
		database.DB.Find(&tags, input.Tags)
		database.DB.Model(&product).Association("Tags").Append(tags)
	}

	// Cargar relaciones
	database.DB.Preload("Categoria").Preload("Tags").Preload("Variantes.Atributos.Attribute").Preload("Imagenes").First(&product, product.ID)

	return utils.Success(c, fiber.StatusCreated, "producto creado exitosamente", product.ToResponse())
}

// UpdateProduct actualiza un producto
// PUT /api/products/:id
func UpdateProduct(c *fiber.Ctx) error {
	currentUser := middleware.GetUser(c)
	if currentUser == nil {
		return utils.ErrorWithCode(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "no autorizado")
	}

	if !currentUser.HasPermission("products.update") {
		return utils.ErrorWithCode(c, fiber.StatusForbidden, "FORBIDDEN", "no tienes permiso para actualizar productos")
	}

	id, err := c.ParamsInt("id")
	if err != nil {
		return utils.ErrorWithCode(c, fiber.StatusBadRequest, "INVALID_ID", "ID de producto inválido")
	}

	var product models.Product
	if result := database.DB.First(&product, id); result.Error != nil {
		return utils.ErrorWithCode(c, fiber.StatusNotFound, "NOT_FOUND", "producto no encontrado")
	}

	var input models.UpdateProductInput
	if err := c.BodyParser(&input); err != nil {
		return utils.ErrorWithCode(c, fiber.StatusBadRequest, "INVALID_BODY", "formato de datos inválido")
	}

	updates := make(map[string]interface{})

	if input.SKU != "" {
		updates["sku"] = input.SKU
	}
	if input.Nombre != "" {
		updates["nombre"] = input.Nombre
		updates["slug"] = generateSlug(input.Nombre)
	}
	if input.Descripcion != "" {
		updates["descripcion"] = input.Descripcion
	}
	if input.DescripcionCorta != "" {
		updates["descripcion_corta"] = input.DescripcionCorta
	}
	if input.Precio > 0 {
		updates["precio"] = input.Precio
		// Recalcular descuento
		if product.PrecioAnterior > 0 && input.Precio < product.PrecioAnterior {
			updates["porcentaje_descuento"] = int(((product.PrecioAnterior - input.Precio) / product.PrecioAnterior) * 100)
		}
	}
	if input.PrecioAnterior > 0 {
		updates["precio_anterior"] = input.PrecioAnterior
	}
	if input.Costo > 0 {
		updates["costo"] = input.Costo
	}
	if input.Stock != nil {
		updates["stock"] = *input.Stock
	}
	if input.StockMinimo > 0 {
		updates["stock_minimo"] = input.StockMinimo
	}
	if input.CategoriaID != nil {
		updates["categoria_id"] = *input.CategoriaID
	}
	if input.ImagenPrincipal != "" {
		updates["imagen_principal"] = input.ImagenPrincipal
	}
	if input.Estado != "" {
		updates["estado"] = input.Estado
	}
	if input.Destacado != nil {
		updates["destacado"] = *input.Destacado
	}
	if input.Nuevo != nil {
		updates["nuevo"] = *input.Nuevo
	}
	if input.TiempoEntregaDias > 0 {
		updates["tiempo_entrega_dias"] = input.TiempoEntregaDias
	}

	if len(updates) > 0 {
		if result := database.DB.Model(&product).Updates(updates); result.Error != nil {
			return utils.ErrorWithCode(c, fiber.StatusInternalServerError, "UPDATE_ERROR", "error al actualizar producto")
		}
	}

	// Actualizar tags si se proporcionan
	if input.Tags != nil {
		var tags []models.Tag
		database.DB.Find(&tags, input.Tags)
		database.DB.Model(&product).Association("Tags").Replace(tags)
	}

	// Recargar
	database.DB.Preload("Categoria").Preload("Tags").Preload("Variantes").Preload("Imagenes").First(&product, product.ID)

	database.InvalidateProductCache(product.ID)

	return utils.Success(c, fiber.StatusOK, "producto actualizado exitosamente", product.ToResponse())
}

// DeleteProduct elimina un producto
// DELETE /api/products/:id
func DeleteProduct(c *fiber.Ctx) error {
	currentUser := middleware.GetUser(c)
	if currentUser == nil {
		return utils.ErrorWithCode(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "no autorizado")
	}

	if !currentUser.HasPermission("products.delete") {
		return utils.ErrorWithCode(c, fiber.StatusForbidden, "FORBIDDEN", "no tienes permiso para eliminar productos")
	}

	id, err := c.ParamsInt("id")
	if err != nil {
		return utils.ErrorWithCode(c, fiber.StatusBadRequest, "INVALID_ID", "ID de producto inválido")
	}

	var product models.Product
	if result := database.DB.First(&product, id); result.Error != nil {
		return utils.ErrorWithCode(c, fiber.StatusNotFound, "NOT_FOUND", "producto no encontrado")
	}

	// Soft delete
	if result := database.DB.Delete(&product); result.Error != nil {
		return utils.ErrorWithCode(c, fiber.StatusInternalServerError, "DELETE_ERROR", "error al eliminar producto")
	}

	database.InvalidateProductCache(product.ID)

	return utils.SuccessMessage(c, fiber.StatusOK, "producto eliminado exitosamente")
}

// SearchProducts busca productos
// GET /api/products/search?q=...
func SearchProducts(c *fiber.Ctx) error {
	q := strings.TrimSpace(c.Query("q", ""))
	limit := c.QueryInt("limit", 10)

	if q == "" {
		return utils.ErrorWithCode(c, fiber.StatusBadRequest, "EMPTY_QUERY", "parámetro de búsqueda requerido")
	}

	searchPattern := "%" + q + "%"
	var products []models.Product

	if err := database.DB.
		Where("nombre ILIKE ? OR sku ILIKE ? OR descripcion ILIKE ?", searchPattern, searchPattern, searchPattern).
		Where("estado = ?", "activo").
		Preload("Categoria").
		Limit(limit).
		Find(&products).Error; err != nil {
		return utils.ErrorWithCode(c, fiber.StatusInternalServerError, "DB_ERROR", "error al buscar productos")
	}

	responses := make([]models.ProductResponse, len(products))
	for i, p := range products {
		responses[i] = p.ToResponse()
	}

	return utils.SuccessData(c, fiber.StatusOK, responses)
}

// GetFeaturedProducts obtiene productos destacados
// GET /api/products/featured
func GetFeaturedProducts(c *fiber.Ctx) error {
	limit := c.QueryInt("limit", 10)

	var products []models.Product
	if err := database.DB.
		Where("destacado = ? AND estado = ?", true, "activo").
		Preload("Categoria").
		Limit(limit).
		Order("created_at DESC").
		Find(&products).Error; err != nil {
		return utils.ErrorWithCode(c, fiber.StatusInternalServerError, "DB_ERROR", "error al obtener productos")
	}

	responses := make([]models.ProductResponse, len(products))
	for i, p := range products {
		responses[i] = p.ToResponse()
	}

	return utils.SuccessData(c, fiber.StatusOK, responses)
}

// GetProductsByCategory obtiene productos por categoría
// GET /api/products/category/:id
func GetProductsByCategory(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return utils.ErrorWithCode(c, fiber.StatusBadRequest, "INVALID_ID", "ID de categoría inválido")
	}

	limit := c.QueryInt("limit", 20)

	var products []models.Product
	if err := database.DB.
		Where("categoria_id = ? AND estado = ?", id, "activo").
		Preload("Categoria").
		Limit(limit).
		Order("created_at DESC").
		Find(&products).Error; err != nil {
		return utils.ErrorWithCode(c, fiber.StatusInternalServerError, "DB_ERROR", "error al obtener productos")
	}

	responses := make([]models.ProductResponse, len(products))
	for i, p := range products {
		responses[i] = p.ToResponse()
	}

	return utils.SuccessData(c, fiber.StatusOK, responses)
}

// GenerateProductQR genera el código QR de un producto
// POST /api/products/:id/qr
func GenerateProductQR(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return utils.ErrorWithCode(c, fiber.StatusBadRequest, "INVALID_ID", "ID de producto inválido")
	}

	var product models.Product
	if result := database.DB.First(&product, id); result.Error != nil {
		return utils.ErrorWithCode(c, fiber.StatusNotFound, "NOT_FOUND", "producto no encontrado")
	}

	barcode := fmt.Sprintf("https://barcode.tec-it.com/barcode.ashx?data=%s&code=Code128&translate-esc=on", product.SKU)

	product.Barcode = barcode
	database.DB.Save(&product)

	database.InvalidateProductCache(product.ID)

	return utils.SuccessData(c, fiber.StatusOK, fiber.Map{
		"barcode":  barcode,
		"producto": product.ToResponse(),
	})
}

// Helpers
func generateSlug(nombre string) string {
	slug := strings.ToLower(nombre)
	slug = strings.ReplaceAll(slug, " ", "-")
	slug = strings.ReplaceAll(slug, "_", "-")

	var result strings.Builder
	for _, r := range slug {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			result.WriteRune(r)
		}
	}

	slug = result.String()
	for strings.Contains(slug, "--") {
		slug = strings.ReplaceAll(slug, "--", "-")
	}
	return strings.Trim(slug, "-")
}

// AdjustProductStock ajusta el stock de un producto
// POST /api/products/:id/stock/adjust
func AdjustProductStock(c *fiber.Ctx) error {
	currentUser := middleware.GetUser(c)
	if currentUser == nil {
		return utils.ErrorWithCode(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "no autorizado")
	}

	if !currentUser.HasPermission("products.update") {
		return utils.ErrorWithCode(c, fiber.StatusForbidden, "FORBIDDEN", "no tienes permiso para ajustar stock")
	}

	id, err := c.ParamsInt("id")
	if err != nil {
		return utils.ErrorWithCode(c, fiber.StatusBadRequest, "INVALID_ID", "ID de producto inválido")
	}

	var input models.StockAdjustInput
	if err := c.BodyParser(&input); err != nil {
		return utils.ErrorWithCode(c, fiber.StatusBadRequest, "INVALID_BODY", "formato de datos inválido")
	}

	if input.Cantidad <= 0 || input.Motivo == "" {
		return utils.ErrorWithCode(c, fiber.StatusBadRequest, "VALIDATION_ERROR", "cantidad y motivo son requeridos")
	}

	tx := database.DB.Begin()

	var product models.Product
	if result := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&product, id); result.Error != nil {
		tx.Rollback()
		return utils.ErrorWithCode(c, fiber.StatusNotFound, "NOT_FOUND", "producto no encontrado")
	}

	if !product.ControlaInventario {
		tx.Rollback()
		return utils.ErrorWithCode(c, fiber.StatusBadRequest, "NO_INVENTORY", "este producto no tiene control de inventario")
	}

	stockAnterior := product.Stock
	var stockNuevo int
	var variant *models.ProductVariant
	var variantID uint

	if input.VariantID != nil {
		variant = &models.ProductVariant{}
		if result := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(variant, *input.VariantID); result.Error != nil {
			tx.Rollback()
			return utils.ErrorWithCode(c, fiber.StatusNotFound, "NOT_FOUND", "variante no encontrada")
		}
		if variant.ProductoID != product.ID {
			tx.Rollback()
			return utils.ErrorWithCode(c, fiber.StatusBadRequest, "INVALID_VARIANT", "la variante no pertenece a este producto")
		}
		variantID = variant.ID
		stockAnterior = variant.Stock
	}

	var tipo string
	switch input.Tipo {
	case "entrada":
		tipo = models.StockEntrada
		stockNuevo = stockAnterior + input.Cantidad
	case "salida":
		tipo = models.StockSalida
		stockNuevo = stockAnterior - input.Cantidad
		if stockNuevo < 0 && !product.PermiteStockNegativo {
			tx.Rollback()
			return utils.ErrorWithCode(c, fiber.StatusBadRequest, "INSUFFICIENT_STOCK", "stock insuficiente")
		}
	case "ajuste":
		tipo = models.StockCorreccion
		stockNuevo = input.Cantidad
	default:
		tx.Rollback()
		return utils.ErrorWithCode(c, fiber.StatusBadRequest, "INVALID_TYPE", "tipo de ajuste inválido")
	}

	if variant != nil {
		variant.Stock = stockNuevo
		tx.Save(variant)
	} else {
		product.Stock = stockNuevo
		tx.Save(&product)
	}

	movimiento := models.StockMovement{
		ProductoID:    product.ID,
		VariantID:     &variantID,
		UsuarioID:     currentUser.ID,
		Tipo:          tipo,
		Cantidad:      input.Cantidad,
		StockAnterior: stockAnterior,
		StockNuevo:    stockNuevo,
		Motivo:        input.Motivo,
	}
	tx.Create(&movimiento)

	if err := tx.Commit().Error; err != nil {
		return utils.ErrorWithCode(c, fiber.StatusInternalServerError, "DB_ERROR", "error al guardar ajuste de stock")
	}

	database.InvalidateProductCache(product.ID)

	return utils.SuccessData(c, fiber.StatusOK, fiber.Map{
		"producto_id":    product.ID,
		"variant_id":     variantID,
		"stock_anterior": stockAnterior,
		"stock_nuevo":    stockNuevo,
		"movimiento":     movimiento,
	})
}

// GetStockAlerts obtiene productos con stock bajo mínimo
// GET /api/dashboard/stock-alerts
func GetStockAlerts(c *fiber.Ctx) error {
	var products []models.Product
	if err := database.DB.
		Where("controla_inventario = ? AND stock <= stock_minimo AND stock_minimo > 0 AND estado = ?", true, "activo").
		Find(&products).Error; err != nil {
		return utils.ErrorWithCode(c, fiber.StatusInternalServerError, "DB_ERROR", "error al obtener alertas")
	}

	alerts := make([]models.StockAlertResponse, 0, len(products))
	for _, p := range products {
		alert := models.StockAlertResponse{
			ID:          p.ID,
			Nombre:      p.Nombre,
			SKU:         p.SKU,
			StockActual: p.Stock,
			StockMinimo: p.StockMinimo,
			Deficit:     p.StockMinimo - p.Stock,
		}
		alerts = append(alerts, alert)
	}

	var variants []models.ProductVariant
	database.DB.
		Preload("Producto").
		Where("stock <= (SELECT stock_minimo FROM products WHERE id = product_variants.producto_id AND controla_inventario = true AND stock_minimo > 0)").
		Or("stock = 0").
		Find(&variants)

	for _, v := range variants {
		if v.Producto.ID != 0 {
			deficit := v.Producto.StockMinimo - v.Stock
			if deficit > 0 {
				variantID := v.ID
				alerts = append(alerts, models.StockAlertResponse{
					ID:          v.Producto.ID,
					Nombre:      v.Producto.Nombre,
					SKU:         v.SKU,
					StockActual: v.Stock,
					StockMinimo: v.Producto.StockMinimo,
					Deficit:     deficit,
					VariantID:   &variantID,
					VariantSKU:  v.SKU,
				})
			}
		}
	}

	return utils.SuccessData(c, fiber.StatusOK, alerts)
}

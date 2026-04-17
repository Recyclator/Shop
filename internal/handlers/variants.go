package handlers

import (
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/nexora/backend/internal/database"
	"github.com/nexora/backend/internal/middleware"
	"github.com/nexora/backend/internal/models"
	"github.com/nexora/backend/internal/utils"
)

func CreateVariant(c *fiber.Ctx) error {
	currentUser := middleware.GetUser(c)
	if currentUser == nil {
		return utils.ErrorWithCode(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "no autorizado")
	}

	if !currentUser.HasPermission("products.create") {
		return utils.ErrorWithCode(c, fiber.StatusForbidden, "FORBIDDEN", "no tienes permiso")
	}

	productID, err := c.ParamsInt("id")
	if err != nil {
		return utils.ErrorWithCode(c, fiber.StatusBadRequest, "INVALID_ID", "ID de producto inválido")
	}

	var product models.Product
	if result := database.DB.First(&product, productID); result.Error != nil {
		return utils.ErrorWithCode(c, fiber.StatusNotFound, "NOT_FOUND", "producto no encontrado")
	}

	var input models.CreateVariantInput
	if err := c.BodyParser(&input); err != nil {
		return utils.ErrorWithCode(c, fiber.StatusBadRequest, "INVALID_BODY", "formato inválido")
	}

	if len(input.Atributos) == 0 {
		return utils.ErrorWithCode(c, fiber.StatusBadRequest, "NO_ATTRIBUTES", "debe especificar al menos un atributo")
	}

	attrMap := make(map[uint]string)
	var attrIDs []uint
	for _, a := range input.Atributos {
		attrIDs = append(attrIDs, a.AttributeID)
	}
	var attrs []models.ProductAttribute
	database.DB.Where("id IN ?", attrIDs).Find(&attrs)
	for _, a := range attrs {
		attrMap[a.ID] = a.Nombre
	}
	if len(attrs) != len(attrIDs) {
		return utils.ErrorWithCode(c, fiber.StatusBadRequest, "INVALID_ATTRIBUTES", "algunos atributos no existen")
	}

	sku := input.SKU
	if sku == "" {
		var skuParts []string
		skuParts = append(skuParts, product.SKU)
		for _, a := range input.Atributos {
			skuParts = append(skuParts, strings.ToUpper(a.Value))
		}
		sku = strings.Join(skuParts, "-")
	}

	nombre := input.Nombre
	if nombre == "" {
		var nombreParts []string
		for _, a := range input.Atributos {
			attrName := attrMap[a.AttributeID]
			nombreParts = append(nombreParts, fmt.Sprintf("%s: %s", attrName, a.Value))
		}
		nombre = product.Nombre + " (" + strings.Join(nombreParts, ", ") + ")"
	}

	variant := models.ProductVariant{
		ProductoID:     product.ID,
		SKU:            sku,
		Nombre:         nombre,
		PrecioOverride: input.PrecioOverride,
		Stock:          input.Stock,
		StockMinimo:    input.StockMinimo,
		Imagen:         input.Imagen,
		Activa:         true,
	}

	if result := database.DB.Create(&variant); result.Error != nil {
		return utils.ErrorWithCode(c, fiber.StatusInternalServerError, "CREATE_ERROR", "error al crear variante")
	}

	for _, a := range input.Atributos {
		attrValue := models.VariantAttributeValue{
			VariantID:   variant.ID,
			AttributeID: a.AttributeID,
			Value:       a.Value,
		}
		database.DB.Create(&attrValue)
	}

	database.DB.Model(&product).Update("tiene_variantes", true)

	database.DB.Preload("Atributos").First(&variant, variant.ID)

	return utils.Success(c, fiber.StatusCreated, "variante creada", variant.ToResponse(product.Precio))
}

func GetVariantsByProduct(c *fiber.Ctx) error {
	productID, err := c.ParamsInt("id")
	if err != nil {
		return utils.ErrorWithCode(c, fiber.StatusBadRequest, "INVALID_ID", "ID de producto inválido")
	}

	var product models.Product
	if result := database.DB.First(&product, productID); result.Error != nil {
		return utils.ErrorWithCode(c, fiber.StatusNotFound, "NOT_FOUND", "producto no encontrado")
	}

	var variantes []models.ProductVariant
	if err := database.DB.Where("producto_id = ?", productID).
		Preload("Atributos").
		Order("orden ASC").
		Find(&variantes).Error; err != nil {
		return utils.ErrorWithCode(c, fiber.StatusInternalServerError, "DB_ERROR", "error al obtener variantes")
	}

	responses := make([]models.VariantResponse, len(variantes))
	for i, v := range variantes {
		responses[i] = v.ToResponse(product.Precio)
	}

	return utils.SuccessData(c, fiber.StatusOK, responses)
}

func UpdateVariant(c *fiber.Ctx) error {
	currentUser := middleware.GetUser(c)
	if currentUser == nil {
		return utils.ErrorWithCode(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "no autorizado")
	}

	if !currentUser.HasPermission("products.update") {
		return utils.ErrorWithCode(c, fiber.StatusForbidden, "FORBIDDEN", "no tienes permiso")
	}

	productID, err := c.ParamsInt("id")
	if err != nil {
		return utils.ErrorWithCode(c, fiber.StatusBadRequest, "INVALID_ID", "ID de producto inválido")
	}

	varID, err := c.ParamsInt("var_id")
	if err != nil {
		return utils.ErrorWithCode(c, fiber.StatusBadRequest, "INVALID_ID", "ID de variante inválido")
	}

	var product models.Product
	var variant models.ProductVariant
	if result := database.DB.Where("id = ? AND producto_id = ?", varID, productID).
		First(&variant); result.Error != nil {
		return utils.ErrorWithCode(c, fiber.StatusNotFound, "NOT_FOUND", "variante no encontrada")
	}

	var input models.UpdateVariantInput
	if err := c.BodyParser(&input); err != nil {
		return utils.ErrorWithCode(c, fiber.StatusBadRequest, "INVALID_BODY", "formato inválido")
	}

	updates := make(map[string]interface{})
	if input.Nombre != "" {
		updates["nombre"] = input.Nombre
	}
	if input.PrecioOverride != nil {
		if *input.PrecioOverride == 0 {
			updates["precio_override"] = nil
		} else {
			updates["precio_override"] = *input.PrecioOverride
		}
	}
	if input.Stock != nil {
		updates["stock"] = *input.Stock
	}
	if input.StockMinimo > 0 {
		updates["stock_minimo"] = input.StockMinimo
	}
	if input.Imagen != "" {
		updates["imagen"] = input.Imagen
	}
	if input.Activa != nil {
		updates["activa"] = *input.Activa
	}

	if len(updates) > 0 {
		if result := database.DB.Model(&variant).Updates(updates); result.Error != nil {
			return utils.ErrorWithCode(c, fiber.StatusInternalServerError, "UPDATE_ERROR", "error al actualizar")
		}
	}

	database.DB.Preload("Atributos").First(&variant, variant.ID)
	database.DB.First(&product, productID)

	return utils.Success(c, fiber.StatusOK, "variante actualizada", variant.ToResponse(product.Precio))
}

func DeleteVariant(c *fiber.Ctx) error {
	currentUser := middleware.GetUser(c)
	if currentUser == nil {
		return utils.ErrorWithCode(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "no autorizado")
	}

	if !currentUser.HasPermission("products.delete") {
		return utils.ErrorWithCode(c, fiber.StatusForbidden, "FORBIDDEN", "no tienes permiso")
	}

	productID, err := c.ParamsInt("id")
	if err != nil {
		return utils.ErrorWithCode(c, fiber.StatusBadRequest, "INVALID_ID", "ID de producto inválido")
	}

	varID, err := c.ParamsInt("var_id")
	if err != nil {
		return utils.ErrorWithCode(c, fiber.StatusBadRequest, "INVALID_ID", "ID de variante inválido")
	}

	var variant models.ProductVariant
	if result := database.DB.Where("id = ? AND producto_id = ?", varID, productID).
		First(&variant); result.Error != nil {
		return utils.ErrorWithCode(c, fiber.StatusNotFound, "NOT_FOUND", "variante no encontrada")
	}

	if result := database.DB.Delete(&variant); result.Error != nil {
		return utils.ErrorWithCode(c, fiber.StatusInternalServerError, "DELETE_ERROR", "error al eliminar variante")
	}

	return utils.Success(c, fiber.StatusOK, "variante eliminada", nil)
}

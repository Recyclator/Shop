package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/nexora/backend/internal/database"
	"github.com/nexora/backend/internal/middleware"
	"github.com/nexora/backend/internal/models"
	"github.com/nexora/backend/internal/utils"
)

func CreateAttribute(c *fiber.Ctx) error {
	currentUser := middleware.GetUser(c)
	if currentUser == nil {
		return utils.ErrorWithCode(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "no autorizado")
	}

	if !currentUser.HasPermission("products.create") {
		return utils.ErrorWithCode(c, fiber.StatusForbidden, "FORBIDDEN", "no tienes permiso")
	}

	categoriaID, err := c.ParamsInt("id")
	if err != nil {
		return utils.ErrorWithCode(c, fiber.StatusBadRequest, "INVALID_ID", "ID de categoría inválido")
	}

	var category models.Category
	if result := database.DB.First(&category, categoriaID); result.Error != nil {
		return utils.ErrorWithCode(c, fiber.StatusNotFound, "NOT_FOUND", "categoría no encontrada")
	}

	var input models.CreateAttributeInput
	if err := c.BodyParser(&input); err != nil {
		return utils.ErrorWithCode(c, fiber.StatusBadRequest, "INVALID_BODY", "formato inválido")
	}

	if input.Nombre == "" {
		return utils.ErrorWithCode(c, fiber.StatusBadRequest, "INVALID_NAME", "el nombre es requerido")
	}

	attr := models.ProductAttribute{
		CategoriaID: &category.ID,
		Nombre:      input.Nombre,
		Tipo:        "select",
		Activo:      true,
	}

	// Support both "valores" and "valores_opciones" from input
	valores := input.Valores
	if len(valores) == 0 && len(input.ValoresOpciones) > 0 {
		valores = input.ValoresOpciones
	}
	attr.SetValores(valores)

	if result := database.DB.Create(&attr); result.Error != nil {
		return utils.ErrorWithCode(c, fiber.StatusInternalServerError, "CREATE_ERROR", "error al crear atributo")
	}

	return utils.Success(c, fiber.StatusCreated, "atributo creado", attr)
}

func GetAttributesByCategory(c *fiber.Ctx) error {
	categoriaID, err := c.ParamsInt("id")
	if err != nil {
		return utils.ErrorWithCode(c, fiber.StatusBadRequest, "INVALID_ID", "ID de categoría inválido")
	}

	var atributos []models.ProductAttribute
	if err := database.DB.Where("categoria_id = ? AND activo = ?", categoriaID, true).
		Order("orden ASC").Find(&atributos).Error; err != nil {
		return utils.ErrorWithCode(c, fiber.StatusInternalServerError, "DB_ERROR", "error al obtener atributos")
	}

	type AttrResponse struct {
		ID      uint     `json:"id"`
		Nombre  string   `json:"nombre"`
		Tipo    string   `json:"tipo"`
		Valores []string `json:"valores"`
		Orden   int      `json:"orden"`
	}

	responses := make([]AttrResponse, len(atributos))
	for i, attr := range atributos {
		responses[i] = AttrResponse{
			ID:      attr.ID,
			Nombre:  attr.Nombre,
			Tipo:    attr.Tipo,
			Valores: attr.GetValores(),
			Orden:   attr.Orden,
		}
	}

	return utils.SuccessData(c, fiber.StatusOK, responses)
}

func UpdateAttribute(c *fiber.Ctx) error {
	currentUser := middleware.GetUser(c)
	if currentUser == nil {
		return utils.ErrorWithCode(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "no autorizado")
	}

	if !currentUser.HasPermission("products.update") {
		return utils.ErrorWithCode(c, fiber.StatusForbidden, "FORBIDDEN", "no tienes permiso")
	}

	id, err := c.ParamsInt("id")
	if err != nil {
		return utils.ErrorWithCode(c, fiber.StatusBadRequest, "INVALID_ID", "ID de atributo inválido")
	}

	var attr models.ProductAttribute
	if result := database.DB.First(&attr, id); result.Error != nil {
		return utils.ErrorWithCode(c, fiber.StatusNotFound, "NOT_FOUND", "atributo no encontrado")
	}

	var input models.UpdateAttributeInput
	if err := c.BodyParser(&input); err != nil {
		return utils.ErrorWithCode(c, fiber.StatusBadRequest, "INVALID_BODY", "formato inválido")
	}

	updates := make(map[string]interface{})
	if input.Nombre != "" {
		updates["nombre"] = input.Nombre
	}
	if input.Tipo != "" {
		updates["tipo"] = input.Tipo
	}
	if len(input.Valores) > 0 {
		attr.SetValores(input.Valores)
		updates["valores"] = attr.Valores
	}
	if input.Activo != nil {
		updates["activo"] = *input.Activo
	}

	if len(updates) > 0 {
		if result := database.DB.Model(&attr).Updates(updates); result.Error != nil {
			return utils.ErrorWithCode(c, fiber.StatusInternalServerError, "UPDATE_ERROR", "error al actualizar")
		}
	}

	database.DB.First(&attr, id)
	return utils.Success(c, fiber.StatusOK, "atributo actualizado", attr)
}

func DeleteAttribute(c *fiber.Ctx) error {
	currentUser := middleware.GetUser(c)
	if currentUser == nil {
		return utils.ErrorWithCode(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "no autorizado")
	}

	if !currentUser.HasPermission("products.delete") {
		return utils.ErrorWithCode(c, fiber.StatusForbidden, "FORBIDDEN", "no tienes permiso")
	}

	id, err := c.ParamsInt("id")
	if err != nil {
		return utils.ErrorWithCode(c, fiber.StatusBadRequest, "INVALID_ID", "ID de atributo inválido")
	}

	var attr models.ProductAttribute
	if result := database.DB.First(&attr, id); result.Error != nil {
		return utils.ErrorWithCode(c, fiber.StatusNotFound, "NOT_FOUND", "atributo no encontrado")
	}

	var count int64
	database.DB.Model(&models.VariantAttributeValue{}).Where("attribute_id = ?", id).Count(&count)
	if count > 0 {
		return utils.ErrorWithCode(c, fiber.StatusBadRequest, "IN_USE", "el atributo está en uso por variantes")
	}

	if result := database.DB.Delete(&attr); result.Error != nil {
		return utils.ErrorWithCode(c, fiber.StatusInternalServerError, "DELETE_ERROR", "error al eliminar")
	}

	return utils.Success(c, fiber.StatusOK, "atributo eliminado", nil)
}

func GetGlobalAttributes(c *fiber.Ctx) error {
	var atributos []models.ProductAttribute
	if err := database.DB.Where("categoria_id IS NULL AND activo = ?", true).
		Order("orden ASC").Find(&atributos).Error; err != nil {
		return utils.ErrorWithCode(c, fiber.StatusInternalServerError, "DB_ERROR", "error al obtener atributos")
	}

	type AttrResponse struct {
		ID      uint     `json:"id"`
		Nombre  string   `json:"nombre"`
		Tipo    string   `json:"tipo"`
		Valores []string `json:"valores"`
	}

	responses := make([]AttrResponse, len(atributos))
	for i, attr := range atributos {
		responses[i] = AttrResponse{
			ID:      attr.ID,
			Nombre:  attr.Nombre,
			Tipo:    attr.Tipo,
			Valores: attr.GetValores(),
		}
	}

	return utils.SuccessData(c, fiber.StatusOK, responses)
}

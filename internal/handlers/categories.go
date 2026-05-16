package handlers

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/nexora/backend/internal/database"
	"github.com/nexora/backend/internal/middleware"
	"github.com/nexora/backend/internal/models"
	"github.com/nexora/backend/internal/utils"
)

// GetCategories obtener todas las categorías
// GET /api/categories
func GetCategories(c *fiber.Ctx) error {
	var categories []models.Category
	if err := database.DB.
		Preload("Padre").
		Preload("Hijos").
		Order("orden ASC, nombre ASC").
		Find(&categories).Error; err != nil {
		return utils.ErrorWithCode(c, fiber.StatusInternalServerError, "DB_ERROR", "error al obtener categorías")
	}

	type Response struct {
		ID          uint   `json:"id"`
		Nombre      string `json:"nombre"`
		Slug        string `json:"slug"`
		Descripcion string `json:"descripcion"`
		Imagen      string `json:"imagen"`
		PadreID     *uint  `json:"padre_id"`
		Orden       int    `json:"orden"`
		Activa      bool   `json:"activa"`
		MostrarMenu bool   `json:"mostrar_menu"`
	}

	responses := make([]Response, len(categories))
	for i, cat := range categories {
		responses[i] = Response{
			ID:          cat.ID,
			Nombre:      cat.Nombre,
			Slug:        cat.Slug,
			Descripcion: cat.Descripcion,
			Imagen:      cat.Imagen,
			PadreID:     cat.PadreID,
			Orden:       cat.Orden,
			Activa:      cat.Activa,
			MostrarMenu: cat.MostrarMenu,
		}
	}

	return utils.SuccessData(c, fiber.StatusOK, responses)
}

// GetCategoryByID obtener categoría por ID
// GET /api/categories/:id
func GetCategoryByID(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return utils.ErrorWithCode(c, fiber.StatusBadRequest, "INVALID_ID", "ID de categoría inválido")
	}

	var category models.Category
	if result := database.DB.
		Preload("Padre").
		Preload("Hijos").
		Preload("Productos").
		First(&category, id); result.Error != nil {
		return utils.ErrorWithCode(c, fiber.StatusNotFound, "NOT_FOUND", "categoría no encontrada")
	}

	return utils.SuccessData(c, fiber.StatusOK, category)
}

// CreateCategory crear nueva categoría
// POST /api/categories
func CreateCategory(c *fiber.Ctx) error {
	currentUser := middleware.GetUser(c)
	if currentUser == nil {
		return utils.ErrorWithCode(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "no autorizado")
	}

	if !currentUser.HasPermission("categories.create") {
		return utils.ErrorWithCode(c, fiber.StatusForbidden, "FORBIDDEN", "no tienes permiso para crear categorías")
	}

	var input models.CreateCategoryInput
	if err := c.BodyParser(&input); err != nil {
		return utils.ErrorWithCode(c, fiber.StatusBadRequest, "INVALID_BODY", "formato de datos inválido")
	}

	if input.Nombre == "" {
		return utils.ErrorWithCode(c, fiber.StatusBadRequest, "VALIDATION_ERROR", "el nombre es requerido")
	}

	// Verificar que no exista una categoría con el mismo nombre
	slug := generateSlug(input.Nombre)
	var existing models.Category
	if result := database.DB.Where("slug = ?", slug).First(&existing); result.RowsAffected > 0 {
		return utils.ErrorWithCode(c, fiber.StatusConflict, "CATEGORY_EXISTS", "ya existe una categoría con ese nombre")
	}

	category := models.Category{
		Nombre:      input.Nombre,
		Slug:        slug,
		Descripcion: input.Descripcion,
		Imagen:      input.Imagen,
		PadreID:     input.PadreID,
		Orden:       input.Orden,
		Activa:      input.Activa,
		MostrarMenu: input.MostrarMenu,
	}
	if category.Orden == 0 {
		category.Orden = 1
	}

	if result := database.DB.Create(&category); result.Error != nil {
		if strings.Contains(result.Error.Error(), "duplicate") || strings.Contains(result.Error.Error(), "unique") {
			return utils.ErrorWithCode(c, fiber.StatusConflict, "CATEGORY_EXISTS", "ya existe una categoría con ese nombre")
		}
		return utils.ErrorWithCode(c, fiber.StatusInternalServerError, "CREATE_ERROR", "error al crear categoría: "+result.Error.Error())
	}

	return utils.Success(c, fiber.StatusCreated, "categoría creada exitosamente", category)
}

// UpdateCategory actualizar categoría
// PUT /api/categories/:id
func UpdateCategory(c *fiber.Ctx) error {
	currentUser := middleware.GetUser(c)
	if currentUser == nil {
		return utils.ErrorWithCode(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "no autorizado")
	}

	if !currentUser.HasPermission("categories.update") {
		return utils.ErrorWithCode(c, fiber.StatusForbidden, "FORBIDDEN", "no tienes permiso para actualizar categorías")
	}

	id, err := c.ParamsInt("id")
	if err != nil {
		return utils.ErrorWithCode(c, fiber.StatusBadRequest, "INVALID_ID", "ID de categoría inválido")
	}

	var category models.Category
	if result := database.DB.First(&category, id); result.Error != nil {
		return utils.ErrorWithCode(c, fiber.StatusNotFound, "NOT_FOUND", "categoría no encontrada")
	}

	var input models.UpdateCategoryInput
	if err := c.BodyParser(&input); err != nil {
		return utils.ErrorWithCode(c, fiber.StatusBadRequest, "INVALID_BODY", "formato de datos inválido")
	}

	updates := make(map[string]interface{})

	if input.Nombre != "" {
		updates["nombre"] = input.Nombre
		updates["slug"] = generateSlug(input.Nombre)
	}
	if input.Descripcion != "" {
		updates["descripcion"] = input.Descripcion
	}
	if input.Imagen != "" {
		updates["imagen"] = input.Imagen
	}
	if input.PadreID != nil {
		if *input.PadreID == category.ID {
			return utils.ErrorWithCode(c, fiber.StatusBadRequest, "INVALID_PARENT", "una categoría no puede ser padre de sí misma")
		}
		updates["padre_id"] = input.PadreID
	}
	if input.Orden > 0 {
		updates["orden"] = input.Orden
	}
	if input.Activa != nil {
		updates["activa"] = *input.Activa
	}
	if input.MostrarMenu != nil {
		updates["mostrar_menu"] = *input.MostrarMenu
	}

	if len(updates) > 0 {
		if result := database.DB.Model(&category).Updates(updates); result.Error != nil {
			if strings.Contains(result.Error.Error(), "duplicate") || strings.Contains(result.Error.Error(), "unique") {
				return utils.ErrorWithCode(c, fiber.StatusConflict, "CATEGORY_EXISTS", "ya existe una categoría con ese nombre")
			}
			return utils.ErrorWithCode(c, fiber.StatusInternalServerError, "UPDATE_ERROR", "error al actualizar categoría: "+result.Error.Error())
		}
	}

	database.DB.Preload("Padre").Preload("Hijos").First(&category, id)

	return utils.Success(c, fiber.StatusOK, "categoría actualizada exitosamente", category)
}

// DeleteCategory eliminar categoría
// DELETE /api/categories/:id
func DeleteCategory(c *fiber.Ctx) error {
	currentUser := middleware.GetUser(c)
	if currentUser == nil {
		return utils.ErrorWithCode(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "no autorizado")
	}

	if !currentUser.HasPermission("categories.delete") {
		return utils.ErrorWithCode(c, fiber.StatusForbidden, "FORBIDDEN", "no tienes permiso para eliminar categorías")
	}

	id, err := c.ParamsInt("id")
	if err != nil {
		return utils.ErrorWithCode(c, fiber.StatusBadRequest, "INVALID_ID", "ID de categoría inválido")
	}

	var category models.Category
	if result := database.DB.First(&category, id); result.Error != nil {
		return utils.ErrorWithCode(c, fiber.StatusNotFound, "NOT_FOUND", "categoría no encontrada")
	}

	var count int64
	database.DB.Model(&models.Product{}).Where("categoria_id = ?", id).Count(&count)
	if count > 0 {
		return utils.ErrorWithCode(c, fiber.StatusConflict, "HAS_PRODUCTS", "la categoría tiene productos asociados")
	}

	var childrenCount int64
	database.DB.Model(&models.Category{}).Where("padre_id = ?", id).Count(&childrenCount)
	if childrenCount > 0 {
		return utils.ErrorWithCode(c, fiber.StatusConflict, "HAS_CHILDREN", "la categoría tiene subcategorías")
	}

	if result := database.DB.Delete(&category); result.Error != nil {
		return utils.ErrorWithCode(c, fiber.StatusInternalServerError, "DELETE_ERROR", "error al eliminar categoría")
	}

	return utils.SuccessMessage(c, fiber.StatusOK, "categoría eliminada exitosamente")
}

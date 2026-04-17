package handlers

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/nexora/backend/internal/database"
	"github.com/nexora/backend/internal/models"
	"github.com/nexora/backend/internal/utils"
)

type POSSearchItem struct {
	ID             uint                     `json:"id"`
	Nombre         string                   `json:"nombre"`
	SKU            string                   `json:"sku"`
	Precio         float64                  `json:"precio"`
	Stock          int                      `json:"stock"`
	ControlaStock  bool                     `json:"controla_stock"`
	Imagen         string                   `json:"imagen"`
	TieneVariantes bool                     `json:"tiene_variantes"`
	Variantes      []models.VariantResponse `json:"variantes,omitempty"`
}

func POSSearch(c *fiber.Ctx) error {
	q := c.Query("q", "")
	includeOutOfStock := c.QueryBool("include_oos", false)

	query := database.DB.Model(&models.Product{}).Where("estado = ?", "activo")

	if q != "" {
		searchPattern := "%" + q + "%"
		query = query.Where("nombre ILIKE ? OR sku ILIKE ?", searchPattern, searchPattern)
	}

	var products []models.Product
	if err := query.Preload("Variantes.Atributos").Order("nombre ASC").Limit(50).Find(&products).Error; err != nil {
		return utils.ErrorWithCode(c, fiber.StatusInternalServerError, "DB_ERROR", "error al buscar productos")
	}

	results := make([]POSSearchItem, 0, len(products))

	for _, product := range products {
		item := POSSearchItem{
			ID:             product.ID,
			Nombre:         product.Nombre,
			SKU:            product.SKU,
			Precio:         product.Precio,
			Stock:          product.Stock,
			ControlaStock:  product.ControlaInventario,
			Imagen:         product.ImagenPrincipal,
			TieneVariantes: product.TieneVariantes,
		}

		if product.TieneVariantes {
			var variantList []models.ProductVariant
			for _, v := range product.Variantes {
				if v.Activa && (includeOutOfStock || v.Stock > 0) {
					variantList = append(variantList, v)
				}
			}

			posVariants := make([]models.VariantResponse, len(variantList))
			for i, v := range variantList {
				posVariants[i] = v.ToResponse(product.Precio)
			}
			item.Variantes = posVariants
		}

		if !product.TieneVariantes {
			if includeOutOfStock || product.Stock > 0 {
				results = append(results, item)
			}
		} else {
			results = append(results, item)
		}
	}

	return utils.SuccessData(c, fiber.StatusOK, results)
}

func POSSearchSimple(c *fiber.Ctx) error {
	q := c.Query("q", "")
	limit := c.QueryInt("limit", 20)
	categoriaID := c.QueryInt("categoria", 0)

	if limit > 50 {
		limit = 50
	}

	query := database.DB.Model(&models.Product{}).Where("estado = ?", "activo")

	if q != "" {
		searchPattern := "%" + strings.ToLower(q) + "%"
		query = query.Where("LOWER(nombre) LIKE ? OR LOWER(sku) LIKE ?", searchPattern, searchPattern)
	}

	if categoriaID > 0 {
		query = query.Where("categoria_id = ?", categoriaID)
	}

	var products []models.Product
	query.Preload("Categoria").Preload("Variantes.Atributos").Order("nombre ASC").Limit(limit).Find(&products)

	type simpleVariant struct {
		ID            uint    `json:"id"`
		Nombre        string  `json:"nombre"`
		SKU           string  `json:"sku"`
		Stock         int     `json:"stock"`
		PrecioOverride float64 `json:"precio_override,omitempty"`
	}
	type SimpleProduct struct {
		ID              uint           `json:"id"`
		Nombre          string         `json:"nombre"`
		SKU             string         `json:"sku"`
		Precio          float64        `json:"precio"`
		PrecioFinal     float64        `json:"precio_final"`
		Stock           int            `json:"stock"`
		Imagen          string         `json:"imagen"`
		TieneVariantes  bool           `json:"tiene_variantes"`
		CategoriaNombre string         `json:"categoria_nombre,omitempty"`
		Variantes       []simpleVariant `json:"variantes,omitempty"`
	}

	results := make([]SimpleProduct, len(products))
	for i, p := range products {
		sp := SimpleProduct{
			ID:              p.ID,
			Nombre:          p.Nombre,
			SKU:             p.SKU,
			Precio:          p.Precio,
			PrecioFinal:     p.Precio,
			Stock:           p.Stock,
			Imagen:          p.ImagenPrincipal,
			TieneVariantes:  p.TieneVariantes,
			CategoriaNombre: "",
		}
		if p.Categoria != nil {
			sp.CategoriaNombre = p.Categoria.Nombre
		}
		if p.TieneVariantes && len(p.Variantes) > 0 {
			for _, v := range p.Variantes {
				if v.Activa {
					sp.Variantes = append(sp.Variantes, simpleVariant{
						ID:            v.ID,
						Nombre:        v.Nombre,
						SKU:           v.SKU,
						Stock:         v.Stock,
						PrecioOverride: v.PrecioOverride,
					})
				}
			}
		}
		results[i] = sp
	}

	return utils.SuccessData(c, fiber.StatusOK, results)
}

package models

import (
	"strings"
	"time"

	"gorm.io/gorm"
)

type Product struct {
	ID               uint   `gorm:"primaryKey" json:"id"`
	SKU              string `gorm:"uniqueIndex:idx_product_sku;size:50;not null" json:"sku"`
	Nombre           string `gorm:"size:255;not null;index" json:"nombre"`
	Slug             string `gorm:"uniqueIndex:idx_product_slug;size:255" json:"slug"`
	Descripcion      string `gorm:"type:text" json:"descripcion"`
	DescripcionCorta string `gorm:"size:500" json:"descripcion_corta"`

	// Precios
	Precio              float64 `gorm:"not null;index" json:"precio"`
	PrecioAnterior      float64 `json:"precio_anterior"`
	Costo               float64 `json:"costo"`
	PorcentajeDescuento int     `json:"porcentaje_descuento"`

	// Inventario
	Stock                int  `gorm:"default:0" json:"stock"`
	StockMinimo          int  `gorm:"default:5" json:"stock_minimo"`
	PermiteStockNegativo bool `gorm:"default:false" json:"permite_stock_negativo"`
	ControlaInventario   bool `gorm:"default:true" json:"controla_inventario"`

	// Información adicional
	Peso         float64 `json:"peso"`  // en gramos
	Alto         float64 `json:"alto"`  // en cm
	Ancho        float64 `json:"ancho"` // en cm
	Largo        float64 `json:"largo"` // en cm
	UnidadMedida string  `gorm:"size:20;default:unidad" json:"unidad_medida"`

	// SEO
	MetaTitulo      string `gorm:"size:200" json:"meta_titulo"`
	MetaDescripcion string `gorm:"size:500" json:"meta_descripcion"`
	PalabrasClave   string `gorm:"size:500" json:"palabras_clave"`

	// Imágenes
	ImagenPrincipal string `gorm:"size:500" json:"imagen_principal"`

	// Categorización
	CategoriaID *uint     `json:"categoria_id"`
	Categoria   *Category `gorm:"foreignKey:CategoriaID" json:"categoria,omitempty"`
	Tags        []Tag     `gorm:"many2many:product_tags" json:"tags"`

	// Estado
	Estado    string `gorm:"size:20;default:'activo';index" json:"estado"` // activo, inactivo, agotado, descontinuado
	Destacado bool   `gorm:"default:false;index" json:"destacado"`
	Nuevo     bool   `gorm:"default:false" json:"nuevo"`

	// Variantes
	TieneVariantes bool             `gorm:"default:false" json:"tiene_variantes"`
	Variantes      []ProductVariant `gorm:"foreignKey:ProductoID" json:"variantes,omitempty"`

	// Código de barras
	Barcode string `gorm:"size:500" json:"barcode"`

	// Tiempos
	TiempoEntregaDias int            `gorm:"default:3" json:"tiempo_entrega_dias"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
	DeletedAt         gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Product) TableName() string {
	return "products"
}

// GenerateSlug genera un slug URL-friendly
func (p *Product) GenerateSlug() string {
	slug := strings.ToLower(p.Nombre)
	slug = strings.ReplaceAll(slug, " ", "-")
	slug = strings.ReplaceAll(slug, "_", "-")
	slug = strings.ReplaceAll(slug, "'", "")
	slug = strings.ReplaceAll(slug, "\"", "")

	// Remover caracteres especiales
	var result strings.Builder
	for _, r := range slug {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			result.WriteRune(r)
		}
	}

	slug = result.String()
	// Remover guiones duplicados
	for strings.Contains(slug, "--") {
		slug = strings.ReplaceAll(slug, "--", "-")
	}
	// Remover guiones al inicio y final
	slug = strings.Trim(slug, "-")

	return slug
}

// Input types
type CreateProductInput struct {
	SKU                  string  `json:"sku" validate:"required"`
	Nombre               string  `json:"nombre" validate:"required"`
	Descripcion          string  `json:"descripcion"`
	DescripcionCorta     string  `json:"descripcion_corta"`
	Precio               float64 `json:"precio" validate:"required"`
	PrecioAnterior       float64 `json:"precio_anterior"`
	Costo                float64 `json:"costo"`
	Stock                int     `json:"stock"`
	StockMinimo          int     `json:"stock_minimo"`
	PermiteStockNegativo bool    `json:"permite_stock_negativo"`
	ControlaInventario   bool    `json:"controla_inventario"`
	CategoriaID          *uint   `json:"categoria_id"`
	Tags                 []uint  `json:"tags"`
	ImagenPrincipal      string  `json:"imagen_principal"`
	Destacado            bool    `json:"destacado"`
	Nuevo                bool    `json:"nuevo"`
	TiempoEntregaDias    int     `json:"tiempo_entrega_dias"`
	MetaTitulo           string  `json:"meta_titulo"`
	MetaDescripcion      string  `json:"meta_descripcion"`
	PalabrasClave        string  `json:"palabras_clave"`
	TieneVariantes       bool    `json:"tiene_variantes"`
}

type UpdateProductInput struct {
	Nombre               string  `json:"nombre"`
	Descripcion          string  `json:"descripcion"`
	DescripcionCorta     string  `json:"descripcion_corta"`
	Precio               float64 `json:"precio"`
	PrecioAnterior       float64 `json:"precio_anterior"`
	Costo                float64 `json:"costo"`
	Stock                *int    `json:"stock"`
	StockMinimo          int     `json:"stock_minimo"`
	PermiteStockNegativo bool    `json:"permite_stock_negativo"`
	ControlaInventario   bool    `json:"controla_inventario"`
	CategoriaID          *uint   `json:"categoria_id"`
	Tags                 []uint  `json:"tags"`
	ImagenPrincipal      string  `json:"imagen_principal"`
	Estado               string  `json:"estado"`
	Destacado            *bool   `json:"destacado"`
	Nuevo                *bool   `json:"nuevo"`
	TiempoEntregaDias    int     `json:"tiempo_entrega_dias"`
	MetaTitulo           string  `json:"meta_titulo"`
	MetaDescripcion      string  `json:"meta_descripcion"`
	PalabrasClave        string  `json:"palabras_clave"`
}

// ProductResponse representa la respuesta de un producto
type ProductResponse struct {
	ID                  uint             `json:"id"`
	SKU                 string           `json:"sku"`
	Nombre              string           `json:"nombre"`
	Slug                string           `json:"slug"`
	Descripcion         string           `json:"descripcion"`
	DescripcionCorta    string           `json:"descripcion_corta"`
	Precio              float64          `json:"precio"`
	PrecioAnterior      float64          `json:"precio_anterior"`
	PorcentajeDescuento int              `json:"porcentaje_descuento"`
	Stock               int              `json:"stock"`
	StockMinimo         int              `json:"stock_minimo"`
	ImagenPrincipal     string           `json:"imagen_principal"`
	Categoria           *Category        `json:"categoria,omitempty"`
	Tags                []Tag            `json:"tags"`
	Estado              string           `json:"estado"`
	Destacado           bool             `json:"destacado"`
	Nuevo               bool             `json:"nuevo"`
	TieneVariantes      bool             `json:"tiene_variantes"`
	Variantes           []ProductVariant `json:"variantes,omitempty"`
	Barcode             string           `json:"barcode"`
	CreatedAt           time.Time        `json:"created_at"`
}

func (p *Product) ToResponse() ProductResponse {
	return ProductResponse{
		ID:                  p.ID,
		SKU:                 p.SKU,
		Nombre:              p.Nombre,
		Slug:                p.Slug,
		Descripcion:         p.Descripcion,
		DescripcionCorta:    p.DescripcionCorta,
		Precio:              p.Precio,
		PrecioAnterior:      p.PrecioAnterior,
		PorcentajeDescuento: p.PorcentajeDescuento,
		Stock:               p.Stock,
		StockMinimo:         p.StockMinimo,
		ImagenPrincipal:     p.ImagenPrincipal,
		Categoria:           p.Categoria,
		Tags:                p.Tags,
		Estado:              p.Estado,
		Destacado:           p.Destacado,
		Nuevo:               p.Nuevo,
		TieneVariantes:      p.TieneVariantes,
		Variantes:           p.Variantes,
		Barcode:             p.Barcode,
		CreatedAt:           p.CreatedAt,
	}
}

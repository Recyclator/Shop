package models

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

type ProductAttribute struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	CategoriaID *uint          `gorm:"index" json:"categoria_id"`
	Nombre      string         `gorm:"size:50;not null" json:"nombre"`
	Tipo        string         `gorm:"size:20;default:'select'" json:"tipo"`
	Valores     string         `gorm:"type:text" json:"valores"`
	Orden       int            `gorm:"default:0" json:"orden"`
	Activo      bool           `gorm:"default:true" json:"activo"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

func (ProductAttribute) TableName() string {
	return "product_attributes"
}

func (pa *ProductAttribute) GetValores() []string {
	var valores []string
	json.Unmarshal([]byte(pa.Valores), &valores)
	return valores
}

func (pa *ProductAttribute) SetValores(valores []string) {
	data, _ := json.Marshal(valores)
	pa.Valores = string(data)
}

type VariantAttributeValue struct {
	ID            uint              `gorm:"primaryKey" json:"id"`
	VariantID     uint              `gorm:"index;not null" json:"variant_id"`
	AttributeID   uint              `gorm:"index;not null" json:"attribute_id"`
	Value         string            `gorm:"size:100;not null" json:"value"`
	Attribute     *ProductAttribute `gorm:"foreignKey:AttributeID" json:"attribute,omitempty"`
	AttributeName string            `gorm:"-" json:"attribute_name,omitempty"`
}

func (VariantAttributeValue) TableName() string {
	return "variant_attribute_values"
}

type ProductVariant struct {
	ID             uint           `gorm:"primaryKey" json:"id"`
	ProductoID     uint           `gorm:"index;not null" json:"producto_id"`
	SKU            string         `gorm:"size:50;not null;uniqueIndex" json:"sku"`
	Barcode        string         `gorm:"size:255;index" json:"barcode"`
	Nombre         string         `gorm:"size:255" json:"nombre"`
	PrecioOverride float64        `json:"precio_override"`
	Stock          int            `gorm:"default:0" json:"stock"`
	StockMinimo    int            `gorm:"default:5" json:"stock_minimo"`
	Imagen         string         `gorm:"size:500" json:"imagen"`
	Activa         bool           `gorm:"default:true" json:"activa"`
	Orden          int            `gorm:"default:0" json:"orden"`
	CreatedAt      time.Time      `gorm:"index" json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`

	Producto  Product                 `gorm:"foreignKey:ProductoID" json:"-"`
	Atributos []VariantAttributeValue `gorm:"foreignKey:VariantID" json:"atributos,omitempty"`
}

func (ProductVariant) TableName() string {
	return "product_variants"
}

type CreateAttributeInput struct {
	CategoriaID     *uint    `json:"-"`
	Nombre          string   `json:"nombre" validate:"required"`
	Tipo            string   `json:"tipo"`
	ValoresOpciones []string `json:"valores_opciones"`
	Valores         []string `json:"-"`
}

type UpdateAttributeInput struct {
	Nombre  string   `json:"nombre"`
	Tipo    string   `json:"tipo"`
	Valores []string `json:"valores"`
	Activo  *bool    `json:"activo"`
}

type CreateVariantInput struct {
	SKU            string                  `json:"sku"`
	Barcode        string                  `json:"barcode"`
	Nombre         string                  `json:"nombre"`
	PrecioOverride float64                 `json:"precio_override"`
	Stock          int                     `json:"stock"`
	StockMinimo    int                     `json:"stock_minimo"`
	Imagen         string                  `json:"imagen"`
	Atributos      []VariantAttributeInput `json:"atributos"`
}

type VariantAttributeInput struct {
	AttributeID   uint   `json:"attribute_id"`
	AttributeName string `json:"attribute_name"`
	Value         string `json:"value" validate:"required"`
}

type UpdateVariantInput struct {
	SKU            string   `json:"sku"`
	Barcode        string   `json:"barcode"`
	Nombre         string   `json:"nombre"`
	PrecioOverride *float64 `json:"precio_override"`
	Stock          *int     `json:"stock"`
	StockMinimo    int      `json:"stock_minimo"`
	Imagen         *string  `json:"imagen"`
	Activa         *bool    `json:"activa"`
}

type VariantResponse struct {
	ID             uint                    `json:"id"`
	ProductoID     uint                    `json:"producto_id"`
	SKU            string                  `json:"sku"`
	Barcode        string                  `json:"barcode"`
	Nombre         string                  `json:"nombre"`
	PrecioOverride float64                 `json:"precio_override"`
	PrecioFinal    float64                 `json:"precio_final"`
	Stock          int                     `json:"stock"`
	StockMinimo    int                     `json:"stock_minimo"`
	Imagen         string                  `json:"imagen"`
	Activa         bool                    `json:"activa"`
	Atributos      []VariantAttributeValue `json:"atributos"`
}

func (pv *ProductVariant) ToResponse(basePrice float64) VariantResponse {
	precioFinal := basePrice
	if pv.PrecioOverride > 0 {
		precioFinal = pv.PrecioOverride
	}
	atributos := make([]VariantAttributeValue, len(pv.Atributos))
	for i, a := range pv.Atributos {
		attrName := a.AttributeName
		if attrName == "" && a.Attribute != nil {
			attrName = a.Attribute.Nombre
		}
		atributos[i] = VariantAttributeValue{
			ID:            a.ID,
			VariantID:     a.VariantID,
			AttributeID:   a.AttributeID,
			Value:         a.Value,
			Attribute:     a.Attribute,
			AttributeName: attrName,
		}
	}
	return VariantResponse{
		ID:             pv.ID,
		ProductoID:     pv.ProductoID,
		SKU:            pv.SKU,
		Barcode:        pv.Barcode,
		Nombre:         pv.Nombre,
		PrecioOverride: pv.PrecioOverride,
		PrecioFinal:    precioFinal,
		Stock:          pv.Stock,
		StockMinimo:    pv.StockMinimo,
		Imagen:         pv.Imagen,
		Activa:         pv.Activa,
		Atributos:      atributos,
	}
}

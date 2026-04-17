package models

import (
	"time"

	"gorm.io/gorm"
)

type ProductImage struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	ProductoID  uint           `gorm:"index;not null" json:"producto_id"`
	URL         string         `gorm:"size:500;not null" json:"url"`
	AltText     string         `gorm:"size:255" json:"alt_text"`
	Orden       int            `gorm:"default:0" json:"orden"`
	EsPrincipal bool           `gorm:"default:false" json:"es_principal"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

func (ProductImage) TableName() string {
	return "product_images"
}

type Tag struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Nombre    string         `gorm:"size:50;not null;uniqueIndex" json:"nombre"`
	Slug      string         `gorm:"size:50;not null;uniqueIndex" json:"slug"`
	Color     string         `gorm:"size:20" json:"color"` // Hex color
	Activa    bool           `gorm:"default:true" json:"activa"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	Productos []Product `gorm:"many2many:product_tags" json:"productos,omitempty"`
}

func (Tag) TableName() string {
	return "tags"
}

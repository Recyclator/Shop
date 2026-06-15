package models

import (
	"time"

	"gorm.io/gorm"
)

// ProductImage stores multiple images per product
type ProductImage struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	ProductoID  uint           `gorm:"index;not null" json:"producto_id"`
	URL         string         `gorm:"size:500;not null" json:"url"`
	Alt         string         `gorm:"size:255" json:"alt"`
	EsPrincipal bool           `gorm:"default:false" json:"es_principal"`
	Orden       int            `gorm:"default:0" json:"orden"`
	CreatedAt   time.Time      `json:"created_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

func (ProductImage) TableName() string {
	return "product_images"
}

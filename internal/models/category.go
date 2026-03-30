package models

import (
	"time"

	"gorm.io/gorm"
)

type Category struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	Nombre      string         `gorm:"size:100;not null" json:"nombre"`
	Slug        string         `gorm:"uniqueIndex:idx_category_slug;size:100" json:"slug"`
	Descripcion string         `gorm:"type:text" json:"descripcion"`
	Imagen      string         `gorm:"size:500" json:"imagen"`
	PadreID     *uint          `json:"padre_id"`
	Padre       *Category      `gorm:"foreignKey:PadreID" json:"padre,omitempty"`
	Hijos       []Category     `gorm:"foreignKey:PadreID" json:"hijos,omitempty"`
	Orden       int            `gorm:"default:0" json:"orden"`
	Activa      bool           `gorm:"default:true" json:"activa"`
	MostrarMenu bool           `gorm:"default:true" json:"mostrar_menu"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	// Relaciones
	Productos []Product `gorm:"foreignKey:CategoriaID" json:"productos,omitempty"`
}

func (Category) TableName() string {
	return "categories"
}

// Input types
type CreateCategoryInput struct {
	Nombre      string `json:"nombre" validate:"required"`
	Descripcion string `json:"descripcion"`
	Imagen      string `json:"imagen"`
	PadreID     *uint  `json:"padre_id"`
	Orden       int    `json:"orden"`
	Activa      bool   `json:"activa"`
	MostrarMenu bool   `json:"mostrar_menu"`
}

type UpdateCategoryInput struct {
	Nombre      string `json:"nombre"`
	Descripcion string `json:"descripcion"`
	Imagen      string `json:"imagen"`
	PadreID     *uint  `json:"padre_id"`
	Orden       int    `json:"orden"`
	Activa      *bool  `json:"activa"`
	MostrarMenu *bool  `json:"mostrar_menu"`
}

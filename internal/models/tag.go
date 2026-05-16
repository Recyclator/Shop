package models

import (
	"time"

	"gorm.io/gorm"
)

type Tag struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Nombre    string         `gorm:"size:50;not null;uniqueIndex" json:"nombre"`
	Slug      string         `gorm:"size:50;not null;uniqueIndex" json:"slug"`
	Color     string         `gorm:"size:20" json:"color"`
	Activa    bool           `gorm:"default:true" json:"activa"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	Productos []Product `gorm:"many2many:product_tags" json:"productos,omitempty"`
}

func (Tag) TableName() string {
	return "tags"
}

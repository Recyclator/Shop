package models

import (
	"time"

	"gorm.io/gorm"
)

type Customer struct {
	ID        uint         `gorm:"primaryKey" json:"id"`
	Cedula    string       `gorm:"uniqueIndex:idx_customer_cedula;size:20;not null" json:"cedula"`
	Nombre    string       `gorm:"size:255;not null" json:"nombre"`
	Email     string       `gorm:"size:255;not null;index" json:"email"`
	Telefono  string       `gorm:"size:20;not null" json:"telefono"`
	CreatedAt time.Time    `json:"created_at"`
	UpdatedAt time.Time    `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Relaciones
	Separados []Layaway `gorm:"foreignKey:CustomerID" json:"separados,omitempty"`
}

func (Customer) TableName() string {
	return "customers"
}

// Input types
type CreateCustomerInput struct {
	Cedula   string `json:"cedula" validate:"required"`
	Nombre   string `json:"nombre" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
	Telefono string `json:"telefono" validate:"required"`
}

type UpdateCustomerInput struct {
	Cedula   string `json:"cedula"`
	Nombre   string `json:"nombre"`
	Email    string `json:"email"`
	Telefono string `json:"telefono"`
}

// CustomerResponse representa la respuesta pública del cliente
type CustomerResponse struct {
	ID        uint      `json:"id"`
	Cedula    string    `json:"cedula"`
	Nombre    string    `json:"nombre"`
	Email     string    `json:"email"`
	Telefono  string    `json:"telefono"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ToResponse convierte el modelo a respuesta pública
func (c *Customer) ToResponse() CustomerResponse {
	return CustomerResponse{
		ID:        c.ID,
		Cedula:    c.Cedula,
		Nombre:    c.Nombre,
		Email:     c.Email,
		Telefono:  c.Telefono,
		CreatedAt: c.CreatedAt,
		UpdatedAt: c.UpdatedAt,
	}
}

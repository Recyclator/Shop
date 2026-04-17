package models

import (
	"time"

	"gorm.io/gorm"
)

type Layaway struct {
	ID               uint         `gorm:"primaryKey" json:"id"`
	CustomerID       uint         `gorm:"not null;index" json:"customer_id"`
	ProductID        uint         `gorm:"not null;index" json:"product_id"`
	Cantidad         int          `gorm:"not null;default:1" json:"cantidad"`
	PrecioUnitario   float64      `gorm:"not null" json:"precio_unitario"`
	PrecioTotal      float64      `gorm:"not null" json:"precio_total"`
	AbonoInicial     float64      `gorm:"not null" json:"abono_inicial"`
	FechaVencimiento time.Time    `gorm:"not null;index" json:"fecha_vencimiento"`
	Estado           string       `gorm:"size:20;default:'activo';index" json:"estado"` // activo, pagado, vencido, cancelado
	SaldoPendiente   float64      `gorm:"not null" json:"saldo_pendiente"`
	CreatedAt        time.Time    `json:"created_at"`
	UpdatedAt        time.Time    `json:"updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"-"`

	// Relaciones
	Customer *Customer `gorm:"foreignKey:CustomerID" json:"customer,omitempty"`
	Product  *Product  `gorm:"foreignKey:ProductID" json:"product,omitempty"`
	Abonos   []Payment `gorm:"foreignKey:LayawayID" json:"abonos,omitempty"`
}

func (Layaway) TableName() string {
	return "layaways"
}

// Input types
type CreateLayawayInput struct {
	CustomerID   uint    `json:"customer_id" validate:"required"`
	ProductID    uint    `json:"product_id" validate:"required"`
	Cantidad     int     `json:"cantidad" validate:"required,min=1"`
	AbonoInicial float64 `json:"abono_inicial" validate:"required,min=0"`
}

type UpdateLayawayInput struct {
	Estado string `json:"estado"`
}

// LayawayResponse representa la respuesta pública del separado
type LayawayResponse struct {
	ID               uint        `json:"id"`
	CustomerID       uint        `json:"customer_id"`
	ProductID        uint        `json:"product_id"`
	Cantidad         int         `json:"cantidad"`
	PrecioUnitario   float64     `json:"precio_unitario"`
	PrecioTotal      float64     `json:"precio_total"`
	AbonoInicial     float64     `json:"abono_inicial"`
	FechaVencimiento time.Time   `json:"fecha_vencimiento"`
	Estado           string      `json:"estado"`
	SaldoPendiente   float64     `json:"saldo_pendiente"`
	Customer         *Customer   `json:"customer,omitempty"`
	Product          *Product    `json:"product,omitempty"`
	CreatedAt        time.Time   `json:"created_at"`
	UpdatedAt        time.Time   `json:"updated_at"`
}

// ToResponse convierte el modelo a respuesta pública
func (l *Layaway) ToResponse() LayawayResponse {
	return LayawayResponse{
		ID:               l.ID,
		CustomerID:       l.CustomerID,
		ProductID:        l.ProductID,
		Cantidad:         l.Cantidad,
		PrecioUnitario:   l.PrecioUnitario,
		PrecioTotal:      l.PrecioTotal,
		AbonoInicial:     l.AbonoInicial,
		FechaVencimiento: l.FechaVencimiento,
		Estado:           l.Estado,
		SaldoPendiente:   l.SaldoPendiente,
		Customer:         l.Customer,
		Product:          l.Product,
		CreatedAt:        l.CreatedAt,
		UpdatedAt:        l.UpdatedAt,
	}
}

// CalculatePendingBalance calculates the pending balance after initial payment
func (l *Layaway) CalculatePendingBalance() float64 {
	return l.PrecioTotal - l.AbonoInicial
}

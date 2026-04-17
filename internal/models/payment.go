package models

import (
	"time"

	"gorm.io/gorm"
)

type Payment struct {
	ID           uint         `gorm:"primaryKey" json:"id"`
	LayawayID    uint         `gorm:"not null;index" json:"layaway_id"`
	Monto        float64      `gorm:"not null" json:"monto"`
	FechaPago    time.Time    `gorm:"not null;index" json:"fecha_pago"`
	CreatedAt    time.Time    `json:"created_at"`
	UpdatedAt    time.Time    `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`

	// Relaciones
	Layaway *Layaway `gorm:"foreignKey:LayawayID" json:"layaway,omitempty"`
}

func (Payment) TableName() string {
	return "payments"
}

// Input types
type CreatePaymentInput struct {
	Monto float64 `json:"monto" validate:"required,min=0.01"`
}

// PaymentResponse representa la respuesta pública del abono
type PaymentResponse struct {
	ID        uint      `json:"id"`
	LayawayID uint      `json:"layaway_id"`
	Monto     float64   `json:"monto"`
	FechaPago time.Time `json:"fecha_pago"`
	CreatedAt time.Time `json:"created_at"`
}

// ToResponse convierte el modelo a respuesta pública
func (p *Payment) ToResponse() PaymentResponse {
	return PaymentResponse{
		ID:        p.ID,
		LayawayID: p.LayawayID,
		Monto:     p.Monto,
		FechaPago: p.FechaPago,
		CreatedAt: p.CreatedAt,
	}
}

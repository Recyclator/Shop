package models

import (
	"time"
)

type StockMovement struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	ProductoID    uint      `gorm:"index;not null" json:"producto_id"`
	VariantID     *uint     `gorm:"index" json:"variant_id"`
	OrdenID       *uint     `gorm:"index" json:"orden_id"`
	UsuarioID     uint      `gorm:"index" json:"usuario_id"`
	Tipo          string    `gorm:"size:20;not null" json:"tipo"`
	Cantidad      int       `gorm:"not null" json:"cantidad"`
	StockAnterior int       `json:"stock_anterior"`
	StockNuevo    int       `json:"stock_nuevo"`
	Motivo        string    `gorm:"size:500" json:"motivo"`
	CreatedAt     time.Time `json:"created_at"`
}

func (StockMovement) TableName() string {
	return "stock_movements"
}

const (
	StockEntrada    = "entrada"
	StockSalida     = "salida"
	StockDevolucion = "devolucion"
	StockCorreccion = "correccion"
	StockVenta      = "venta"
)

type StockAdjustInput struct {
	Tipo      string `json:"tipo" validate:"required"`
	Cantidad  int    `json:"cantidad" validate:"required,min=1"`
	Motivo    string `json:"motivo" validate:"required"`
	VariantID *uint  `json:"variant_id"`
}

type StockAlertResponse struct {
	ID          uint   `json:"id"`
	Nombre      string `json:"nombre"`
	SKU         string `json:"sku"`
	StockActual int    `json:"stock_actual"`
	StockMinimo int    `json:"stock_minimo"`
	Deficit     int    `json:"deficit"`
	VariantID   *uint  `json:"variant_id,omitempty"`
	VariantSKU  string `json:"variant_sku,omitempty"`
}

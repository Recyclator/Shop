package models

import (
	"time"

	"gorm.io/gorm"
)

type Invoice struct {
	ID      uint   `gorm:"primaryKey" json:"id"`
	OrdenID uint   `gorm:"uniqueIndex;not null" json:"orden_id"`
	Orden   *Order `gorm:"foreignKey:OrdenID" json:"orden,omitempty"`

	// Número de factura (DIAN)
	Prefijo        string `gorm:"size:5;default:FV" json:"prefijo"`
	Numero         uint   `gorm:"not null" json:"numero"`
	NumeroCompleto string `gorm:"uniqueIndex;size:20;not null" json:"numero_completo"`

	// Fechas DIAN
	FechaEmision     time.Time  `gorm:"not null;index" json:"fecha_emision"`
	HoraEmision      string     `gorm:"size:8;not null" json:"hora_emision"`
	FechaVencimiento *time.Time `json:"fecha_vencimiento"`

	// Estado
	Estado             string     `gorm:"size:20;default:'borrador';index" json:"estado"`        // borrador, validada, enviada, aceptada, rechazada
	EstadoDIAN         string     `gorm:"size:20;default:'no_enviada';index" json:"estado_dian"` // no_enviada, envidada, aceptada, rechazada
	UUID               string     `gorm:"size:40" json:"uuid"`                             // UUID DIAN
	FechaRespuestaDIAN *time.Time `json:"fecha_respuesta_dian"`

	// Totales
	Subtotal           float64 `gorm:"not null" json:"subtotal"`
	DescuentoMonto     float64 `gorm:"default:0" json:"descuento_monto"`
	BaseImponible      float64 `gorm:"not null" json:"base_imponible"`
	ImpuestoMonto      float64 `gorm:"not null" json:"impuesto_monto"`
	ImpuestoPorcentaje float64 `gorm:"default:19" json:"impuesto_porcentaje"`
	CostoEnvio         float64 `gorm:"default:0" json:"costo_envio"`
	Total              float64 `gorm:"not null" json:"total"`

	// Información del cliente (copia para factura)
	ClienteNIT       string `gorm:"size:20" json:"cliente_nit"`
	ClienteNombre    string `gorm:"size:255;not null" json:"cliente_nombre"`
	ClienteDireccion string `gorm:"size:500" json:"cliente_direccion"`
	ClienteCiudad    string `gorm:"size:100" json:"cliente_ciudad"`
	ClienteTelefono  string `gorm:"size:20" json:"cliente_telefono"`
	ClienteEmail     string `gorm:"size:255" json:"cliente_email"`

	// Régimen fiscal
	ClienteRegimen         string `gorm:"size:20" json:"cliente_regimen"`         // comun, simplificado
	ClienteResponsabilidad string `gorm:"size:20" json:"cliente_responsabilidad"` // responsable_iva, no_responsable_iva

	// Método de pago
	MetodoPago string `gorm:"size:20;default:'contado'" json:"metodo_pago"` // contado, credito
	MedioPago  string `gorm:"size:20;default:'efectivo'" json:"medio_pago"` // efectivo, tarjeta, pse

	// Referencias
	Referencia1 string `gorm:"size:100" json:"referencia_1"`
	Referencia2 string `gorm:"size:100" json:"referencia_2"`

	// Documento relacionado (para notas crédito/débito)
	DocumentoRelacionadoID *uint `json:"documento_relacionado_id"`

	// Notas
	Notas string `gorm:"type:text" json:"notas"`

	// XML
	XMLContent    string `gorm:"type:text" json:"xml_content"`
	XMLFirmado    string `gorm:"type:text" json:"xml_firmado"`
	RespuestaDIAN string `gorm:"type:text" json:"respuesta_dian"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Invoice) TableName() string {
	return "invoices"
}

// Input types
type CreateInvoiceInput struct {
	OrdenID          uint    `json:"orden_id" validate:"required"`
	Prefijo          string  `json:"prefijo"`
	FechaVencimiento *string `json:"fecha_vencimiento"`
	Notas            string  `json:"notas"`
}

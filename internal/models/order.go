package models

import (
	"time"

	"gorm.io/gorm"
)

type Order struct {
	ID            uint   `gorm:"primaryKey" json:"id"`
	NumeroOrden   string `gorm:"uniqueIndex:idx_order_numero;size:20;not null" json:"numero_orden"`
	NumeroFactura string `gorm:"size:20" json:"numero_factura"`

	// Cliente
	ClienteID uint  `gorm:"index" json:"cliente_id"`
	Cliente   *User `gorm:"foreignKey:ClienteID" json:"cliente,omitempty"`

	// Información del cliente (copia para historial)
	ClienteNombre    string `gorm:"size:200" json:"cliente_nombre"`
	ClienteEmail     string `gorm:"size:255" json:"cliente_email"`
	ClienteTelefono  string `gorm:"size:20" json:"cliente_telefono"`
	ClienteDocumento string `gorm:"size:20" json:"cliente_documento"`

	// Dirección de entrega
	DireccionEnvio    string `gorm:"size:500" json:"direccion_envio"`
	CiudadEnvio       string `gorm:"size:100" json:"ciudad_envio"`
	DepartamentoEnvio string `gorm:"size:100" json:"departamento_envio"`
	CodigoPostalEnvio string `gorm:"size:10" json:"codigo_postal_envio"`

	// Vendedor
	VendedorID *uint `gorm:"index" json:"vendedor_id"`
	Vendedor   *User `gorm:"foreignKey:VendedorID" json:"vendedor,omitempty"`

	// Totales
	Subtotal           float64 `gorm:"not null" json:"subtotal"`
	Descuento          float64 `gorm:"default:0" json:"descuento"`
	Impuesto           float64 `gorm:"default:0" json:"impuesto"`
	ImpuestoPorcentaje float64 `gorm:"default:19" json:"impuesto_porcentaje"`
	CostoEnvio         float64 `gorm:"default:0" json:"costo_envio"`
	Total              float64 `gorm:"not null" json:"total"`

	// Descuentos aplicados
	CodigoDescuento     string  `gorm:"size:50" json:"codigo_descuento"`
	PorcentajeDescuento float64 `gorm:"default:0" json:"porcentaje_descuento"`

	// Estado y método de pago
	Estado         string     `gorm:"size:20;default:'pendiente';index" json:"estado"` // pendiente, confirmado, procesamiento, enviado, entregado, cancelado, reembolso
	MetodoPago     string     `gorm:"size:50" json:"metodo_pago"`                      // efectivo, tarjeta, pse, nequi, transferencia
	ReferenciaPago string     `gorm:"size:100" json:"referencia_pago"`
	FechaPago      *time.Time `json:"fecha_pago"`
	EstadoPago     string     `gorm:"size:20;default:'pendiente'" json:"estado_pago"` // pendiente, pagado, fallido, reembolso

	// Notas
	Notas         string `gorm:"type:text" json:"notas"`
	NotasInternas string `gorm:"type:text" json:"notas_internas"`

	// Fechas
	FechaEntregaEsperada *time.Time `json:"fecha_entrega_esperada"`
	FechaEntregaReal     *time.Time `json:"fecha_entrega_real"`

	// Tracking
	NumeroTracking string `gorm:"size:100" json:"numero_tracking"`
	EmpresaEnvio   string `gorm:"size:100" json:"empresa_envio"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Relaciones
	Items   []OrderItem `gorm:"foreignKey:OrdenID" json:"items"`
	Factura *Invoice    `gorm:"foreignKey:OrdenID" json:"factura,omitempty"`
}

func (Order) TableName() string {
	return "orders"
}

type OrderItem struct {
	ID      uint `gorm:"primaryKey" json:"id"`
	OrdenID uint `gorm:"index;not null" json:"orden_id"`

	ProductoID uint     `gorm:"index" json:"producto_id"`
	Producto   *Product `gorm:"foreignKey:ProductoID" json:"producto,omitempty"`

	NombreProducto string `gorm:"size:255;not null" json:"nombre_producto"`
	SKU            string `gorm:"size:50" json:"sku"`

	VarianteID   *uint  `json:"variante_id"`
	VarianteInfo string `gorm:"size:255" json:"variante_info"` // "Color: Rojo, Talla: XL"

	Cantidad       int     `gorm:"not null" json:"cantidad"`
	PrecioUnitario float64 `gorm:"not null" json:"precio_unitario"`
	Descuento      float64 `gorm:"default:0" json:"descuento"`
	Impuesto       float64 `gorm:"default:0" json:"impuesto"`
	Total          float64 `gorm:"not null" json:"total"`

	// Estado
	Estado string `gorm:"size:20;default:'pendiente'" json:"estado"` // pendiente, confirmado, enviado, entregado, cancelado, reembolso
	Notas  string `gorm:"size:500" json:"notas"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Orden Order `gorm:"foreignKey:OrdenID" json:"-"`
}

func (OrderItem) TableName() string {
	return "order_items"
}

// Input types
type CreateOrderInput struct {
	ClienteID         *uint                  `json:"cliente_id"`
	DireccionEnvio    string                 `json:"direccion_envio"`
	CiudadEnvio       string                 `json:"ciudad_envio"`
	DepartamentoEnvio string                 `json:"departamento_envio"`
	CodigoPostalEnvio string                 `json:"codigo_postal_envio"`
	MetodoPago        string                 `json:"metodo_pago"`
	ReferenciaPago    string                 `json:"referencia_pago"`
	Notas             string                 `json:"notas"`
	NotasInternas     string                 `json:"notas_internas"`
	CodigoDescuento   string                 `json:"codigo_descuento"`
	Items             []CreateOrderItemInput `json:"items" validate:"required,min=1"`
}

type CreateOrderItemInput struct {
	ProductoID     uint    `json:"producto_id" validate:"required"`
	VarianteID     *uint   `json:"variante_id"`
	Cantidad       int     `json:"cantidad" validate:"required,min=1"`
	PrecioUnitario float64 `json:"precio_unitario"`
	Notas          string  `json:"notas"`
}

type OrderResponse struct {
	ID            uint        `json:"id"`
	NumeroOrden   string      `json:"numero_orden"`
	ClienteID     uint        `json:"cliente_id"`
	Cliente       *User       `json:"cliente,omitempty"`
	ClienteNombre string      `json:"cliente_nombre"`
	Subtotal      float64     `json:"subtotal"`
	Descuento     float64     `json:"descuento"`
	Impuesto      float64     `json:"impuesto"`
	CostoEnvio    float64     `json:"costo_envio"`
	Total         float64     `json:"total"`
	Estado        string      `json:"estado"`
	MetodoPago    string      `json:"metodo_pago"`
	FechaPago     *time.Time  `json:"fecha_pago"`
	EstadoPago    string      `json:"estado_pago"`
	VendedorID    *uint       `json:"vendedor_id"`
	Vendedor      *User       `json:"vendedor,omitempty"`
	Items         []OrderItem `json:"items"`
	CreatedAt     time.Time   `json:"created_at"`
}

func (o Order) ToResponse() OrderResponse {
	var clienteID uint
	if o.ClienteID > 0 {
		clienteID = o.ClienteID
	}
	return OrderResponse{
		ID:            o.ID,
		NumeroOrden:   o.NumeroOrden,
		ClienteID:     clienteID,
		Cliente:       o.Cliente,
		ClienteNombre: o.ClienteNombre,
		Subtotal:      o.Subtotal,
		Descuento:     o.Descuento,
		Impuesto:      o.Impuesto,
		CostoEnvio:    o.CostoEnvio,
		Total:         o.Total,
		Estado:        o.Estado,
		MetodoPago:    o.MetodoPago,
		FechaPago:     o.FechaPago,
		EstadoPago:    o.EstadoPago,
		VendedorID:    o.VendedorID,
		Vendedor:      o.Vendedor,
		Items:         o.Items,
		CreatedAt:     o.CreatedAt,
	}
}

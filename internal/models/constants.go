package models

const (
	OrderPendiente    = "pendiente"
	OrderCompletado   = "completado"
	OrderCancelado    = "cancelado"
	OrderEnTransito   = "en_transito"
	OrderEnPreparacion = "en_preparacion"
)

const (
	PaymentPagado     = "pagado"
	PaymentPendiente  = "pendiente"
	PaymentReembolsado = "reembolsado"
	PaymentParcial    = "parcial"
)

const (
	ProductActivo   = "activo"
	ProductInactivo = "inactivo"
	ProductAgotado  = "agotado"
)

const (
	RoleSuperadmin = "superadmin"
	RoleAdmin      = "admin"
	RoleVendedor   = "vendedor"
	RoleContador   = "contador"
	RoleCliente    = "cliente"
)

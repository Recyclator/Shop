package models

import (
	"time"

	"gorm.io/gorm"
)

type Permission struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	Codigo      string         `gorm:"uniqueIndex:idx_permission_codigo;size:50;not null" json:"codigo"`
	Nombre      string         `gorm:"size:100;not null" json:"nombre"`
	Descripcion string         `gorm:"size:255" json:"descripcion"`
	Modulo      string         `gorm:"size:50;not null;index" json:"modulo"`
	Accion      string         `gorm:"size:50;not null" json:"accion"`
	Activo      bool           `gorm:"default:true" json:"activo"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	// Relaciones
	Roles []Role `gorm:"many2many:role_permissions" json:"roles,omitempty"`
}

func (Permission) TableName() string {
	return "permissions"
}

func GetDefaultPermissions() []Permission {
	return []Permission{
		// Usuarios
		{Codigo: "users.create", Nombre: "Crear usuarios", Modulo: "users", Accion: "create", Descripcion: "Permite crear nuevos usuarios"},
		{Codigo: "users.read", Nombre: "Ver usuarios", Modulo: "users", Accion: "read", Descripcion: "Permite ver listados y detalles de usuarios"},
		{Codigo: "users.update", Nombre: "Actualizar usuarios", Modulo: "users", Accion: "update", Descripcion: "Permite modificar datos de usuarios"},
		{Codigo: "users.delete", Nombre: "Eliminar usuarios", Modulo: "users", Accion: "delete", Descripcion: "Permite eliminar usuarios"},

		// Roles
		{Codigo: "roles.create", Nombre: "Crear roles", Modulo: "roles", Accion: "create", Descripcion: "Permite crear nuevos roles"},
		{Codigo: "roles.read", Nombre: "Ver roles", Modulo: "roles", Accion: "read", Descripcion: "Permite ver listados y detalles de roles"},
		{Codigo: "roles.update", Nombre: "Actualizar roles", Modulo: "roles", Accion: "update", Descripcion: "Permite modificar roles"},
		{Codigo: "roles.delete", Nombre: "Eliminar roles", Modulo: "roles", Accion: "delete", Descripcion: "Permite eliminar roles"},
		{Codigo: "roles.assign", Nombre: "Asignar roles", Modulo: "roles", Accion: "assign", Descripcion: "Permite asignar roles a usuarios"},

		// Productos
		{Codigo: "products.create", Nombre: "Crear productos", Modulo: "products", Accion: "create", Descripcion: "Permite crear nuevos productos"},
		{Codigo: "products.read", Nombre: "Ver productos", Modulo: "products", Accion: "read", Descripcion: "Permite ver listados y detalles de productos"},
		{Codigo: "products.update", Nombre: "Actualizar productos", Modulo: "products", Accion: "update", Descripcion: "Permite modificar productos"},
		{Codigo: "products.delete", Nombre: "Eliminar productos", Modulo: "products", Accion: "delete", Descripcion: "Permite eliminar productos"},

		// Categorías
		{Codigo: "categories.create", Nombre: "Crear categorías", Modulo: "categories", Accion: "create", Descripcion: "Permite crear categorías"},
		{Codigo: "categories.read", Nombre: "Ver categorías", Modulo: "categories", Accion: "read", Descripcion: "Permite ver categorías"},
		{Codigo: "categories.update", Nombre: "Actualizar categorías", Modulo: "categories", Accion: "update", Descripcion: "Permite modificar categorías"},
		{Codigo: "categories.delete", Nombre: "Eliminar categorías", Modulo: "categories", Accion: "delete", Descripcion: "Permite eliminar categorías"},

		// Pedidos
		{Codigo: "orders.create", Nombre: "Crear pedidos", Modulo: "orders", Accion: "create", Descripcion: "Permite crear pedidos"},
		{Codigo: "orders.read", Nombre: "Ver pedidos", Modulo: "orders", Accion: "read", Descripcion: "Permite ver pedidos"},
		{Codigo: "orders.update", Nombre: "Actualizar pedidos", Modulo: "orders", Accion: "update", Descripcion: "Permite modificar pedidos"},
		{Codigo: "orders.delete", Nombre: "Eliminar pedidos", Modulo: "orders", Accion: "delete", Descripcion: "Permite eliminar pedidos"},
		{Codigo: "orders.cancel", Nombre: "Cancelar pedidos", Modulo: "orders", Accion: "cancel", Descripcion: "Permite cancelar pedidos"},

		// Facturas
		{Codigo: "invoices.create", Nombre: "Crear facturas", Modulo: "invoices", Accion: "create", Descripcion: "Permite crear facturas electrónicas"},
		{Codigo: "invoices.read", Nombre: "Ver facturas", Modulo: "invoices", Accion: "read", Descripcion: "Permite ver facturas"},
		{Codigo: "invoices.update", Nombre: "Actualizar facturas", Modulo: "invoices", Accion: "update", Descripcion: "Permite modificar facturas"},
		{Codigo: "invoices.send", Nombre: "Enviar facturas", Modulo: "invoices", Accion: "send", Descripcion: "Permite enviar facturas por email"},

		// Reportes
		{Codigo: "reports.read", Nombre: "Ver reportes", Modulo: "reports", Accion: "read", Descripcion: "Permite ver reportes"},
		{Codigo: "reports.export", Nombre: "Exportar reportes", Modulo: "reports", Accion: "export", Descripcion: "Permite exportar reportes"},

		// Configuración
		{Codigo: "settings.read", Nombre: "Ver configuración", Modulo: "settings", Accion: "read", Descripcion: "Permite ver configuración"},
		{Codigo: "settings.update", Nombre: "Actualizar configuración", Modulo: "settings", Accion: "update", Descripcion: "Permite modificar configuración"},

		// Pagos
		{Codigo: "payments.read", Nombre: "Ver pagos", Modulo: "payments", Accion: "read", Descripcion: "Permite ver pagos"},
		{Codigo: "payments.process", Nombre: "Procesar pagos", Modulo: "payments", Accion: "process", Descripcion: "Permite procesar pagos"},

		// Clientes
		{Codigo: "customers.read", Nombre: "Ver clientes", Modulo: "customers", Accion: "read", Descripcion: "Permite ver clientes"},
		{Codigo: "customers.update", Nombre: "Actualizar clientes", Modulo: "customers", Accion: "update", Descripcion: "Permite modificar clientes"},

		// Inventario
		{Codigo: "inventory.read", Nombre: "Ver inventario", Modulo: "inventory", Accion: "read", Descripcion: "Permite ver inventario"},
		{Codigo: "inventory.update", Nombre: "Actualizar inventario", Modulo: "inventory", Accion: "update", Descripcion: "Permite modificar inventario"},
	}
}

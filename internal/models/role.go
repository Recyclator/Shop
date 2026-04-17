package models

import (
	"time"

	"gorm.io/gorm"
)

type Role struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	Nombre      string         `gorm:"uniqueIndex:idx_role_nombre;size:50;not null" json:"nombre"`
	DisplayName string         `gorm:"size:100;not null" json:"display_name"`
	Descripcion string         `gorm:"size:255" json:"descripcion"`
	Nivel       int            `gorm:"default:100;index" json:"nivel"` // 1=superadmin, 10=admin, 20=vendedor, 100=cliente
	Activo      bool           `gorm:"default:true" json:"activo"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	// Relaciones
	Permisos []Permission `gorm:"many2many:role_permissions" json:"permisos,omitempty"`
	Usuarios []User       `gorm:"many2many:user_roles" json:"usuarios,omitempty"`
}

func (Role) TableName() string {
	return "roles"
}

func GetDefaultRoles() []Role {
	return []Role{
		{
			Nombre:      "superadmin",
			DisplayName: "Super Administrador",
			Descripcion: "Acceso total al sistema, gestión de empresas y configuraciones globales",
			Nivel:       1,
			Activo:      true,
		},
		{
			Nombre:      "admin",
			DisplayName: "Administrador",
			Descripcion: "Gestión completa de la tienda, productos, pedidos y clientes",
			Nivel:       10,
			Activo:      true,
		},
		{
			Nombre:      "vendedor",
			DisplayName: "Vendedor",
			Descripcion: "Punto de venta, creación de pedidos y atención al cliente",
			Nivel:       20,
			Activo:      true,
		},
		{
			Nombre:      "contador",
			DisplayName: "Contador/Finanzas",
			Descripcion: "Reportes, finanzas y gestión de facturas",
			Nivel:       15,
			Activo:      true,
		},
		{
			Nombre:      "cliente",
			DisplayName: "Cliente",
			Descripcion: "Cliente final del e-commerce",
			Nivel:       100,
			Activo:      true,
		},
	}
}

func AssignDefaultPermissions(db *gorm.DB) error {
	// Obtener permisos por nombre de acción
	var permissions []Permission
	if err := db.Find(&permissions).Error; err != nil {
		return err
	}

	permisosMap := make(map[string]Permission)
	for _, p := range permissions {
		permisosMap[p.Codigo] = p
	}

	// Superadmin: todos los permisos
	var superadmin Role
	if err := db.Where("nombre = ?", "superadmin").First(&superadmin).Error; err != nil {
		return err
	}
	var allPerms []Permission
	for _, p := range permissions {
		allPerms = append(allPerms, p)
	}
	superadmin.Permisos = allPerms
	db.Save(&superadmin)

	// Admin: todos excepto eliminar usuarios y roles
	adminPerms := []string{
		"users.create", "users.read", "users.update",
		"roles.read",
		"products.create", "products.read", "products.update", "products.delete",
		"categories.create", "categories.read", "categories.update", "categories.delete",
		"orders.create", "orders.read", "orders.update", "orders.cancel",
		"invoices.create", "invoices.read", "invoices.update", "invoices.send",
		"reports.read", "reports.export",
		"settings.read", "settings.update",
		"payments.read", "payments.process",
		"customers.read", "customers.update",
		"inventory.read", "inventory.update",
	}
	var admin Role
	if err := db.Where("nombre = ?", "admin").First(&admin).Error; err != nil {
		return err
	}
	var adminPermisos []Permission
	for _, cod := range adminPerms {
		if p, ok := permisosMap[cod]; ok {
			adminPermisos = append(adminPermisos, p)
		}
	}
	admin.Permisos = adminPermisos
	db.Save(&admin)

	// Vendedor
	vendedorPerms := []string{
		"products.read",
		"categories.read", "categories.create", "categories.update", "categories.delete",
		"orders.create", "orders.read", "orders.update",
		"customers.read", "customers.create", "customers.update",
		"layaways.read", "layaways.create", "layaways.update",
		"inventory.read", "inventory.update",
	}
	var vendedor Role
	if err := db.Where("nombre = ?", "vendedor").First(&vendedor).Error; err != nil {
		return err
	}
	var vendedorPermisos []Permission
	for _, cod := range vendedorPerms {
		if p, ok := permisosMap[cod]; ok {
			vendedorPermisos = append(vendedorPermisos, p)
		}
	}
	vendedor.Permisos = vendedorPermisos
	db.Save(&vendedor)

	// Contador
	contadorPerms := []string{
		"orders.read",
		"invoices.create", "invoices.read", "invoices.update", "invoices.send",
		"reports.read", "reports.export",
		"payments.read",
		"customers.read",
	}
	var contador Role
	if err := db.Where("nombre = ?", "contador").First(&contador).Error; err != nil {
		return err
	}
	var contadorPermisos []Permission
	for _, cod := range contadorPerms {
		if p, ok := permisosMap[cod]; ok {
			contadorPermisos = append(contadorPermisos, p)
		}
	}
	contador.Permisos = contadorPermisos
	db.Save(&contador)

	// Cliente
	clientePerms := []string{
		"products.read",
		"orders.create", "orders.read",
	}
	var cliente Role
	if err := db.Where("nombre = ?", "cliente").First(&cliente).Error; err != nil {
		return err
	}
	var clientePermisos []Permission
	for _, cod := range clientePerms {
		if p, ok := permisosMap[cod]; ok {
			clientePermisos = append(clientePermisos, p)
		}
	}
	cliente.Permisos = clientePermisos
	db.Save(&cliente)

	return nil
}

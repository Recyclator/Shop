package models

import (
	"time"
)

// RolePermission representa la relación entre roles y permisos
type RolePermission struct {
	RoleID       uint      `gorm:"primaryKey" json:"role_id"`
	PermissionID uint      `gorm:"primaryKey" json:"permission_id"`
	GrantedAt    time.Time `gorm:"autoCreateTime" json:"granted_at"`
	GrantedBy    uint      `json:"granted_by"`

	Role       Role       `gorm:"foreignKey:RoleID" json:"-"`
	Permission Permission `gorm:"foreignKey:PermissionID" json:"-"`
}

func (RolePermission) TableName() string {
	return "role_permissions"
}

// Default role-permission join table name used by GORM many2many
const RolePermissionJoinTable = "role_permissions"

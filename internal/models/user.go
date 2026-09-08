package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID                  uint           `gorm:"primaryKey" json:"id"`
	Email               string         `gorm:"uniqueIndex:idx_user_email;size:255;not null" json:"email"`
	Password            string         `gorm:"not null" json:"-"`
	Nombre              string         `gorm:"size:100;not null" json:"nombre"`
	Apellido            string         `gorm:"size:100" json:"apellido"`
	Telefono            string         `gorm:"size:20" json:"telefono"`
	TipoDocumento       string         `gorm:"size:2" json:"tipo_documento"` // CC, CE, NIT, TI, RC, PA
	NumeroDocumento     string         `gorm:"size:20;index" json:"numero_documento"`
	Direccion           string         `gorm:"size:500" json:"direccion"`
	Ciudad              string         `gorm:"size:100" json:"ciudad"`
	Departamento        string         `gorm:"size:100" json:"departamento"`
	Pais                string         `gorm:"size:2;default:CO" json:"pais"`
	CodigoPostal        string         `gorm:"size:10" json:"codigo_postal"`
	Foto                string         `gorm:"size:500" json:"foto"`
	Activo              bool           `gorm:"default:true;index" json:"activo"`
	Theme               string         `gorm:"size:10;default:dark" json:"theme"` // "light" or "dark"
	EmailVerificado     bool           `gorm:"default:false" json:"email_verificado"`
	UltimoLogin         *time.Time     `json:"ultimo_login"`
	PasswordChangedAt   *time.Time     `json:"password_changed_at"`
	FailedLoginAttempts int            `gorm:"default:0" json:"failed_login_attempts"`
	LockedUntil         *time.Time     `json:"locked_until"`
	TwoFactorEnabled    bool           `gorm:"default:false" json:"two_factor_enabled"`
	TwoFactorSecret     string         `gorm:"size:255" json:"-"`
	CreatedAt           time.Time      `gorm:"index" json:"created_at"`
	UpdatedAt           time.Time      `json:"updated_at"`
	DeletedAt           gorm.DeletedAt `gorm:"index" json:"-"`

	// Relaciones
	Roles []Role `gorm:"many2many:user_roles" json:"roles,omitempty"`
}

// TableName returns the table name
func (User) TableName() string {
	return "users"
}

// Input types para auth
type RegisterInput struct {
	Email           string `json:"email" validate:"required,email"`
	Password        string `json:"password" validate:"required,min=12"`
	ConfirmPassword string `json:"confirm_password" validate:"required"`
	Nombre          string `json:"nombre" validate:"required"`
	Apellido        string `json:"apellido"`
	Telefono        string `json:"telefono"`
	TipoDocumento   string `json:"tipo_documento"`
	NumeroDocumento string `json:"numero_documento"`
}

type LoginInput struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type RefreshTokenInput struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type ChangePasswordInput struct {
	CurrentPassword string `json:"current_password" validate:"required"`
	NewPassword     string `json:"new_password" validate:"required,min=12"`
	ConfirmPassword string `json:"confirm_password" validate:"required"`
}

type UpdateUserInput struct {
	Nombre          string `json:"nombre"`
	Apellido        string `json:"apellido"`
	Telefono        string `json:"telefono"`
	TipoDocumento   string `json:"tipo_documento"`
	NumeroDocumento string `json:"numero_documento"`
	Direccion       string `json:"direccion"`
	Ciudad          string `json:"ciudad"`
	Departamento    string `json:"departamento"`
	Pais            string `json:"pais"`
	CodigoPostal    string `json:"codigo_postal"`
	Foto            string `json:"foto"`
	Activo          *bool  `json:"activo"`
}

type UpdateProfileInput struct {
	Nombre          string `json:"nombre"`
	Apellido        string `json:"apellido"`
	Telefono        string `json:"telefono"`
	TipoDocumento   string `json:"tipo_documento"`
	NumeroDocumento string `json:"numero_documento"`
	Direccion       string `json:"direccion"`
	Ciudad          string `json:"ciudad"`
	Departamento    string `json:"departamento"`
	Pais            string `json:"pais"`
	CodigoPostal    string `json:"codigo_postal"`
	Foto            string `json:"foto"`
}

// UserResponse define la respuesta pública del usuario
type UserResponse struct {
	ID              uint      `json:"id"`
	Email           string    `json:"email"`
	Nombre          string    `json:"nombre"`
	Apellido        string    `json:"apellido"`
	Telefono        string    `json:"telefono"`
	TipoDocumento   string    `json:"tipo_documento"`
	NumeroDocumento string    `json:"numero_documento"`
	Activo          bool      `json:"activo"`
	Theme           string    `json:"theme"`
	Roles           []Role    `json:"roles"`
	CreatedAt       time.Time `json:"created_at"`
}

// ToResponse convierte el modelo a respuesta pública
func (u *User) ToResponse() UserResponse {
	return UserResponse{
		ID:              u.ID,
		Email:           u.Email,
		Nombre:          u.Nombre,
		Apellido:        u.Apellido,
		Telefono:        u.Telefono,
		TipoDocumento:   u.TipoDocumento,
		NumeroDocumento: u.NumeroDocumento,
		Activo:          u.Activo,
		Theme:           u.Theme,
		Roles:           u.Roles,
		CreatedAt:       u.CreatedAt,
	}
}

// GetUserRoleNames retorna los nombres de los roles del usuario
func (u *User) GetUserRoleNames() []string {
	var roles []string
	for _, role := range u.Roles {
		roles = append(roles, role.Nombre)
	}
	return roles
}

// HasRole verifica si el usuario tiene un rol específico
func (u *User) HasRole(roleName string) bool {
	for _, role := range u.Roles {
		if role.Nombre == roleName {
			return true
		}
	}
	return false
}

// HasPermission verifica si el usuario tiene un permiso específico
func (u *User) HasPermission(permissionCod string) bool {
	for _, role := range u.Roles {
		for _, perm := range role.Permisos {
			if perm.Codigo == permissionCod {
				return true
			}
		}
	}
	return false
}

// GetHighestRoleLevel retorna el nivel más alto (menor número) de los roles del usuario
func (u *User) GetHighestRoleLevel() int {
	level := 999
	for _, role := range u.Roles {
		if role.Nivel < level {
			level = role.Nivel
		}
	}
	return level
}

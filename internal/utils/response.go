package utils

import (
	"github.com/gofiber/fiber/v2"
)

type Response struct {
	Success bool         `json:"success"`
	Message string       `json:"message,omitempty"`
	Data    interface{}  `json:"data,omitempty"`
	Error   *ErrorDetail `json:"error,omitempty"`
}

type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

type PaginatedResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
	Meta    *Meta       `json:"meta"`
}

type Meta struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

// Respuestas estándar
func Success(c *fiber.Ctx, status int, message string, data interface{}) error {
	return c.Status(status).JSON(Response{
		Success: true,
		Message: message,
		Data:    data,
	})
}

func SuccessMessage(c *fiber.Ctx, status int, message string) error {
	return c.Status(status).JSON(Response{
		Success: true,
		Message: message,
	})
}

func SuccessData(c *fiber.Ctx, status int, data interface{}) error {
	return c.Status(status).JSON(Response{
		Success: true,
		Data:    data,
	})
}

func Error(c *fiber.Ctx, status int, message string) error {
	return c.Status(status).JSON(Response{
		Success: false,
		Error: &ErrorDetail{
			Message: message,
		},
	})
}

func ErrorWithCode(c *fiber.Ctx, status int, code, message string) error {
	return c.Status(status).JSON(Response{
		Success: false,
		Error: &ErrorDetail{
			Code:    code,
			Message: message,
		},
	})
}

func ErrorWithDetails(c *fiber.Ctx, status int, code, message, details string) error {
	return c.Status(status).JSON(Response{
		Success: false,
		Error: &ErrorDetail{
			Code:    code,
			Message: message,
			Details: details,
		},
	})
}

func Paginated(c *fiber.Ctx, status int, data interface{}, page, limit int, total int64) error {
	totalPages := int(total) / limit
	if int(total)%limit > 0 {
		totalPages++
	}

	return c.Status(status).JSON(PaginatedResponse{
		Success: true,
		Data:    data,
		Meta: &Meta{
			Page:       page,
			Limit:      limit,
			Total:      total,
			TotalPages: totalPages,
		},
	})
}

// Errores comunes
var (
	ErrUnauthorized    = fiber.NewError(fiber.StatusUnauthorized, "no autorizado")
	ErrForbidden       = fiber.NewError(fiber.StatusForbidden, "acceso prohibido")
	ErrNotFound        = fiber.NewError(fiber.StatusNotFound, "no encontrado")
	ErrBadRequest      = fiber.NewError(fiber.StatusBadRequest, "solicitud inválida")
	ErrConflict        = fiber.NewError(fiber.StatusConflict, "conflicto de datos")
	ErrInternalServer  = fiber.NewError(fiber.StatusInternalServerError, "error interno del servidor")
	ErrTooManyRequests = fiber.NewError(fiber.StatusTooManyRequests, "demasiadas solicitudes")
)

// Errores de autenticación
var (
	ErrInvalidCredentials = fiber.NewError(fiber.StatusUnauthorized, "credenciales inválidas")
	ErrTokenExpired       = fiber.NewError(fiber.StatusUnauthorized, "token expirado")
	ErrTokenInvalid       = fiber.NewError(fiber.StatusUnauthorized, "token inválido")
	ErrUserInactive       = fiber.NewError(fiber.StatusForbidden, "usuario inactivo")
	ErrUserLocked         = fiber.NewError(fiber.StatusForbidden, "usuario bloqueado")
)

// Errores de validación
var (
	ErrInvalidEmail     = fiber.NewError(fiber.StatusBadRequest, "email inválido")
	ErrInvalidPassword  = fiber.NewError(fiber.StatusBadRequest, "contraseña inválida")
	ErrPasswordMismatch = fiber.NewError(fiber.StatusBadRequest, "las contraseñas no coinciden")
	ErrEmailExists      = fiber.NewError(fiber.StatusConflict, "el email ya está registrado")
	ErrSKUExists        = fiber.NewError(fiber.StatusConflict, "el SKU ya existe")
	ErrResourceNotFound = fiber.NewError(fiber.StatusNotFound, "recurso no encontrado")
)

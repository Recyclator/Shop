package utils

import (
	"log"
	"time"
)

// AuditEvent representa un evento de auditoría
type AuditEvent struct {
	Timestamp   time.Time `json:"timestamp"`
	Action      string    `json:"action"`       // login, logout, create, update, delete, password_change, etc.
	UserID      uint      `json:"user_id"`      // ID del usuario que realizó la acción
	UserEmail   string    `json:"user_email"`   // Email del usuario
	Resource    string    `json:"resource"`     // Tipo de recurso: user, product, order, etc.
	ResourceID  uint      `json:"resource_id"`  // ID del recurso afectado
	IP          string    `json:"ip"`           // IP del cliente
	UserAgent   string    `json:"user_agent"`   // User-Agent del cliente
	Success     bool      `json:"success"`      // Si la acción fue exitosa
	Description string    `json:"description"`  // Descripción de la acción
}

// AuditLogger maneja el logging de eventos de seguridad
type AuditLogger struct {
	enabled bool
}

var auditLogger *AuditLogger

func init() {
	auditLogger = &AuditLogger{enabled: true}
}

// Log registra un evento de auditoría en el log del sistema
func (a *AuditLogger) Log(event AuditEvent) {
	if !a.enabled {
		return
	}

	// Asegurar que el timestamp esté establecido
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	// Log estructurado en formato simple para debugging
	log.Printf("[AUDIT] [%s] Action=%s UserID=%d User=%s Resource=%s:%d IP=%s Success=%v Desc=%s",
		event.Timestamp.Format(time.RFC3339),
		event.Action,
		event.UserID,
		event.UserEmail,
		event.Resource,
		event.ResourceID,
		event.IP,
		event.Success,
		event.Description,
	)
}

// LogSimple crea un log de auditoría simple con los parámetros básicos
func LogSimple(userID uint, userEmail, action, resource string, resourceID uint, success bool, description string) {
	auditLogger.Log(AuditEvent{
		Timestamp:   time.Now(),
		Action:      action,
		UserID:      userID,
		UserEmail:   userEmail,
		Resource:    resource,
		ResourceID:  resourceID,
		Success:     success,
		Description: description,
	})
}

// LogWithContext crea un log de auditoría con contexto de request (IP y User-Agent)
func LogWithContext(action, resource string, resourceID uint, success bool, description string, ip, userAgent string) {
	auditLogger.Log(AuditEvent{
		Timestamp:   time.Now(),
		Action:      action,
		Resource:    resource,
		ResourceID:  resourceID,
		IP:          ip,
		UserAgent:   userAgent,
		Success:     success,
		Description: description,
	})
}

// Disable desactiva el logging de auditoría (útil para tests)
func Disable() {
	auditLogger.enabled = false
}

// Enable activa el logging de auditoría
func Enable() {
	auditLogger.enabled = true
}

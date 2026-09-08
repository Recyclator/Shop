package handlers

import (
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/nexora/backend/internal/config"
	"github.com/nexora/backend/internal/database"
	"github.com/nexora/backend/internal/middleware"
	"github.com/nexora/backend/internal/models"
	"github.com/nexora/backend/internal/utils"
)

// Register maneja el registro de nuevos usuarios
// POST /api/auth/register
func Register(c *fiber.Ctx) error {
	var input models.RegisterInput

	if err := c.BodyParser(&input); err != nil {
		return utils.ErrorWithCode(c, fiber.StatusBadRequest, "INVALID_BODY", "formato de datos inválido")
	}

	// Validar estructura con go-playground validator
	if validationErrors := utils.ValidateStruct(input); validationErrors != nil {
		return utils.ErrorWithCode(c, fiber.StatusBadRequest, "VALIDATION_ERROR", utils.FirstValidationError(validationErrors))
	}

	// Validar formato de email
	if !isValidEmail(input.Email) {
		return utils.ErrorWithCode(c, fiber.StatusBadRequest, "INVALID_EMAIL", "formato de email inválido")
	}

	// Validar contraseña
	if err := utils.ValidatePassword(input.Password); err != nil {
		return utils.ErrorWithCode(c, fiber.StatusBadRequest, "WEAK_PASSWORD", err.Error())
	}

	// Verificar que las contraseñas coincidan
	if input.Password != input.ConfirmPassword {
		return utils.ErrorWithCode(c, fiber.StatusBadRequest, "PASSWORD_MISMATCH", "las contraseñas no coinciden")
	}

	// Verificar si el email ya existe
	var existingUser models.User
	if result := database.DB.Where("email = ?", input.Email).First(&existingUser); result.RowsAffected > 0 {
		return utils.ErrorWithCode(c, fiber.StatusConflict, "EMAIL_EXISTS", "el email ya está registrado")
	}

	// Hash de la contraseña
	hashedPassword, err := utils.HashPassword(input.Password)
	if err != nil {
		return utils.ErrorWithCode(c, fiber.StatusInternalServerError, "HASH_ERROR", "error al procesar la contraseña")
	}

	// Obtener el rol de cliente por defecto
	var defaultRole models.Role
	if result := database.DB.Where("nombre = ?", "cliente").First(&defaultRole); result.Error != nil {
		// Si no existe el rol, crear usuario sin rol
		defaultRole.ID = 0
	}

	// Crear el usuario
	user := models.User{
		Email:           input.Email,
		Password:        hashedPassword,
		Nombre:          input.Nombre,
		Apellido:        input.Apellido,
		Telefono:        input.Telefono,
		TipoDocumento:   input.TipoDocumento,
		NumeroDocumento: input.NumeroDocumento,
		Activo:          true,
		EmailVerificado: false,
	}

	// Asignar rol de cliente si existe
	if defaultRole.ID > 0 {
		user.Roles = append(user.Roles, defaultRole)
	}

	if result := database.DB.Create(&user); result.Error != nil {
		return utils.ErrorWithCode(c, fiber.StatusInternalServerError, "CREATE_ERROR", "error al crear el usuario")
	}

	// Cargar roles para la respuesta
	database.DB.Preload("Roles").First(&user, user.ID)

	// Generar token JWT
	cfg := config.Load()
	tokenPair, err := utils.GenerateTokenPair(user.ID, user.Email, "cliente", cfg.JWTExpire, time.Hour*24*30)
	if err != nil {
		return utils.ErrorWithCode(c, fiber.StatusInternalServerError, "TOKEN_ERROR", "error al generar token")
	}

	return utils.Success(c, fiber.StatusCreated, "usuario registrado exitosamente", fiber.Map{
		"user":  user.ToResponse(),
		"token": tokenPair,
	})
}

// Login maneja el inicio de sesión
// POST /api/auth/login
func Login(c *fiber.Ctx) error {
	var input models.LoginInput

	if err := c.BodyParser(&input); err != nil {
		return utils.ErrorWithCode(c, fiber.StatusBadRequest, "INVALID_BODY", "formato de datos inválido")
	}

	// Validar estructura con go-playground validator
	if validationErrors := utils.ValidateStruct(input); validationErrors != nil {
		return utils.ErrorWithCode(c, fiber.StatusBadRequest, "VALIDATION_ERROR", utils.FirstValidationError(validationErrors))
	}

	// Buscar usuario
	var user models.User
	if result := database.DB.Preload("Roles.Permisos").Where("email = ?", input.Email).First(&user); result.Error != nil {
		return utils.ErrorWithCode(c, fiber.StatusUnauthorized, "INVALID_CREDENTIALS", "credenciales inválidas")
	}

	// Verificar si está activo
	if !user.Activo {
		return utils.ErrorWithCode(c, fiber.StatusForbidden, "USER_INACTIVE", "cuenta inactiva")
	}

	// Verificar si está bloqueado temporalmente
	if user.LockedUntil != nil && time.Now().Before(*user.LockedUntil) {
		return utils.ErrorWithCode(c, fiber.StatusForbidden, "USER_LOCKED", "cuenta temporalmente bloqueada")
	}

	// Verificar contraseña
	if !utils.CheckPassword(input.Password, user.Password) {
		// Incrementar intentos fallidos
		user.FailedLoginAttempts++

		// Bloquear después de 5 intentos fallidos
		if user.FailedLoginAttempts >= 5 {
			lockedUntil := time.Now().Add(15 * time.Minute)
			user.LockedUntil = &lockedUntil
			database.DB.Save(&user)
			return utils.ErrorWithCode(c, fiber.StatusForbidden, "USER_LOCKED", "cuenta bloqueada por demasiados intentos fallidos")
		}

		database.DB.Save(&user)
		return utils.ErrorWithCode(c, fiber.StatusUnauthorized, "INVALID_CREDENTIALS", "credenciales inválidas")
	}

	// Resetear intentos fallidos y actualizar último login
	user.FailedLoginAttempts = 0
	user.LockedUntil = nil
	now := time.Now()
	user.UltimoLogin = &now
	database.DB.Save(&user)

	// Log de auditoría de login exitoso
	utils.LogSimple(user.ID, user.Email, "login", "auth", user.ID, true, "login exitoso desde "+c.IP())

	// Obtener rol principal
	roleName := "cliente"
	if len(user.Roles) > 0 {
		roleName = user.Roles[0].Nombre
	}

	// Generar tokens
	cfg := config.Load()
	tokenPair, err := utils.GenerateTokenPair(user.ID, user.Email, roleName, cfg.JWTExpire, time.Hour*24*30)
	if err != nil {
		return utils.ErrorWithCode(c, fiber.StatusInternalServerError, "TOKEN_ERROR", "error al generar token")
	}

	return utils.Success(c, fiber.StatusOK, "login exitoso", fiber.Map{
		"user":  user.ToResponse(),
		"token": tokenPair,
	})
}

// RefreshToken renueva el token de acceso
// POST /api/auth/refresh
func RefreshToken(c *fiber.Ctx) error {
	var input models.RefreshTokenInput

	if err := c.BodyParser(&input); err != nil {
		return utils.ErrorWithCode(c, fiber.StatusBadRequest, "INVALID_BODY", "formato de datos inválido")
	}

	if input.RefreshToken == "" {
		return utils.ErrorWithCode(c, fiber.StatusBadRequest, "MISSING_TOKEN", "refresh token requerido")
	}

	// Validar refresh token
	claims, err := utils.ValidateJWT(input.RefreshToken)
	if err != nil {
		return utils.ErrorWithCode(c, fiber.StatusUnauthorized, "INVALID_TOKEN", "token inválido o expirado")
	}

	// Verificar si el refresh token ha sido revocado
	if utils.IsRevoked(claims.RegisteredClaims.ID) {
		return utils.ErrorWithCode(c, fiber.StatusUnauthorized, "TOKEN_REVOKED", "token revocado")
	}

	// Rotación de token: revocar el refresh token actual para prevenir replay attacks
	if claims.RegisteredClaims.ID != "" && claims.ExpiresAt != nil {
		utils.RevokeToken(claims.RegisteredClaims.ID, claims.ExpiresAt.Time)
	}

	// Buscar usuario
	var user models.User
	if result := database.DB.Preload("Roles.Permisos").First(&user, claims.UserID); result.Error != nil {
		return utils.ErrorWithCode(c, fiber.StatusNotFound, "USER_NOT_FOUND", "usuario no encontrado")
	}

	if !user.Activo {
		return utils.ErrorWithCode(c, fiber.StatusForbidden, "USER_INACTIVE", "usuario inactivo")
	}

	// Generar nuevos tokens
	cfg := config.Load()
	roleName := "cliente"
	if len(user.Roles) > 0 {
		roleName = user.Roles[0].Nombre
	}

	tokenPair, err := utils.GenerateTokenPair(user.ID, user.Email, roleName, cfg.JWTExpire, time.Hour*24*30)
	if err != nil {
		return utils.ErrorWithCode(c, fiber.StatusInternalServerError, "TOKEN_ERROR", "error al generar token")
	}

	return utils.Success(c, fiber.StatusOK, "token actualizado", fiber.Map{
		"token": tokenPair,
	})
}

// GetCurrentUser obtiene el usuario actual autenticado
// GET /api/auth/me
func GetCurrentUser(c *fiber.Ctx) error {
	user := middleware.GetUser(c)
	if user == nil {
		return utils.ErrorWithCode(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "no autorizado")
	}

	// Recargar con relaciones
	database.DB.Preload("Roles").First(&user, user.ID)

	return utils.SuccessData(c, fiber.StatusOK, user.ToResponse())
}

// ChangePassword cambia la contraseña del usuario
// POST /api/auth/change-password
func ChangePassword(c *fiber.Ctx) error {
	user := middleware.GetUser(c)
	if user == nil {
		return utils.ErrorWithCode(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "no autorizado")
	}

	var input models.ChangePasswordInput
	if err := c.BodyParser(&input); err != nil {
		return utils.ErrorWithCode(c, fiber.StatusBadRequest, "INVALID_BODY", "formato de datos inválido")
	}

	// Consultar contraseña actual directamente de la DB para evitar desincronización con caché en Redis
	var dbUser models.User
	if err := database.DB.Select("id", "password").First(&dbUser, user.ID).Error; err != nil {
		return utils.ErrorWithCode(c, fiber.StatusNotFound, "USER_NOT_FOUND", "usuario no encontrado")
	}

	// Verificar contraseña actual
	if !utils.CheckPassword(input.CurrentPassword, dbUser.Password) {
		return utils.ErrorWithCode(c, fiber.StatusBadRequest, "INVALID_PASSWORD", "contraseña actual incorrecta")
	}

	// Validar nueva contraseña
	if err := utils.ValidatePassword(input.NewPassword); err != nil {
		return utils.ErrorWithCode(c, fiber.StatusBadRequest, "WEAK_PASSWORD", err.Error())
	}

	// Verificar que las contraseñas nuevas coincidan
	if input.NewPassword != input.ConfirmPassword {
		return utils.ErrorWithCode(c, fiber.StatusBadRequest, "PASSWORD_MISMATCH", "las contraseñas no coinciden")
	}

	// Hash de la nueva contraseña
	hashedPassword, err := utils.HashPassword(input.NewPassword)
	if err != nil {
		return utils.ErrorWithCode(c, fiber.StatusInternalServerError, "HASH_ERROR", "error al procesar la contraseña")
	}

	now := time.Now()
	if result := database.DB.Model(&models.User{}).Where("id = ?", user.ID).Updates(map[string]interface{}{
		"password":            hashedPassword,
		"password_changed_at": &now,
	}); result.Error != nil {
		return utils.ErrorWithCode(c, fiber.StatusInternalServerError, "SAVE_ERROR", "error al guardar la contraseña")
	}

	database.InvalidateUserCache(user.ID)

	return utils.SuccessMessage(c, fiber.StatusOK, "contraseña cambiada exitosamente")
}

// Logout maneja el cierre de sesión (revoca el token JWT actual)
// POST /api/auth/logout
func Logout(c *fiber.Ctx) error {
	authHeader := c.Get("Authorization")
	if authHeader != "" {
		parts := strings.Split(authHeader, " ")
		if len(parts) == 2 && parts[0] == "Bearer" {
			// Revocar el token actual
			utils.RevokeTokenString(parts[1])
		}
	}
	return utils.SuccessMessage(c, fiber.StatusOK, "sesión cerrada exitosamente")
}

// UpdateProfile actualiza el perfil del usuario actual
// PUT /api/auth/profile
func UpdateProfile(c *fiber.Ctx) error {
	user := middleware.GetUser(c)
	if user == nil {
		return utils.ErrorWithCode(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "no autorizado")
	}

	var input models.UpdateProfileInput
	if err := c.BodyParser(&input); err != nil {
		return utils.ErrorWithCode(c, fiber.StatusBadRequest, "INVALID_BODY", "formato de datos inválido")
	}

	// Actualizar campos
	updates := make(map[string]interface{})

	if input.Nombre != "" {
		updates["nombre"] = input.Nombre
	}
	if input.Apellido != "" {
		updates["apellido"] = input.Apellido
	}
	if input.Telefono != "" {
		updates["telefono"] = input.Telefono
	}
	if input.TipoDocumento != "" {
		updates["tipo_documento"] = input.TipoDocumento
	}
	if input.NumeroDocumento != "" {
		updates["numero_documento"] = input.NumeroDocumento
	}
	if input.Direccion != "" {
		updates["direccion"] = input.Direccion
	}
	if input.Ciudad != "" {
		updates["ciudad"] = input.Ciudad
	}
	if input.Departamento != "" {
		updates["departamento"] = input.Departamento
	}
	if input.Pais != "" {
		updates["pais"] = input.Pais
	}
	if input.CodigoPostal != "" {
		updates["codigo_postal"] = input.CodigoPostal
	}
	if input.Foto != "" {
		updates["foto"] = input.Foto
	}

	if len(updates) > 0 {
		if result := database.DB.Model(user).Updates(updates); result.Error != nil {
			return utils.ErrorWithCode(c, fiber.StatusInternalServerError, "SAVE_ERROR", "error al actualizar el perfil")
		}
	}

	// Recargar usuario
	database.DB.Preload("Roles").First(&user, user.ID)

	database.InvalidateUserCache(user.ID)

	return utils.SuccessData(c, fiber.StatusOK, fiber.Map{
		"message": "perfil actualizado exitosamente",
		"user":    user,
	})
}

// UpdateTheme actualiza la preferencia de tema del usuario
// PUT /api/auth/theme
func UpdateTheme(c *fiber.Ctx) error {
	user := middleware.GetUser(c)
	if user == nil {
		return utils.ErrorWithCode(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "no autorizado")
	}

	var input struct {
		Theme string `json:"theme" validate:"required"`
	}
	if err := c.BodyParser(&input); err != nil {
		return utils.ErrorWithCode(c, fiber.StatusBadRequest, "INVALID_BODY", "formato de datos inválido")
	}

	if input.Theme != "light" && input.Theme != "dark" {
		return utils.ErrorWithCode(c, fiber.StatusBadRequest, "INVALID_THEME", "tema inválido (light/dark)")
	}

	if result := database.DB.Model(user).Update("theme", input.Theme); result.Error != nil {
		return utils.ErrorWithCode(c, fiber.StatusInternalServerError, "SAVE_ERROR", "error al actualizar el tema")
	}

	database.InvalidateUserCache(user.ID)

	return utils.SuccessMessage(c, fiber.StatusOK, "tema actualizado exitosamente")
}

// GetFullProfile obtiene el perfil completo del usuario actual
// GET /api/auth/profile
func GetFullProfile(c *fiber.Ctx) error {
	user := middleware.GetUser(c)
	if user == nil {
		return utils.ErrorWithCode(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "no autorizado")
	}

	// Recargar con relaciones
	database.DB.Preload("Roles").Preload("Roles.Permisos").First(&user, user.ID)

	return utils.SuccessData(c, fiber.StatusOK, user)
}

// Helpers
func isValidEmail(email string) bool {
	// Validación simple de email
	if len(email) < 3 || len(email) > 254 {
		return false
	}

	atIndex := -1
	for i := 0; i < len(email); i++ {
		if email[i] == '@' {
			atIndex = i
			break
		}
	}

	if atIndex <= 0 || atIndex == len(email)-1 {
		return false
	}

	// Verificar que tenga un punto después del @
	for i := atIndex; i < len(email); i++ {
		if email[i] == '.' && i < len(email)-1 {
			return true
		}
	}

	return false
}

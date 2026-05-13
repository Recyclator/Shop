package utils

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

// Validate es la instancia global del validador
var validate *validator.Validate

func init() {
	validate = validator.New()

	// Usar el nombre del campo JSON en los errores
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 1)[0]
		if name == "-" {
			return ""
		}
		return name
	})
}

// ValidateStruct valida un struct y retorna un mapa de errores (campo -> mensaje)
func ValidateStruct(s interface{}) map[string]string {
	err := validate.Struct(s)
	if err == nil {
		return nil
	}

	errors := make(map[string]string)
	for _, err := range err.(validator.ValidationErrors) {
		field := err.Field()
		var msg string
		switch err.Tag() {
		case "required":
			msg = fmt.Sprintf("el campo %s es requerido", field)
		case "email":
			msg = fmt.Sprintf("el campo %s debe ser un email v\u00e1lido", field)
		case "min":
			msg = fmt.Sprintf("el campo %s debe tener al menos %s caracteres", field, err.Param())
		case "max":
			msg = fmt.Sprintf("el campo %s debe tener m\u00e1ximo %s caracteres", field, err.Param())
		default:
			msg = fmt.Sprintf("validaci\u00f3n %s fallida para el campo %s", err.Tag(), field)
		}
		// Solo guardar el primer error por campo
		if _, exists := errors[field]; !exists {
			errors[field] = msg
		}
	}
	return errors
}

// ValidateAndRespond valida un struct y retorna un error Fiber si falla
func ValidateAndRespond(c *fiber.Ctx, s interface{}) error {
	errors := ValidateStruct(s)
	if errors != nil {
		return ErrorWithCode(c, fiber.StatusBadRequest, "VALIDATION_ERROR", "datos inv\u00e1lidos")
	}
	return nil
}

// FirstValidationError retorna el primer mensaje de error de un mapa de validaci\u00f3n
func FirstValidationError(errors map[string]string) string {
	for _, msg := range errors {
		return msg
	}
	return "datos inv\u00e1lidos"
}

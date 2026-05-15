package validator

import (
	"fmt"
	"strings"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type ValidationErr struct {
	Code   string            `json:"code"`
	Msg    string            `json:"message"`
	Fields map[string]string `json:"fields"`
}

func Bind(c *gin.Context, req any) bool {
	if err := c.ShouldBindJSON(req); err != nil {
		c.JSON(422, fromErr(err))
		return false
	}
	return true
}

func fromErr(err error) ValidationErr {
	fields := map[string]string{}
	if ve, ok := err.(validator.ValidationErrors); ok {
		for _, fe := range ve {
			fields[strings.ToLower(fe.Field())] = msg(fe)
		}
	}
	return ValidationErr{Code:"VALIDATION_ERROR", Msg:"datos inválidos", Fields:fields}
}

func msg(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required": return "requerido"
	case "email":    return "email inválido"
	case "min":      return fmt.Sprintf("mínimo %s", fe.Param())
	case "max":      return fmt.Sprintf("máximo %s", fe.Param())
	case "uuid":     return "UUID inválido"
	case "oneof":    return fmt.Sprintf("debe ser: %s", fe.Param())
	default:         return fmt.Sprintf("falla '%s'", fe.Tag())
	}
}

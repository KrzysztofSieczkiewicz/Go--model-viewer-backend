package models

import (
	"github.com/go-playground/validator"
)

// Validates Image fields against predefined regexp. Returns error on any field missing
func (m *Model) Validate() error {
	return validate.Struct(m)
}

func validateModelType(fl validator.FieldLevel) bool {
	return regexModelType.MatchString(fl.Field().String())
}

func validateModelDetailLevel(fl validator.FieldLevel) bool {
	return regexLOD.MatchString(fl.Field().String())
}

func validateModelExtension(fl validator.FieldLevel) bool {
	return regexModelExtension.MatchString(fl.Field().String())
}
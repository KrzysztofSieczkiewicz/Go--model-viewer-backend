package models

import (
	"regexp"

	"github.com/go-playground/validator"
)

var (
	regexModelType       = regexp.MustCompile(`^[a-zA-Z]+$`)
	regexLOD 			 = regexp.MustCompile(`^LOD[0-9]$`)
	regexModelExtension  = regexp.MustCompile(`^(obj|glTF)$`)
)

// Validates Image fields against predefined regexp. Returns error on any field missing
func (m *Model) Validate() error {
	validate := validator.New()

	validate.RegisterValidation("collection", validateCollection)

	validate.RegisterValidation("type", validateModelType)
	validate.RegisterValidation("resolution", validateModelDetailLevel)
	validate.RegisterValidation("extension", validateModelExtension)

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
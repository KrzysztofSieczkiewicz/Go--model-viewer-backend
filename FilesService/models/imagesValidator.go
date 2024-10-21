package models

import (
	"regexp"

	"github.com/go-playground/validator"
)

var (
	regexType       = regexp.MustCompile(`^[a-zA-Z]+$`)
	regexResolution = regexp.MustCompile(`^\d{3,4}x\d{3,4}$`)
	regexExtension  = regexp.MustCompile(`^(jpg|jpeg|png|gif|bmp|tiff)$`)
)

// Validates Image fields against predefined regexp. Returns error on any field missing
func (i *Image) Validate() error {
	validate := validator.New()

	validate.RegisterValidation("category", validateFilepath)
	validate.RegisterValidation("id", validateID)

	validate.RegisterValidation("type", validateImageType)
	validate.RegisterValidation("resolution", validateImageResolution)
	validate.RegisterValidation("extension", validateImageExtension)

	return validate.Struct(i)
}

func validateImageType(fl validator.FieldLevel) bool {
	return regexType.MatchString(fl.Field().String())
}

func validateImageResolution(fl validator.FieldLevel) bool {
	return regexResolution.MatchString(fl.Field().String())
}

func validateImageExtension(fl validator.FieldLevel) bool {
	return regexExtension.MatchString(fl.Field().String())
}

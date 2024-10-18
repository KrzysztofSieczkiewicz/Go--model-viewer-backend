package models

import (
	"regexp"

	"github.com/go-playground/validator"
)

// Validates Image fields against predefined regexp. Returns error on any field missing
func (i *Image) Validate() error {
	validate := validator.New()

	validate.RegisterValidation("type", validateImageType)
	validate.RegisterValidation("resolution", validateImageResolution)
	validate.RegisterValidation("extension", validateImageExtension)

	return validate.Struct(i)
}

func validateImageType(fl validator.FieldLevel) bool {
	re := regexp.MustCompile(`^[a-zA-Z]+$`)
	matches := re.FindAllString(fl.Field().String(), -1)

	return len(matches) == 1
}

func validateImageResolution(fl validator.FieldLevel) bool {
	re := regexp.MustCompile(`^\d{3,4}x\d{3,4}$`)
	matches := re.FindAllString(fl.Field().String(), -1)

	return len(matches) == 1
}

func validateImageExtension(fl validator.FieldLevel) bool {
	re := regexp.MustCompile(`^\.[a-zA-Z]{1,5}$`)
	matches := re.FindAllString(fl.Field().String(), -1)

	return len(matches) == 1
}

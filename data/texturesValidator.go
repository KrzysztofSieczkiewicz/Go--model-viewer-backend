package data

import (
	"regexp"

	validator "github.com/go-playground/validator/v10"
)

func (t *Texture) Validate() error {
	validate := validator.New()

	validate.RegisterValidation("filepath", validateFilePath)
	validate.RegisterValidation("name", validateName)

	return validate.Struct(t)
}

func validateFilePath(fl validator.FieldLevel) bool {
	re := regexp.MustCompile(`^(.*)\/([^\/]*)$`)
	matches := re.FindAllString(fl.Field().String(), -1)

	return len(matches) == 1
}

func validateName(fl validator.FieldLevel) bool {
	re := regexp.MustCompile(`^([a-zA-Z -_]*([_][0-9]*)?)$`)
	matches := re.FindAllString(fl.Field().String(), -1)

	return len(matches) == 1
}
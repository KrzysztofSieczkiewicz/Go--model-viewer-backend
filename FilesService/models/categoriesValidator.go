package models

import (
	"regexp"

	"github.com/go-playground/validator"
)

var (
	regexFilepath = regexp.MustCompile(`^(.*)\/([^\/]*)$`)
)

func (i *Category) Validate() error {
	validate := validator.New()

	validate.RegisterValidation("filepath", validateCategoryFilepath)

	return validate.Struct(i)
}

func validateCategoryFilepath(fl validator.FieldLevel) bool {
	return regexFilepath.MatchString(fl.Field().String())
}

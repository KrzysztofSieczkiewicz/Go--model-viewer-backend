package models

import (
	"github.com/go-playground/validator"
)

func (i *Category) Validate() error {
	return validate.Struct(i)
}

func validateCategoryFilepath(fl validator.FieldLevel) bool {
	return regexFilepath.MatchString(fl.Field().String())
}

package models

import (
	"github.com/go-playground/validator"
)

func (c *Collection) Validate() error {
	return validate.Struct(c)
}

func validateID(fl validator.FieldLevel) bool {
	return regexID.MatchString(fl.Field().String())
}

func validateCategory(fl validator.FieldLevel) bool {
	category := fl.Field().Interface().(Category)

	err := category.Validate()
	return err == nil
}

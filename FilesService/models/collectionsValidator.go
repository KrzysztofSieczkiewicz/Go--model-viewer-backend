package models

import (
	"regexp"

	"github.com/go-playground/validator"
)

var (
	regexID = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_]{1,63}$`) // don't allow '_' as the first character
)

func (c *Collection) Validate() error {
	validate := validator.New()

	validate.RegisterValidation("id", validateID)
	validate.RegisterValidation("category", validateCategory)

	return validate.Struct(c)
}

func validateID(fl validator.FieldLevel) bool {
	return regexID.MatchString(fl.Field().String())
}

func validateCategory(fl validator.FieldLevel) bool {
	category := fl.Field().Interface().(*Category)

	err := category.Validate()
	return err == nil
}

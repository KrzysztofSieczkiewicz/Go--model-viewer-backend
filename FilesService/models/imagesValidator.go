package models

import (
	"github.com/go-playground/validator"
)

// Validates Image fields against predefined regexp. Returns error on any field missing
func (i *Image) Validate() error {
	return validate.Struct(i)
}

func validateImageType(fl validator.FieldLevel) bool {
	return regexImageType.MatchString(fl.Field().String())
}

func validateImageResolution(fl validator.FieldLevel) bool {
	return regexImageResolution.MatchString(fl.Field().String())
}

func validateImageExtension(fl validator.FieldLevel) bool {
	return regexImageExtension.MatchString(fl.Field().String())
}

func validateCollection(fl validator.FieldLevel) bool {
	collection := fl.Field().Interface().(*Collection)

	err := collection.Validate()
	return err == nil
}
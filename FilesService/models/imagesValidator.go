package models

import (
	"regexp"

	"github.com/go-playground/validator"
)

var (
	regexImageType       = regexp.MustCompile(`^[a-zA-Z]+$`)
	regexImageResolution = regexp.MustCompile(`^\d{3,4}x\d{3,4}$`)
	regexImageExtension  = regexp.MustCompile(`^(jpg|jpeg|png|gif|bmp|tiff)$`)
)

// Validates Image fields against predefined regexp. Returns error on any field missing
func (i *Image) Validate() error {
	validate := validator.New()

	validate.RegisterValidation("collection", validateCollection)

	validate.RegisterValidation("type", validateImageType)
	validate.RegisterValidation("resolution", validateImageResolution)
	validate.RegisterValidation("extension", validateImageExtension)

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
package models

import (
	"regexp"

	"github.com/go-playground/validator"
)

var (
	// prevents first character from being an underscore
	regexID = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_]{1,63}$`)
	regexCategory = regexp.MustCompile(`^(.*)\/([^\/]*)$`)
)

func (c *AssetsCollection) Validate() error {
	validate := validator.New()

	//validate.RegisterValidation("id", validateID)
	//validate.RegisterValidation("category", validateCategory)

	return validate.Struct(c)
}
/*
// DELETE
func (is *ImageSet) Validate() error {
	validate := validator.New()

	validate.RegisterValidation("filepath", validateID)
	validate.RegisterValidation("name", validateCategory)

	return validate.Struct(is)
}
// END DELETE

func validateID(fl validator.FieldLevel) bool {
	return regexID.MatchString(fl.Field().String())
}

func validateCategory(fl validator.FieldLevel) bool {
	return regexCategory.MatchString(fl.Field().String())
}
*/
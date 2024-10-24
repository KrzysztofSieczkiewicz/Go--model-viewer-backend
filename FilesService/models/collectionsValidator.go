package models

import (
	"regexp"

	"github.com/go-playground/validator"
)

// TODO: rework

var (
	regexID = regexp.MustCompile(`^[a-zA-Z0-9]{2,64}$`)
	regexCategory = regexp.MustCompile(`^(.*)\/([^\/]*)$`)
)

// Validates ImageSet fields against predefined regexp. Returns error on any field missing
func (is *ImageSet) Validate() error {
	validate := validator.New()

	validate.RegisterValidation("filepath", validateID)
	validate.RegisterValidation("name", validateCategory)
	//validate.RegisterValidation("images", validateImages)

	return validate.Struct(is)
}

func validateID(fl validator.FieldLevel) bool {
	return regexID.MatchString(fl.Field().String())
}

func validateCategory(fl validator.FieldLevel) bool {
	return regexCategory.MatchString(fl.Field().String())
}

/*
func validateImages(fl validator.FieldLevel) bool {
	images, err := fl.Field().Interface().([]*Image)
	if err {
		return false
	}

	// Allow empty images slice
	if len(images) == 0 {
		images = []*Image{}
	}

	for _, img := range images {
        err := img.Validate()
		if err != nil {
			return false
		}
	}

	return true
}
*/
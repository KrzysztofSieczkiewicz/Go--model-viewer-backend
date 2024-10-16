package data

import (
	"regexp"

	"github.com/go-playground/validator"
)

// Validates ImageSet fields against predefined regexp. Returns error on any field missing
func (is *ImageSet) Validate() error {
	validate := validator.New()

	validate.RegisterValidation("filepath", validateID)
	validate.RegisterValidation("name", validateCategory)
	validate.RegisterValidation("Images", validateImages)

	return validate.Struct(is)
}

func validateID(fl validator.FieldLevel) bool {
	re := regexp.MustCompile(`^[a-zA-Z0-9]{2,64}$`)
	matches := re.FindAllString(fl.Field().String(), -1)

	return len(matches) == 1
}

func validateCategory(fl validator.FieldLevel) bool {
	re := regexp.MustCompile(`^(.*)\/([^\/]*)$`)
	matches := re.FindAllString(fl.Field().String(), -1)

	return len(matches) == 1
}

func validateImages(fl validator.FieldLevel) bool {
	images, err := fl.Field().Interface().([]*Image)
	if err {
		return false
	}

	// Allow empty slice
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
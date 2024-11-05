package models

import (
	"regexp"

	"github.com/go-playground/validator"
)

var (
	validate      *validator.Validate
	regexFilepath = regexp.MustCompile(`^([a-zA-Z ]+)(\/[a-zA-Z ]+)*$`)
	regexID = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_]{1,63}$`) // don't allow '_' as the first character

	regexImageType       = regexp.MustCompile(`^[a-zA-Z]+$`)
	regexImageResolution = regexp.MustCompile(`^\d{3,4}x\d{3,4}$`)
	regexImageExtension  = regexp.MustCompile(`^(jpg|jpeg|png|gif|bmp|tiff)$`)

	regexModelType       = regexp.MustCompile(`^[a-zA-Z]+$`)
	regexLOD 			 = regexp.MustCompile(`^LOD[0-9]$`)
	regexModelExtension  = regexp.MustCompile(`^(obj|glTF)$`)
)

func init() {
	validate = validator.New()

	// CATEGORY AND COLLECTION
	validate.RegisterValidation("categoryPath", validateCategoryFilepath)
	validate.RegisterValidation("collectionID", validateID)
	validate.RegisterValidation("category", validateCategory)

	// IMAGE
	validate.RegisterValidation("imgType", validateImageType)
	validate.RegisterValidation("resolution", validateImageResolution)
	validate.RegisterValidation("imgExtension", validateImageExtension)
	validate.RegisterValidation("collection", validateCollection)
}
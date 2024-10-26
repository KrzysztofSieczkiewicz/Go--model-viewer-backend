package models

import (
	"fmt"
	"path/filepath"
	"strings"
)

// Defines a properties of a image file that are used to construct/deconstruct filename
// swagger:model Image
type Image struct {
	// Image parent collection
	Collection AssetsCollection	`json:"collection" validate:"required"`

	// Image type determining general image purpose (eg. Albedo, Roughness)
	// required: true
	ImgType string `json:"type" validate:"required"`

	// required: true
	Resolution string `json:"resolution" validate:"required"`

	// required: true
	FileExtension string `json:"extension" validate:"required"`
}

// Returns filename string from image properties
func (i *Image) ConstructName() string {
	return fmt.Sprintf(
		"%s_%s.%s",
		i.ImgType,
		i.Resolution,
		i.FileExtension,
	)
}

// Combines image properties into filepath
func (i *Image) ConstructFilepath() string {
	return filepath.Join(
		i.Collection.ConstructCollectionPath(),
		i.ConstructName(),
	)
}

// Deconstructs image properties from the filename
func (i *Image) DeconstructName(filename string) error {
	// retrieve file extension
	parts := strings.Split(filename, ".")
	if len(parts) != 2 {
		return fmt.Errorf("invalid filename format - unable to retrieve file extension: %s", filename)
	}

	// retrieve type and extension
	typeAndRes := strings.Split(parts[0], "_")
	if len(typeAndRes) != 2 {
		return fmt.Errorf("invalid filename - unable to retrieve image type and resolution: %s", parts[1])
	}

	i.ImgType = typeAndRes[0]
	i.Resolution = typeAndRes[1]
	i.FileExtension = parts[1]

	return nil
}
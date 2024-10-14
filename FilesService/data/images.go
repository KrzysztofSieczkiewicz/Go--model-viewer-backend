package data

import (
	"fmt"
	"strings"
)

// Image defines a properties of a image file that are used to construct filename
// swagger:model Image
type Image struct {
	// Image type determining general image purpose (eg. Albedo, Roughness)
	// required: true
	// min length: 2
	// max length: 64
	ImgType       string `json:"type"`

	// required: true
	// min length: 7
	// max length: 16
	Resolution    string `json:"resolution"`
	
	// required: true
	// min length: 2
	// max length: 8
	FileExtension string `json:"extension"`
}

type Images []*Image

// Returns filename string from image properties
func (i *Image) ConstructImageName() string {
	return fmt.Sprintf(
		"%s_%s.%s",
		i.ImgType,
		i.Resolution,
		i.FileExtension,
	)
}

// Deconstructs image properties from the filename
func (i *Image) DeconstructImageName(filename string) error {
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

// Deconstruct slice of filenames into slice of Images
func (i *Images) DeconstructImageNames(filenames []string) error {
	// Iterate over each filename and deconstruct it
	for _, filename := range filenames {
		image := &Image{}
		err := image.DeconstructImageName(filename)
		if err != nil {
			return err
		}
		*i = append(*i, image)
	}

	return nil
}
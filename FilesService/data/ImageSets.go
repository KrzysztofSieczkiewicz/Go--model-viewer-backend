package data

import (
	"fmt"
	"strings"
)

type ImageSet struct {
	ID       string  `json:"id"`
	Category string  `json:"category"`
	Images   []*Image `json:"images,omitempty"`
}

type Image struct {
	ImgType       string `json:"type"`
	Resolution    string `json:"resolution"`
	FileExtension string `json:"extension"`
}

// Returns filename string from image properties
func (i *Image) ConstructImageName() string {
	return fmt.Sprintf(
		"%s_%s.%s",
		i.ImgType,
		i.Resolution,
		i.FileExtension,
	)
}

type Images []*Image


// Deconstructs image properties from the filename
func (i *Image) DeconstructImageName(filename string) (error) {
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
func (i *Images) DeconstructImageNames(filenames []string) (error) {
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
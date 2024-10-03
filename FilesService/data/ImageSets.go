package data

import "fmt"

type ImageSet struct {
	ID       string  `json:"id"`
	Category string  `json:"category"`
	Images   []Image `json:"images,omitempty"`
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

// Deconstructs image properties from the filename
func (i *Image) DeconstructImageName() (*Image, error) {

	return nil, nil
}

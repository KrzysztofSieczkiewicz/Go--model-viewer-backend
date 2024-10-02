package data

import "fmt"

type ImageSet struct {
	ID       string  `json:"id"`
	Category string  `json:"category"`
	Images   []Image `json:"images"`
}

type Image struct {
	ImgType       string `json:"type"`
	Resolution    string `json:"resolution"`
	FileExtension string `json:"extension"`
}

func (i *Image) GetImageName() string {
	return fmt.Sprintf(
		"%s_%s.%s",
		i.ImgType,
		i.Resolution,
		i.FileExtension,
	)
}

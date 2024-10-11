package data

type ImageSet struct {
	ID       string  `json:"id"`
	Category string  `json:"category"`
	Images   []*Image `json:"images,omitempty"`
}
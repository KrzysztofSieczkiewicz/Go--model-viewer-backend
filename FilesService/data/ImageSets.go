package data

// ImageSet defines a properties of a set of images contributing to an entire texture with various resolutions or image types
// swagger:model imageSet
type ImageSet struct {
	// ID as it is stored in the database
	// min length: 2
	// max length: 64
	ID       string  `json:"id,omitempty"`

	// Category determining storage subdirectory
	// min length: 3
	// max length: 128
	Category string  `json:"category,omitempty"`
	
	// List of images contriburing to this particular imageSet
	Images   []*Image `json:"images,omitempty"`
}
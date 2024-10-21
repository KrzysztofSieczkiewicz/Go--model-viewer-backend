package models

// ImageSet defines a properties of a set of images contributing to an entire texture with various resolutions or image types
// swagger:model imageSet
type ImageSet struct {
	// ID as it is stored in the database
	ID       string  `json:"id," validate:"required"`

	// Category determining storage subdirectory
	Category string  `json:"category" validate:"required"`
}

// PutImageSetRequest defines combination of initial imageset and the new properties that it should be updated to
// swagger:model updateImageSet
type PutImageSetRequest struct {
	// Current image set properties
	Existing	ImageSet	`json:"existing" validate:"required"`

	// Desired image set properties
	New			ImageSet 	`json:"new" validate:"required"`
}
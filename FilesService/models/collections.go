package models

// Defines a set of files contributing to the same object/texture
// swagger:model collection
type Collection struct {
	// Collection ID
	ID string `json:"id" validate:"required"`

	// Category structure describing collection
	Category string `json:"category" validate:"required"`
}

// Defines properties of current collection and properties that are to be changed to
// swagger:model putCollectionRequest
type PutCollectionRequest struct {
	// Current properties
	Existing Collection `json:"existing" validate:"required"`

	// Desired properties
	New Collection `json:"new" validate:"required"`
}

// Defines a response to GET collection request
// swagger:model getCollectionResponse
type GetCollectionResponse struct {
	// Collection contents
	Contents	[]CollectionContent	`json:"contents" validate:"required"`
}

/*
func (c *CollectionContent) Validate() error {
	switch c.FileType {
	case FileTypeDirectory, FileTypeFile:
		return nil
	default:
		return errors.New("invalid FileType, must be 'directory' or 'file'")
	}
}
*/

// ImageSet defines a properties of a set of images contributing to an entire texture with various resolutions or image types
// swagger:model imageSet
type ImageSet struct {
	// ID as it is stored in the database
	ID string `json:"id" validate:"required"`

	// Category determining storage subdirectory
	Category string `json:"category" validate:"required"`
}

// PutImageSetRequest defines combination of initial imageset and the new properties that it should be updated to
// swagger:model updateImageSet
type PutImageSetRequest struct {
	// Current image set properties
	Existing ImageSet `json:"existing" validate:"required"`

	// Desired image set properties
	New ImageSet `json:"new" validate:"required"`
}
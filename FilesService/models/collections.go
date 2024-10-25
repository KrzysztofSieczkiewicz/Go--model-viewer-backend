package models

import "path/filepath"

// Defines a set of files contributing to the same object/texture
// swagger:model collection
type AssetsCollection struct {
	// Collection ID
	ID string `json:"id" validate:"required"`

	// Category structure describing collection
	Category Category `json:"category" validate:"required"`
}

func (c *AssetsCollection) ConstructCollectionPath() string {
	return filepath.Join(
		c.Category.ConstructCategoryPath(), 
		c.ID,
	)
}



// Defines properties of current collection and properties that are to be changed to
// swagger:model putCollectionRequest
type PutCollectionRequest struct {
	// Current properties
	Existing AssetsCollection `json:"existing" validate:"required"`

	// Desired properties
	New AssetsCollection `json:"new" validate:"required"`
}

func (c *AssetsCollection) Collection() *AssetsCollection {
	return c
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
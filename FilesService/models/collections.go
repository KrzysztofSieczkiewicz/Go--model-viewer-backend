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
package models

import "path/filepath"

// Defines a set of files contributing to the same object/texture
// swagger:model collection
type Collection struct {
	// Collection ID
	ID string `json:"id" validate:"required"`

	// Category structure describing collection
	Category Category `json:"category" validate:"required"`
}

// Returns filename string from collection properties
func (c *Collection) ConstructCollectionPath() string {
	return filepath.Join(
		c.Category.ConstructCategoryPath(), 
		c.ID,
	)
}
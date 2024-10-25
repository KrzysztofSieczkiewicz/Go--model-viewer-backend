package models

import (
	"errors"
	"path/filepath"
	"strings"
)

// Category defines a filepath of given category
// swagger:model category
type Category struct {
	Filepath	string	`json:"filepath" validate:"required"`
}

// Converts the collection properties to the filesystem compliant category path
func (c *Category) ConstructCategoryPath() string {
	dirs := strings.Split(c.Filepath, string(filepath.Separator))

	for i, dir := range dirs {
		dirs[i] = "_" + dir
	}

	newPath := strings.Join(dirs, string(filepath.Separator))

	return newPath
}

// Converts the provided filesystem compliant category path to the collection properties
func (c *Category) DeconstructCategoryPath() error {
	dirs := strings.Split(c.Filepath, string(filepath.Separator))

	for i, dir := range dirs {
		if strings.HasPrefix(dir, "_") {
			dirs[i] = dir[1:]
		} else {
			return errors.New("unable to decostruct category path")
		}
	}

	c.Filepath = strings.Join(dirs, string(filepath.Separator))
	return nil
}

// PutCategoryRequest defines combination of initial category filepath and the new filepath it should be updated to
// swagger:model updateCategory
type PutCategoryRequest struct {
	// Current image set properties
	Existing	Category	`json:"existing" validate:"required"`

	// Desired image set properties
	New			Category 	`json:"new" validate:"required"`
}
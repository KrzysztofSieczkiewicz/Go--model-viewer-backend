package models

import (
	"errors"
	"path/filepath"
	"strings"
)

// Category defines a filepath of given category
// swagger:model category
type Category struct {
	Path	string	`json:"path" validate:"required"`
}

// Converts the collection properties to the filesystem compliant category path
func (c *Category) ConstructCategoryPath() string {
	if c.Path == "" {
		return ""
	}

	// TODO: Extend this method so it handles separators with more flexibility
	dirs := strings.Split(c.Path, "/")
	for i, dir := range dirs {
		if dir != "" {
			dirs[i] = "_" + dir
		}
	}

	newPath := strings.Join(dirs, string(filepath.Separator))

	return newPath
}

// Converts the provided filesystem compliant category path to the collection properties
func (c *Category) DeconstructCategoryPath() error {
	dirs := strings.Split(c.Path, string(filepath.Separator))

	for i, dir := range dirs {
		if strings.HasPrefix(dir, "_") {
			dirs[i] = dir[1:]
		} else {
			return errors.New("unable to decostruct category path")
		}
	}

	c.Path = strings.Join(dirs, string(filepath.Separator))
	return nil
}
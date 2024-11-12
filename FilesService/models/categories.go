package models

import (
	"path/filepath"
	"strings"
)

// Category defines a filepath of given category
// swagger:model category
type Category struct {
	Path	string	`json:"path" validate:"required,categoryPath"`
}

// Converts the collection properties to the filesystem compliant category path
func (c *Category) ConstructCategoryPath() string {
	if c.Path == "" {
		return ""
	}

	dirs := strings.Split(c.Path, "/")
	for i, dir := range dirs {
		if dir != "" {
			dirs[i] = "_" + dir
		}
	}

	newPath := strings.Join(dirs, string(filepath.Separator))

	return newPath
}

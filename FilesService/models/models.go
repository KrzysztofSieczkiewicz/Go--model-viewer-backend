package models

import (
	"fmt"
	"path/filepath"
	"strings"
)

// Defines a properties of a 3D file (model) that are used to construct/deconstruct filename
// swagger:model Model
type Model struct {
	// Model parent collection
	Collection *Collection	`json:"collection" validate:"required,collection"`

	// Type of the model
	ModelType		string	`json:"type" validate:"required,modelType"`

	// Level of Detail of the model
	LOD				string	`json:"lod" validate:"required,lod"`

	// file extension
	FileExtension 	string	`json:"extension" validate:"required,modelExtension"` 
}

// Returns filename string from asset properties
func (m *Model) ConstructName() string {
	return fmt.Sprintf(
		"%s_%s.%s",
		m.ModelType,
		m.LOD,
		m.FileExtension,
	)
}

func (m *Model) ConstructFilepath() string {
	return filepath.Join(
		m.Collection.ConstructCollectionPath(),
		m.ConstructName(),
	)
}

// Deconstructs asset properties from the filename
func (m *Model) DeconstructName(filename string) error {
	// retrieve file extension
	parts := strings.Split(filename, ".")
	if len(parts) != 2 {
		return fmt.Errorf("invalid filename format - unable to retrieve file extension: %s", filename)
	}

	// retrieve type and LOD
	typeAndLOD := strings.Split(parts[0], "_")
	if len(typeAndLOD) != 2 {
		return fmt.Errorf("invalid filename - unable to retrieve image type and resolution: %s", parts[1])
	}

	m.ModelType = typeAndLOD[0]
	m.LOD = typeAndLOD[1]
	m.FileExtension = parts[1]

	return nil
}
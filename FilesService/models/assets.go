package models

import (
	"fmt"
	"strings"
)

// Defines a properties of a 3D file (asset) that are used to construct/deconstruct filename
// swagger:model Asset
type Asset struct {
	Collection

	// Type of an asset
	AssetType		string	`json:"type" validate:"required"`

	// Level of Detail of the asset
	LOD				string	`json:"lod" validate:"required"`

	// extension of the file
	FileExtension 	string	`json:"extension" validate:"required"` 
}

// Returns filename string from asset properties
func (a *Asset) ConstructName() string {
	return fmt.Sprintf(
		"%s_%s.%s",
		a.AssetType,
		a.LOD,
		a.FileExtension,
	)
}

// Deconstructs asset properties from the filename
func (a *Asset) DeconstructName(filename string) error {
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

	a.AssetType = typeAndLOD[0]
	a.LOD = typeAndLOD[1]
	a.FileExtension = parts[1]

	return nil
}
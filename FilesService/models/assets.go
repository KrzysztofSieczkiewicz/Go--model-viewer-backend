package models

// Asset represents an interface for asset (model/image) management.
type Asset interface {
	// Validate checks if the asset is valid and returns an error if not.
	Validate() error

	// ConstructName combines asset properties into a filesystem-friendly filename.
	ConstructName() string

	// DeconstructName retrieves asset properties from a given filename.
	DeconstructName(filename string) error

	// ConstructFilepath combines asset and collection properties into a filesystem-compliant filepath.
	ConstructFilepath() string
}
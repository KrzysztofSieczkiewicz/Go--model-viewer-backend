package models

type Asset interface {
	Validate() error

	// Combine asset properties into filesystem-friendly filename
	ConstructName() string

	// Retrieve asset properties from filename
	DeconstructName(filename string) error

	// Combine asset and collection properties into filesystem-compliant filepath
	ConstructFilepath() string

	// Get assets parent collection
	Collection() *AssetsCollection
}
package models

type PutRequest[T any] struct {
	// Current properties
	Existing T `json:"existing" validate:"required"`

	// Desired properties
	New T `json:"new" validate:"required"`
}
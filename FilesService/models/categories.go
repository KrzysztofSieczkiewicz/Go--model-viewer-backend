package models

// Category defines a filepath of given category
// swagger:model category
type Category struct {
	Filepath	string	`json:"filepath" validate:"required"`
}

// PutCategoryRequest defines combination of initial category filepath and the new filepath it should be updated to
// swagger:model updateCategory
type PutCategoryRequest struct {
	// Current image set properties
	Existing	Category	`json:"existing" validate:"required"`

	// Desired image set properties
	New			Category 	`json:"new" validate:"required"`
}
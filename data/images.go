package data

// Image defines a structure for an API texture
// swagger:model Image
type Image struct {
	// Unique ID identifying the image in the database
	// required: false
	// min length: 8
	// max length: 255
	ID	string	`json:"id,omitempty"`

	// Image name as displayed to the end-user
	// required: true
	// min length: 3
	// max length: 255
	Name	string	`json:"name" validate:"required"`

	ImgTypes	[]string	`json:"types"`
	ImgResolutions	[]string	`json:"resolutions"`

	// Filepath under which all associated files can be found
	// required: true
	// pattern: ^(.*)\/([^\/]*)$
	FilePath	string	`json:"-" validate:"required"`
}

type ImageType struct {
	Type	string	`json:"type"`
	AvailableSizes	[]string	`json:"sizes"`	
}


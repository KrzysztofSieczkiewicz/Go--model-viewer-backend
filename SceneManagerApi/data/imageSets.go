package data

import "github.com/KrzysztofSieczkiewicz/go--model-viewer-backend/SceneManagerApi/internal/types"

// ImageSet defines a structure for a set of images that combine to single texture
// swagger:model ImageSet
type ImageSet struct {
	// Unique ID identifying the imageSet in the database
	// required: false
	// min length: 8
	// max length: 255
	ID	string	`json:"id,omitempty"`

	// ImageSet name as displayed to the end-user
	// required: true
	// min length: 3
	// max length: 255
	Name	string	`json:"name" validate:"required"`

	// Url pointing to the 

	// Image type contains type to display only images relevant to requesting field
	// required: true
	ImgTypes	[]types.Image	`json:"types"`

	// Image resolutions determine which quality levels are available for given texture
	// required: true
	ImgResolutions	[]types.Resolution	`json:"resolutions"`

	// Filepath under which all associated images can be found
	// required: true
	// pattern: ^(.*)\/([^\/]*)$
	FilePath	string	`json:"-"`
}

type ImageSets []*ImageSet


func GetImageSets() ImageSets {
	return imageSetsList
}

func GetImageSetsQueried(types []types.Image, resolutions []types.Resolution) ImageSets {


	return imageSetsList
}


func filterImageSets() {
	
}


var imageSetsList = ImageSets{
	{
		ID: "1",
		Name: "Pear",
		ImgTypes: []types.Image{
			types.AmbientOcclusionMap,
			types.ColorMap, 
			types.DisplacementMap,
			types.NormalMap,
			types.RoughnessMap,
		},
		ImgResolutions: []types.Resolution{
			types.Resolution2048,
			types.Resolution4096,
		},
		FilePath: "./textures/",
	},
}
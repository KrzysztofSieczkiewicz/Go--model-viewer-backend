package data

import (
	"fmt"
	"log"
	"time"

	gonanoid "github.com/matoous/go-nanoid"
)

// Texture defines the structure for an API texture
// swagger:model Texture
type Texture struct {
	// Unique id identifying texture in the database
	// required: false
	// min length: 8
	// max length: 255
	ID        string    `json:"id,omitempty"`

	// Texture name for identification by the end-user
	// required: true
	// min length: 8
	// max length: 255
	// pattern: ^([a-zA-Z -_]*([_][0-9]*)?)$
	Name      string    `json:"name" validate:"required,name"`

	// Filepath under which the texture can be found in the filesystem
	// required: true
	// min length: 8
	// max length: 255
	// pattern: ^(.*)\/([^\/]*)$
	FilePath  string    `json:"path" validate:"required,filepath"`

	// Tags roughly describing the texture properties
	// required: false
	Tags      []string  `json:"tags"`
	CreatedOn time.Time `json:"-"`
	UpdatedOn time.Time `json:"-"`

	// Fields that should be created:
	// Type - what kind of texture is this - Physical/Standard?
	// QualityLevel - functions just like LOD
	// Properties - object depending on texture type
	// Properties.stuff - most numerical properties
	// Properties.imageId - id of each image based property - if null, there is no texture

	// User can have:
	// Scanned assets - either preset texture or editable
	// Primitives and 3D models - only editable

	// Editable texture have some different types which differ by amount of properties and images
	// You have the following texture types:
	// Phong
	// Physical
	// Standard
	// others

	// Each texture type consists from few properties interchangeable with images.
	// Image can have subtypes (normal/color/roughness) etc, but that doesn't need to be visible to the end user
	// So each image set can be described with three properties:
	// NAME_TYPE_SIZE
	// which can be further reduced by naming a folder with imageId:
	// imageId/TYPE_SIZE
	// Where user by himself controls only the id (indirectly as he's choosing by name and thumbnail). 
	// Type is depending on field type and size is controlled by quality level
	
	// So You can save an Image as an object with property of ID, name and filepath to the folder where all images are being stored (optionally - list available sizes)
	// keeping in mind that images must follow certain naming convention and limiting files manipulation to internalApi with admin privilleges
	// would create a solution where database is always consistent with filesystem

	// In that case:
	// TODO: Create a data model for image
	// TODO: Create a folder and example image files including: thumbnail, 512x512, 1024x1024, 2048x2048
	// TODO: Create a handler (consider mocking or creating a basic database)
	// TODO: Handle images by separate subrouter
	// TODO: Make some requests and test if files are properly being returned
	// TODO: Move to textures enpoint remodelling
}


type Textures []*Texture

func GetTextures() Textures {
	return texturesList
}

func GetTexture(id string) (*Texture, error){
	texture, _, err := findTexture(id)
	if err != nil {
		return nil, err
	}

	return texture, nil
}

// TODO: how to add texture so 'tags' is not null, but empty (is there a point though?)
func AddTexture(t *Texture) {
	t.ID = getNextID()
	texturesList = append(texturesList, t)
}

func UpdateTexture(id string, t *Texture) error {
	_, index, err := findTexture(id)
	if err != nil {
		return err
	}

	t.ID = id
	texturesList[index] = t

	return nil
}

func DeleteTexture(id string) error {
	_, index, err := findTexture(id)
	if err != nil {
		return err
	}

	texturesList = append(
		texturesList[:index], 
		texturesList[index+1:]...
	)
	
	return nil
}

func findTexture(id string) (*Texture, int, error) {
	for i, t := range texturesList {
		if t.ID == id {
			return t, i, nil
		}
	}

	return nil, -1, ErrTextureNotFound
}

func getNextID() string {
	id, err := gonanoid.Nanoid()
	if err != nil {
		log.Fatal(err)
	}
    return id
}



var texturesList = []*Texture{
	{
		ID:        "FUCCNu--2Lru2QoKhR3zc",
		Name:      "TestTexture1",
		FilePath: "/test",
		Tags: []string{"testTag1", "TestTag2"},
		CreatedOn: time.Now().UTC(),
		UpdatedOn:	time.Now().UTC(),
	},
	{
		ID:        "FGZO-fMtXeyAYRwgayFmb",
		Name:      "TestTexture2",
		FilePath: "/test",
		Tags: []string{"testTag2", "TestTag3"},
		CreatedOn: time.Now().UTC(),
		UpdatedOn:	time.Now().UTC(),
	},
}


var ErrTextureNotFound = fmt.Errorf("Texture not found")
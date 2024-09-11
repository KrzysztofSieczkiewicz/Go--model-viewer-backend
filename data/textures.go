package data

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"regexp"
	"time"

	validator "github.com/go-playground/validator/v10"
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
}

func (t *Texture) FromJSON(r io.Reader) error {
	d := json.NewDecoder(r)
	return d.Decode(t)
}

func (t *Texture) ToJSON(w io.Writer) error {
	e := json.NewEncoder(w)
	return e.Encode(t)
}


func (t *Texture) Validate() error {
	validate := validator.New()

	validate.RegisterValidation("filepath", validateFilePath)
	validate.RegisterValidation("name", validateName)

	return validate.Struct(t)
}

func validateFilePath(fl validator.FieldLevel) bool {
	re := regexp.MustCompile(`^(.*)\/([^\/]*)$`)
	matches := re.FindAllString(fl.Field().String(), -1)

	return len(matches) == 1
}

func validateName(fl validator.FieldLevel) bool {
	re := regexp.MustCompile(`^([a-zA-Z -_]*([_][0-9]*)?)$`)
	matches := re.FindAllString(fl.Field().String(), -1)

	return len(matches) == 1
}



type Textures []*Texture

func (t *Textures) ToJSON(w io.Writer) error {
	e := json.NewEncoder(w)
	return e.Encode(t)
}

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
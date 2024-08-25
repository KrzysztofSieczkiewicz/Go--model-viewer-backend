package models

import (
	"encoding/json"
	"io"
	"time"
)

type Texture struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	FilePath  string `json:"path"`
	Tags      []string `json:"tags"`
	CreatedOn time.Time `json:"-"`
	UpdatedOn time.Time `json:"-"`
}

func (t *Texture) FromJSON(r io.Reader) error {
	d := json.NewDecoder(r)
	return d.Decode(t)
}

type Textures []*Texture

func (t *Textures) ToJSON(w io.Writer) error {
	e := json.NewEncoder(w)
	return e.Encode(t)
}

func GetTextures() Textures {
	return texturesList
}

func AddTexture(t *Texture) {

}

func getNextID() int {
// TODO: how to generate random uuid using golang
	return 1
}

var texturesList = []*Texture{
	{
		ID:        "1",
		Name:      "TestTexture1",
		FilePath: "/test",
		Tags: []string{"testTag1", "TestTag2"},
		CreatedOn: time.Now().UTC(),
		UpdatedOn:	time.Now().UTC(),
	},
	{
		ID:        "2",
		Name:      "TestTexture2",
		FilePath: "/test",
		Tags: []string{"testTag2", "TestTag3"},
		CreatedOn: time.Now().UTC(),
		UpdatedOn:	time.Now().UTC(),
	},
}
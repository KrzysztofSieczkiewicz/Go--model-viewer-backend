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

type Textures []*Texture

func (t *Textures) ToJSON(w io.Writer) error {
	e := json.NewEncoder(w)
	return e.Encode(t)
}

func GetTextures() Textures {
	return texturesList
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
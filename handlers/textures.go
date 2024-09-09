// Package classification of SceneManager API
//
// Documentation for SceneManager API
//
// Schemes: http
// BasePath: /textures
// version: 0.0.1
//
// Consumes:
// - application/json
//
// Produces:
// - application/json
// swagger:meta

package handlers

import (
	"log"
	"net/http"

	"github.com/KrzysztofSieczkiewicz/ModelViewerBackend/data"
	"github.com/KrzysztofSieczkiewicz/ModelViewerBackend/middleware"
)

type Textures struct {
	logger *log.Logger
}

func NewHandler(logger*log.Logger) *Textures {
	return &Textures{logger}
}

func (t*Textures) GetTexture(rw http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	texture, err := data.GetTexture(id)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusNotFound)
		return
	}

	err = texture.ToJSON(rw)
	if err != nil {
		http.Error(rw, "Unable to encode textures data to json", http.StatusInternalServerError)
		return
	}
}

func (t*Textures) GetTextures(rw http.ResponseWriter, r *http.Request) {
	texturesList := data.GetTextures()

	err := texturesList.ToJSON(rw)
	if err != nil {
		http.Error(rw, "Unable to encode textures data to json", http.StatusInternalServerError)
		return
	}
}

func (t*Textures) PostTexture(rw http.ResponseWriter, r *http.Request) {
	texture := r.Context().Value(middleware.KeyTexture{}).(*data.Texture)
	data.AddTexture(texture)
}

func (t*Textures) PutTexture(rw http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	texture := r.Context().Value(middleware.KeyTexture{}).(*data.Texture)

	err := data.UpdateTexture(id, texture)
	if err == data.ErrTextureNotFound {
		http.Error(rw, "Texture not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(rw, "Issue occured during search for texture", http.StatusInternalServerError)
		return
	}
}

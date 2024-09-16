// Package classification of SceneManager API
//
// # Documentation for SceneManager API
//
// Schemes: http
// BasePath: /
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
	"github.com/KrzysztofSieczkiewicz/ModelViewerBackend/internal/utils"
	"github.com/KrzysztofSieczkiewicz/ModelViewerBackend/middleware"
)

// Textures is an http Handler
type Textures struct {
	logger *log.Logger
}

func NewHandler(logger*log.Logger) *Textures {
	return &Textures{logger}
}

// swagger:route GET /textures/{id} getTexture
// Returns single texture based on id
// responses:
//  200: getTexture
//  404: error
//  500: error

// GetTexture returns matched texture from the database
func (t*Textures) GetTexture(rw http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	rw.Header().Add("Content-Type", "application/json")

	texture, err := data.GetTexture(id)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusNotFound)
		return
	}

	err = utils.ToJSON(texture, rw)
	if err != nil {
		http.Error(rw, "Unable to encode textures data to json", http.StatusInternalServerError)
		return
	}
}

// swagger:route GET /textures getTextures
// Returns all available textures based on id
// responses:
//  200: getTextures
//  500: error

// GetTextures returns all textures available in the database
func (t*Textures) GetTextures(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Add("Content-Type", "application/json")
	
	texturesList := data.GetTextures()

	err := utils.ToJSON(texturesList, rw)
	if err != nil {
		http.Error(rw, "Unable to encode textures data to json", http.StatusInternalServerError)
		return
	}
}

// swagger:route POST /textures postTexture
// Adds single texture to the database
// responses:
//  201: empty

// PostTexture adds provided texture to the database
func (t*Textures) PostTexture(rw http.ResponseWriter, r *http.Request) {
	texture := r.Context().Value(middleware.KeyTexture{}).(*data.Texture)
	data.AddTexture(texture)
}

// swagger:route PUT /textures/{id} putTexture
// Updates single texture based on id
// responses:
//  201: createdResponse
//  404: errorResponse
//  500: errorResponse

// PutTexture adds provided texture to the database
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

// swagger:route DELETE /textures/{id} deleteTexture
// Deletes a texture from the database 
// responses:
//  200: noContentResponse
//  404: errorResponse
//  500: errorResponse

// DeleteTexture deletes texture from the database
func (t*Textures) DeleteTexture(rw http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	err := data.DeleteTexture(id)
	if err == data.ErrTextureNotFound {
		http.Error(rw, "Texture not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(rw, "Issue occured during texture deletion", http.StatusInternalServerError)
		return
	}
}

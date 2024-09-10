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
	"github.com/KrzysztofSieczkiewicz/ModelViewerBackend/middleware"
)

// swagger:response textureResponse
type textureResponseWrapper struct {
	// Single texture with matching id
	// in: body
	Body data.Texture
}

// swagger:response texturesResponse
type texturesResponseWrapper struct {
	// All textures in the database
	// in: body
	Body []data.Texture
}

// swagger:response noContent
type textureNoContent struct {
}

// swagger:parameters [getTexture, deleteTexture]
type textureIdParameter struct {
	// The id of the texture in the database
	// in:path
	// required: true
	ID string `json:"id"`
}

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
//  200: texturesResponse

// GetTexture returns matched texture from the database
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

// swagger:route GET /textures getTextures
// Returns all available textures based on id
// responses:
//  200: textureResponse

// GetTextures returns all textures available in the database
func (t*Textures) GetTextures(rw http.ResponseWriter, r *http.Request) {
	texturesList := data.GetTextures()

	err := texturesList.ToJSON(rw)
	if err != nil {
		http.Error(rw, "Unable to encode textures data to json", http.StatusInternalServerError)
		return
	}
}

// swagger:route POST /textures getTexture
// Adds single texture to the database
// responses:
//  201: noContent

// PostTexture adds provided texture to the database
func (t*Textures) PostTexture(rw http.ResponseWriter, r *http.Request) {
	texture := r.Context().Value(middleware.KeyTexture{}).(*data.Texture)
	data.AddTexture(texture)
}

// swagger:route PUT /textures/{id} getTexture
// Updates single texture based on id
// responses:
//  201: noContent
//  404: noContent
//  500: noContent

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
//  200: noContent
//  404: noContent
//  500: noContent

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

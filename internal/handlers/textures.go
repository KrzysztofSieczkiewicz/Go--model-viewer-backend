package handlers

import (
	"log"
	"net/http"

	"github.com/KrzysztofSieczkiewicz/ModelViewerBackend/internal/models"
)

type Textures struct {
	logger *log.Logger
}

func NewHandler(logger*log.Logger) *Textures {
	return &Textures{logger}
}

func (t*Textures) GetTexture(rw http.ResponseWriter, r *http.Request) {
	t.logger.Println("Handle GET request")

	id := r.PathValue("id")

	texture, err := models.GetTexture(id)
	if err != nil {
		http.Error(rw, "Unable to encode textures data to json", http.StatusInternalServerError)
		return
	}

	err = texture.ToJSON(rw)
	if err != nil {
		http.Error(rw, "Unable to encode textures data to json", http.StatusInternalServerError)
		return
	}
}

func (t*Textures) GetTextures(rw http.ResponseWriter, r *http.Request) {
	t.logger.Println("Handle GET request")

	texturesList := models.GetTextures()

	err := texturesList.ToJSON(rw)
	if err != nil {
		http.Error(rw, "Unable to encode textures data to json", http.StatusInternalServerError)
		return
	}
}

func (t*Textures) PostTexture(rw http.ResponseWriter, r *http.Request) {
	t.logger.Println("Handle POST request")

	texture := &models.Texture{}

	err := texture.FromJSON(r.Body)
	if err != nil {
		http.Error(rw,  err.Error(), http.StatusBadRequest)
		return
	}

	models.AddTexture(texture)
}

func (t*Textures) PutTexture(rw http.ResponseWriter, r *http.Request) {
	t.logger.Println("Handle PUT request")

	id := r.PathValue("id")

	texture := &models.Texture{}

	err := texture.FromJSON(r.Body)
	if err != nil {
		http.Error(rw,  err.Error(), http.StatusBadRequest)
		return
	}

	err = models.UpdateTexture(id, texture)
	if err == models.ErrTextureNotFound {
		http.Error(rw, "Texture not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(rw, "Issue occured during search for texture", http.StatusInternalServerError)
		return
	}
}
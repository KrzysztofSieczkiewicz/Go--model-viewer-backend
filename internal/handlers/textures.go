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

func (t*Textures) GetTexture(w http.ResponseWriter, r *http.Request) {
	t.logger.Println("Handle GET request")

	id := r.PathValue("id")

	texture, err := models.GetTexture(id)
	if err != nil {
		http.Error(w, "Unable to encode textures data to json", http.StatusInternalServerError)
	}

	err = texture.ToJSON(w)
	if err != nil {
		http.Error(w, "Unable to encode textures data to json", http.StatusInternalServerError)
	}
}

func (t*Textures) GetTextures(w http.ResponseWriter, r *http.Request) {
	t.logger.Println("Handle GET request")

	texturesList := models.GetTextures()

	err := texturesList.ToJSON(w)
	if err != nil {
		http.Error(w, "Unable to encode textures data to json", http.StatusInternalServerError)
	}
}

func (t*Textures) PostTexture(w http.ResponseWriter, r *http.Request) {
	t.logger.Println("Handle POST request")

	texture := &models.Texture{}

	err := texture.FromJSON(r.Body)
	if err != nil {
		http.Error(w,  err.Error(), http.StatusBadRequest)
	}

	models.AddTexture(texture)
}

func (t*Textures) PutTexture(w http.ResponseWriter, r *http.Request) {
	t.logger.Println("Handle PUT request")

	id := r.PathValue("id")

	texture := &models.Texture{}

	err := texture.FromJSON(r.Body)
	if err != nil {
		http.Error(w,  err.Error(), http.StatusBadRequest)
	}

	err = models.UpdateTexture(id, texture)
	if err == models.ErrTextureNotFound {
		http.Error(w, "Texture not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "Issue occured during search for texture", http.StatusInternalServerError)
		return
	}
}
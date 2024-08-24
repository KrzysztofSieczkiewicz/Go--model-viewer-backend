package handlers

import (
	"log"
	"net/http"

	"github.com/KrzysztofSieczkiewicz/ModelViewerBackend/internal/models"
)

type Textures struct {
	l *log.Logger
}

func NewTextures(l*log.Logger) *Textures {
	return &Textures{}
}

func (t*Textures) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	texturesList := models.GetTextures()
	err := texturesList.ToJSON(writer)

	if err != nil {
		http.Error(writer, "Unable to encode textures data to json", http.StatusInternalServerError)
	}
}
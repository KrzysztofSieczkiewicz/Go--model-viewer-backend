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
	switch request.Method {
	case http.MethodGet:
		t.getTextures(writer, request)
	case http.MethodPost:
		return
	}

	// Catch all
	writer.WriteHeader(http.StatusMethodNotAllowed)
}

func (t*Textures) getTextures(writer http.ResponseWriter, request *http.Request) {
	texturesList := models.GetTextures()
	err := texturesList.ToJSON(writer)

	if err != nil {
		http.Error(writer, "Unable to encode textures data to json", http.StatusInternalServerError)
	}
}
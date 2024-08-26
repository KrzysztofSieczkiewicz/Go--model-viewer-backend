package handlers

import (
	"log"
	"net/http"

	"github.com/KrzysztofSieczkiewicz/ModelViewerBackend/internal/models"
)

type Textures struct {
	logger *log.Logger
}

func NewTextures(logger*log.Logger) *Textures {
	return &Textures{logger}
}

func (t*Textures) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	switch request.Method {
	case http.MethodGet:
		t.getTextures(writer, request)
	case http.MethodPost:
		t.addTexture(writer, request)
	default:
		writer.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (t*Textures) getTextures(writer http.ResponseWriter, request *http.Request) {
	t.logger.Println("Handle GET request")

	texturesList := models.GetTextures()
	err := texturesList.ToJSON(writer)

	if err != nil {
		http.Error(writer, "Unable to encode textures data to json", http.StatusInternalServerError)
	}
}

func (t*Textures) addTexture(writer http.ResponseWriter, request *http.Request) {
	t.logger.Println("Handle POST request")

	texture := &models.Texture{}
	err := texture.FromJSON(request.Body)

	if err != nil {
		http.Error(writer,  err.Error(), http.StatusBadRequest)
	}

	models.AddTexture(texture)

	t.logger.Printf("Texture: %#v", texture)
}
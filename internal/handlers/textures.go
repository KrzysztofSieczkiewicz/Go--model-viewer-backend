package handlers

import (
	"log"
	"net/http"
	"regexp"

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
	case http.MethodPut:
		// Expect texture id in the URI
		regex := regexp.MustCompile(`[A-Za-z0-9_-]+$`)
		group := regex.FindAllStringSubmatch(request.URL.Path, -1)
		
		if len(group) != 1 {
			http.Error(writer, "Invalid URI, more than one ID matched", http.StatusBadRequest)
			return
		}
		if len(group[0]) != 1 {
			http.Error(writer, "Invalid URI, more than one capture group", http.StatusBadRequest)
			return
		}

		textureId := group[0][0]
		t.updateTexture(writer, request, textureId)

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
}

func (t*Textures) updateTexture(writer http.ResponseWriter, request *http.Request, id string) {
	t.logger.Println("Handle PUT request")

	texture := &models.Texture{}

	err := texture.FromJSON(request.Body)
	if err != nil {
		http.Error(writer,  err.Error(), http.StatusBadRequest)
	}

	err = models.UpdateTexture(id, texture)
	if err == models.ErrTextureNotFound {
		http.Error(writer, "Product not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(writer, "Issue occured during search for texture", http.StatusInternalServerError)
		return
	}
}
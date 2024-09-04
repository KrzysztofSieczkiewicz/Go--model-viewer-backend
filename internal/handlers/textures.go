package handlers

import (
	"log"
	"net/http"

	"github.com/KrzysztofSieczkiewicz/ModelViewerBackend/internal/data"
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
	texture := &data.Texture{}

	err := texture.FromJSON(r.Body)
	if err != nil {
		http.Error(rw,  err.Error(), http.StatusBadRequest)
		return
	}

	data.AddTexture(texture)
}

func (t*Textures) PutTexture(rw http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	texture := &data.Texture{}

	err := texture.FromJSON(r.Body)
	if err != nil {
		http.Error(rw,  err.Error(), http.StatusBadRequest)
		return
	}

	err = data.UpdateTexture(id, texture)
	if err == data.ErrTextureNotFound {
		http.Error(rw, "Texture not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(rw, "Issue occured during search for texture", http.StatusInternalServerError)
		return
	}
}

/*
type KeyTexture struct {}

func (t Textures) MiddlewareTexturesValidation(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		texture := &data.Texture{}

		err := texture.FromJSON(r.Body)
		if err != nil {
			http.Error(rw,  "Unable to unmarshal Texture object from JSON:\n" + err.Error(), http.StatusBadRequest)
			return
		}

		ctx := r.Context().Value(KeyTexture, texture)
		req := r.Context(ctx)

		next.ServeHTTP(rw, req)
	})
}
*/
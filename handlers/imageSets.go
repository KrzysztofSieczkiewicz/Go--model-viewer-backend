package handlers

import (
	"log"
	"net/http"

	"github.com/KrzysztofSieczkiewicz/ModelViewerBackend/data"
	"github.com/KrzysztofSieczkiewicz/ModelViewerBackend/internal/utils"
)

// ImageSets is an http Handler
type ImageSets struct {
	logger *log.Logger
}

func NewImageSetsHandler(l *log.Logger) *ImageSets {
	return &ImageSets{l}
}


func (is *ImageSets) GetImageSets(rw http.ResponseWriter, r *http.Request) {
	imgSets := data.GetImageSets()

	err := utils.ToJSON(imgSets, rw)
	if (err != nil) {
		http.Error(rw, "Unable to encode textures data to json", http.StatusInternalServerError)
		return
	}

	rw.Header().Add("Content-Type", "application/json")
}
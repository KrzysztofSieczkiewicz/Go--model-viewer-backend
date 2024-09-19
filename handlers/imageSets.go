package handlers

import (
	"fmt"
	"log"
	"net/http"
)

// ImageSets is an http Handler
type ImageSets struct {
	logger *log.Logger
}

func NewImageSetsHandler(l *log.Logger) *ImageSets {
	return &ImageSets{l}
}


func (is *ImageSets) GetImageSet(rw *http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	fmt.Print(id)
}
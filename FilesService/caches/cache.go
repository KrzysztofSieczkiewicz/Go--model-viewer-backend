package caches

import (
	"github.com/KrzysztofSieczkiewicz/go--model-viewer-backend/FilesService/models"
	"github.com/google/uuid"
)

type Cache interface {
	Set(key string, asset models.Asset)
	Get(key string) (string, error)
}

func GenerateUUID() string {
	return uuid.New().String()
}
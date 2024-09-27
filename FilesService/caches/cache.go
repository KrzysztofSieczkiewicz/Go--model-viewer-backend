package caches

import "github.com/google/uuid"

type Cache interface {
	Set(key string, value string)
	Get(key string) (string, error)
}

func GenerateUUID() string {
	return uuid.New().String()
}
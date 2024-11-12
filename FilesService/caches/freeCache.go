package caches

import (
	"github.com/KrzysztofSieczkiewicz/go--model-viewer-backend/FilesService/models"
	"github.com/coocood/freecache"
)

type freeCacheWrapper struct {
	cache *freecache.Cache
	defaultExpiraton int
}

func NewFreeCache(cacheSizeMB int, defaultExpSeconds int) *freeCacheWrapper {
	cacheSize := cacheSizeMB * 1024 * 1024
	expirationTime := defaultExpSeconds

	return &freeCacheWrapper{
		cache: freecache.NewCache(cacheSize),
		defaultExpiraton: expirationTime,
	}
}

func (fcw *freeCacheWrapper) Set(key string, asset models.Asset) {
	value := asset.ConstructFilepath()

	fcw.cache.Set(
		[]byte(key),
		[]byte(value),
		int(fcw.defaultExpiraton),
	)
}

func (fcw *freeCacheWrapper) Get(key string) (string, error) {
	data, err := fcw.cache.Get([]byte(key))
	if err != nil {
		return "", freecache.ErrNotFound
	}

	return string(data), nil
}
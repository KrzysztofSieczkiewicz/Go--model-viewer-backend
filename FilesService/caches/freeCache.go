package caches

import (
	"github.com/coocood/freecache"
)

type freeCacheWrapper struct {
	cache *freecache.Cache
	defaultExpiraton int
}

func NewFreeCache(cacheSizeMB int, defaultExpMinutes int) *freeCacheWrapper {
	cacheSize := cacheSizeMB * 1024 * 1024
	expirationTime := defaultExpMinutes * 60

	return &freeCacheWrapper{
		cache: freecache.NewCache(cacheSize),
		defaultExpiraton: expirationTime,
	}
}

func (fcw *freeCacheWrapper) Set(key string, value string) {
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
package caches

type Cache interface {
	Set(key string, value string)
	Get(key string) (string, error)
}
package caches_test

import (
	"testing"
	"time"

	"github.com/KrzysztofSieczkiewicz/go--model-viewer-backend/FilesService/caches"
	"github.com/KrzysztofSieczkiewicz/go--model-viewer-backend/FilesService/models"
	"github.com/coocood/freecache"
	"github.com/stretchr/testify/assert"
)

func TestFreeCacheWrapper_Set(t *testing.T) {
	cacheSizeMB := 1       // 1 MB cache size for testing
	expirationSeconds := 30 // Cache expiration time set to 1 minute
	cache := caches.NewFreeCache(cacheSizeMB, expirationSeconds)

	category := &models.Category{
		Path: "path/to/",
	}
	collection := &models.Collection{
		Category: category,
		ID: "collection",
	}
	key := "testKey"

	t.Run("Set and retrieve image", func(t *testing.T) {
		// Set an asset in the cache
		image := &models.Image{
			Collection: collection,
			ImgType: "albedo",
			Resolution: "2048x2048",
			FileExtension: "png",
		}
		expectedFilepath := image.ConstructFilepath()
		cache.Set(key, image)

		// Retrieve the asset from the cache
		value, err := cache.Get(key)

		assert.NoError(t, err, "Expected no error when retrieving cached value")
		assert.Equal(t, expectedFilepath, value, "Expected retrieved value to match the set filepath")
	})

	t.Run("Set and retrieve model", func(t *testing.T) {
		// Set an asset in the cache
		model := &models.Model{
			Collection: collection,
			ModelType: "scan",
			LOD: "LOD1",
			FileExtension: "gltf",
		}
		expectedFilepath := model.ConstructFilepath()
		cache.Set(key, model)

		// Retrieve the asset from the cache
		value, err := cache.Get(key)

		assert.NoError(t, err, "Expected no error when retrieving cached value")
		assert.Equal(t, expectedFilepath, value, "Expected retrieved value to match the set filepath")
	})
}

func TestFreeCacheWrapper_Get(t *testing.T) {
	cacheSizeMB := 1       // 1 MB cache size for testing
	expirationSeconds := 1 // Cache expiration time set to 1 minute
	cache := caches.NewFreeCache(cacheSizeMB, expirationSeconds)

	category := &models.Category{
		Path: "path/to/",
	}
	collection := &models.Collection{
		Category: category,
		ID: "collection",
	}
	key := "testKey"

	t.Run("Get non-existent item", func(t *testing.T) {
		// Attempt to get a non-existent item
		value, err := cache.Get("nonExistentKey")

		assert.Equal(t, "", value, "Expected empty string for a non-existent cache key")
		assert.ErrorIs(t, err, freecache.ErrNotFound, "Expected ErrNotFound for non-existent cache key")
	})

	t.Run("Get expired item", func(t *testing.T) {
		image := &models.Image{
			Collection: collection,
			ImgType: "albedo",
			Resolution: "2048x2048",
			FileExtension: "png",
		}

		// Set an asset in the cache
		cache.Set(key, image)

		// Wait for the cache item to expire
		time.Sleep(1 * time.Second)

		// Attempt to get the expired item
		value, err := cache.Get(key)

		assert.Equal(t, "", value, "Expected empty string for an expired cache item")
		assert.ErrorIs(t, err, freecache.ErrNotFound, "Expected ErrNotFound for expired cache item")
	})
}
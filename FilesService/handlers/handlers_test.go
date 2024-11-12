package handlers_test

import (
	"bytes"
	"io"
	"log/slog"
	"testing"

	"github.com/KrzysztofSieczkiewicz/go--model-viewer-backend/FilesService/files"
	"github.com/KrzysztofSieczkiewicz/go--model-viewer-backend/FilesService/models"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// Helper function to create Local instance for testing
func setupStorage(t *testing.T, maxFileSizeMB int) (*files.Local) {
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))

	tempDir := t.TempDir()

	local, err := files.NewLocal(tempDir, maxFileSizeMB, logger)
	require.NoError(t, err)
	return local
}

func setupCategory(t *testing.T, storage *files.Local, path string) *models.Category {
	category := &models.Category{
		Path: path,
	}
	err := storage.CreateCategory(category)
    require.NoError(t, err)

	return category
}

func setupCollection(t *testing.T, storage *files.Local, category *models.Category, id string) *models.Collection {
	collection := &models.Collection{
		Category: category,
		ID: id,
	}
	err := storage.CreateCollection(collection)
	require.NoError(t, err)

	return collection
}

func setupImage(t *testing.T, storage *files.Local, collection *models.Collection, imgType string, resolution string, extension string) *models.Image{
	require.NotNil(t, collection)
	require.NotNil(t, storage)
	
	mockReader := bytes.NewReader([]byte("mock file content"))

	image := &models.Image{
		Collection: collection,
		ImgType: imgType,
		Resolution: resolution,
		FileExtension: extension,
	}
	err := storage.CreateAsset(image, mockReader)
	require.NoError(t, err)

	return image
}

func setupModel(t *testing.T, storage *files.Local, collection *models.Collection, modelType string, lod string, extension string) *models.Model{
	require.NotNil(t, collection)
	require.NotNil(t, storage)
	
	mockReader := bytes.NewReader([]byte("mock file content"))

	model := &models.Model{
		Collection: collection,
		ModelType: modelType,
		LOD: lod,
		FileExtension: extension,
	}
	err := storage.CreateAsset(model, mockReader)
	require.NoError(t, err)

	return model
}

// Mock cache to simulate caching operations
type MockCache struct {
	mock.Mock
}
func (m *MockCache) Set(key string, asset models.Asset) {
	m.Called(key, asset)
}
func (m *MockCache) Get(key string) (string, error) {
	args := m.Called(key)
	return args.String(0), args.Error(1)
}
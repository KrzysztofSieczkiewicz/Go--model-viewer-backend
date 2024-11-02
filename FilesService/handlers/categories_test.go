package handlers_test

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/KrzysztofSieczkiewicz/go--model-viewer-backend/FilesService/files"
	"github.com/KrzysztofSieczkiewicz/go--model-viewer-backend/FilesService/handlers"
	"github.com/KrzysztofSieczkiewicz/go--model-viewer-backend/FilesService/models"
	"github.com/KrzysztofSieczkiewicz/go--model-viewer-backend/FilesService/response"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupHandler(storage files.Storage) (*handlers.CategoriesHandler) {
	return handlers.NewCategories(
		"http://localhost:3001",
		storage,
		slog.New(slog.NewJSONHandler(io.Discard, nil)),
		nil,
	)
}

// Helper function to create Local instance for testing
func setupStorage(t *testing.T, maxFileSizeMB int) (*files.Local) {
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))

	tempDir := t.TempDir()

	local, err := files.NewLocal(tempDir, maxFileSizeMB, logger)
	require.NoError(t, err)
	return local
}

func setupCategory(t *testing.T, storage *files.Local) *models.Category {
	category := &models.Category{
		Path: "test/category",
	}
	err := storage.CreateCategory(category)
    require.NoError(t, err)

	return category
}

func setupCollection(t *testing.T, storage *files.Local, category *models.Category, id string) *models.Collection {
	collection := &models.Collection{
		Category: *category,
		ID: id,
	}
	err := storage.CreateCollection(collection)
	require.NoError(t, err)

	return collection
}

func TestGetCategory_Success(t *testing.T) {
	storage := setupStorage(t, 1)
	handler := setupHandler(storage)
	category := setupCategory(t, storage)

	// Add two collections to the category
	_ = setupCollection(t, storage, category, "FirstCollection")
	_ = setupCollection(t, storage, category, "SecondCollection")

	reqBody, _ := json.Marshal(category)
	req := httptest.NewRequest(http.MethodGet, "/categories", bytes.NewReader(reqBody))
	rec := httptest.NewRecorder()

	handler.GetCategory(rec, req)

	res := rec.Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusOK, res.StatusCode)

	var responseData response.DirectoryContentsResponse
	err := json.NewDecoder(res.Body).Decode(&responseData)
	require.NoError(t, err)
	assert.Equal(t, 2, len(responseData.Contents))
}

func TestGetCategory_BadRequest(t *testing.T) {
	storage := setupStorage(t, 1)
	handler := setupHandler(storage)
	category := setupCategory(t, storage)

	// Add two collections to the category
	_ = setupCollection(t, storage, category, "FirstCollection")
	_ = setupCollection(t, storage, category, "SecondCollection")

	// Malform the category before marshalling
	category.Path = "/./"

	reqBody, _ := json.Marshal(category)
	req := httptest.NewRequest(http.MethodGet, "/categories", bytes.NewReader(reqBody))
	rec := httptest.NewRecorder()

	handler.GetCategory(rec, req)

	res := rec.Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)

	var responseData response.MessageResponse
	err := json.NewDecoder(res.Body).Decode(&responseData)
	assert.NoError(t, err)
	assert.Equal(t, response.MessaggeInvalidData, responseData.Message)

	// Repeat check with empty path
	category.Path = ""

	reqBody, _ = json.Marshal(category)
	req = httptest.NewRequest(http.MethodGet, "/categories", bytes.NewReader(reqBody))
	rec = httptest.NewRecorder()

	handler.GetCategory(rec, req)

	res = rec.Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)

	err = json.NewDecoder(res.Body).Decode(&responseData)
	assert.NoError(t, err)
	assert.Equal(t, response.MessaggeInvalidData, responseData.Message)

	// Repeat check with wrong model data
	reqBody, _ = json.Marshal("{entirely:\"Wrong Field\"}")
	req = httptest.NewRequest(http.MethodGet, "/categories", bytes.NewReader(reqBody))
	rec = httptest.NewRecorder()

	handler.GetCategory(rec, req)

	res = rec.Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)

	err = json.NewDecoder(res.Body).Decode(&responseData)
	assert.NoError(t, err)
	assert.Equal(t, response.MessageInvalidJsonFormat, responseData.Message)
}

func TestGetCategory_NotFound(t *testing.T) {
	storage := setupStorage(t, 1)
	handler := setupHandler(storage)
	category := setupCategory(t, storage)

	// Malform the category
	category.Path = "NonExistingCategory"

	reqBody, _ := json.Marshal(category)
	req := httptest.NewRequest(http.MethodGet, "/categories", bytes.NewReader(reqBody))
	rec := httptest.NewRecorder()

	handler.GetCategory(rec, req)

	res := rec.Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusNotFound, res.StatusCode)

	var responseData response.MessageResponse
	err := json.NewDecoder(res.Body).Decode(&responseData)
	assert.NoError(t, err)
	assert.Equal(t, response.MessageNotFound, responseData.Message)
}

func TestPostCategory_Success(t *testing.T) {
	storage := setupStorage(t, 1)
	handler := setupHandler(storage)

	category := &models.Category{
		Path: "example/category/path",
	}

	reqBody, _ := json.Marshal(category)
	req := httptest.NewRequest(http.MethodPost, "/categories", bytes.NewReader(reqBody))
	rec := httptest.NewRecorder()

	handler.PostCategory(rec, req)

	res := rec.Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusNoContent, res.StatusCode)

	body, err := io.ReadAll(res.Body)
	require.NoError(t, err)
	assert.Empty(t, string(body))
}

func TestPostCategory_BadRequest(t *testing.T) {
	storage := setupStorage(t, 1)
	handler := setupHandler(storage)

	category := &models.Category{
		Path: "/./",
	}

	reqBody, _ := json.Marshal(category)
	req := httptest.NewRequest(http.MethodGet, "/categories", bytes.NewReader(reqBody))
	rec := httptest.NewRecorder()

	handler.PostCategory(rec, req)

	res := rec.Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)

	var responseData response.MessageResponse
	err := json.NewDecoder(res.Body).Decode(&responseData)
	assert.NoError(t, err)
	assert.Equal(t, response.MessaggeInvalidData, responseData.Message)

	// Repeat check with empty path
	category.Path = ""

	reqBody, _ = json.Marshal(category)
	req = httptest.NewRequest(http.MethodGet, "/categories", bytes.NewReader(reqBody))
	rec = httptest.NewRecorder()

	handler.PostCategory(rec, req)

	res = rec.Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)

	err = json.NewDecoder(res.Body).Decode(&responseData)
	assert.NoError(t, err)
	assert.Equal(t, response.MessaggeInvalidData, responseData.Message)

	// Repeat check with wrong model data
	reqBody, _ = json.Marshal("{entirely:\"Wrong Field\"}")
	req = httptest.NewRequest(http.MethodGet, "/categories", bytes.NewReader(reqBody))
	rec = httptest.NewRecorder()

	handler.PostCategory(rec, req)

	res = rec.Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)

	err = json.NewDecoder(res.Body).Decode(&responseData)
	assert.NoError(t, err)
	assert.Equal(t, response.MessageInvalidJsonFormat, responseData.Message)
}

func TestPostCategory_Forbidden(t *testing.T) {
	storage := setupStorage(t, 1)
	handler := setupHandler(storage)
	category := setupCategory(t, storage)

	reqBody, _ := json.Marshal(category)
	req := httptest.NewRequest(http.MethodGet, "/categories", bytes.NewReader(reqBody))
	rec := httptest.NewRecorder()

	handler.PostCategory(rec, req)

	res := rec.Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusForbidden, res.StatusCode)

	var responseData response.MessageResponse
	err := json.NewDecoder(res.Body).Decode(&responseData)
	assert.NoError(t, err)
	assert.Equal(t, response.MessageAlreadyExists, responseData.Message)
}
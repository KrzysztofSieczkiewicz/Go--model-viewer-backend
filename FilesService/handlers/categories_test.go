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

func setupCategoriesHandler(storage files.Storage) (*handlers.CategoriesHandler) {
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
		Category: *category,
		ID: id,
	}
	err := storage.CreateCollection(collection)
	require.NoError(t, err)

	return collection
}

func TestGetCategory_Success(t *testing.T) {
	storage := setupStorage(t, 1)
	handler := setupCategoriesHandler(storage)
	category := setupCategory(t, storage, "category/path")

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
	handler := setupCategoriesHandler(storage)
	category := setupCategory(t, storage, "category/path")

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
	handler := setupCategoriesHandler(storage)
	category := setupCategory(t, storage, "category/path")

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
	handler := setupCategoriesHandler(storage)

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
	handler := setupCategoriesHandler(storage)

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
	handler := setupCategoriesHandler(storage)
	category := setupCategory(t, storage, "category/path")

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

func TestPutCategory_Success(t *testing.T) {
	storage := setupStorage(t, 1)
	handler := setupCategoriesHandler(storage)
	category := setupCategory(t, storage, "category/path")

	putCategory := &models.PutRequest[models.Category]{
		Existing: *category,
		New: models.Category{Path: "category/new"},
	}

	reqBody, _ := json.Marshal(putCategory)
	req := httptest.NewRequest(http.MethodPut, "/categories", bytes.NewReader(reqBody))
	rec := httptest.NewRecorder()

	handler.PutCategory(rec, req)

	res := rec.Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusOK, res.StatusCode)

	var responseData response.MessageResponse
	err := json.NewDecoder(res.Body).Decode(&responseData)
	require.NoError(t, err)
	assert.Equal(t, response.MessageUpdateSuccessful, responseData.Message)
}

func TestPutCategory_BadRequest(t *testing.T) {
	storage := setupStorage(t, 1)
	handler := setupCategoriesHandler(storage)
	category := setupCategory(t, storage, "category/path")

	putCategory := &models.PutRequest[models.Category]{
		Existing: *category,
		New: models.Category{Path: "test/new"},
	}
	reqBody, _ := json.Marshal(putCategory)
	req := httptest.NewRequest(http.MethodPut, "/categories", bytes.NewReader(reqBody))
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
	putCategory.New.Path = ""

	reqBody, _ = json.Marshal(putCategory)
	req = httptest.NewRequest(http.MethodPut, "/categories", bytes.NewReader(reqBody))
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
	req = httptest.NewRequest(http.MethodPut, "/categories", bytes.NewReader(reqBody))
	rec = httptest.NewRecorder()

	handler.PostCategory(rec, req)

	res = rec.Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)

	err = json.NewDecoder(res.Body).Decode(&responseData)
	assert.NoError(t, err)
	assert.Equal(t, response.MessageInvalidJsonFormat, responseData.Message)
}

func TestPutCategory_NotFound(t *testing.T) {
	storage := setupStorage(t, 1)
	handler := setupCategoriesHandler(storage)

	putCategory := &models.PutRequest[models.Category]{
		Existing: models.Category{Path: "test/nonExisting/path"},
		New: models.Category{Path: "test/nonExisting/path"},
	}

	reqBody, _ := json.Marshal(putCategory)
	req := httptest.NewRequest(http.MethodPut, "/categories", bytes.NewReader(reqBody))
	rec := httptest.NewRecorder()

	handler.PutCategory(rec, req)

	res := rec.Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusNotFound, res.StatusCode)

	var responseData response.MessageResponse
	err := json.NewDecoder(res.Body).Decode(&responseData)
	require.NoError(t, err)
	assert.Equal(t, response.MessageNotFound, responseData.Message)
}

func TestPutCategory_AlreadyExists(t *testing.T) {
	storage := setupStorage(t, 1)
	handler := setupCategoriesHandler(storage)
	category := setupCategory(t, storage, "category/path")
	newCategory := setupCategory(t, storage, "category/new")

	putCategory := &models.PutRequest[models.Category]{
		Existing: *category,
		New: *newCategory,
	}

	reqBody, _ := json.Marshal(putCategory)
	req := httptest.NewRequest(http.MethodPut, "/categories", bytes.NewReader(reqBody))
	rec := httptest.NewRecorder()

	handler.PutCategory(rec, req)

	res := rec.Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusForbidden, res.StatusCode)

	var responseData response.MessageResponse
	err := json.NewDecoder(res.Body).Decode(&responseData)
	require.NoError(t, err)
	assert.Equal(t, response.MessageAlreadyExists, responseData.Message)
}

func TestPutCategory_UpdateFailed(t *testing.T) {
	storage := setupStorage(t, 1)
	handler := setupCategoriesHandler(storage)
	category := setupCategory(t, storage, "category/path")

	putCategory := &models.PutRequest[models.Category]{
		Existing: *category,
		New: models.Category{Path: "test/nonExisting/path"},
	}

	reqBody, _ := json.Marshal(putCategory)
	req := httptest.NewRequest(http.MethodPut, "/categories", bytes.NewReader(reqBody))
	rec := httptest.NewRecorder()

	handler.PutCategory(rec, req)

	res := rec.Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusInternalServerError, res.StatusCode)

	var responseData response.MessageResponse
	err := json.NewDecoder(res.Body).Decode(&responseData)
	require.NoError(t, err)
	assert.Equal(t, response.MessageFailedUpdate, responseData.Message)
}

func TestDeleteCategory_Success(t *testing.T) {
	storage := setupStorage(t, 1)
	handler := setupCategoriesHandler(storage)
	category := setupCategory(t, storage, "category/path")

	reqBody, _ := json.Marshal(category)
	req := httptest.NewRequest(http.MethodDelete, "/categories", bytes.NewReader(reqBody))
	rec := httptest.NewRecorder()

	handler.DeleteCategory(rec, req)

	res := rec.Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusNoContent, res.StatusCode)
}

func TestDeleteCategory_BadRequest(t *testing.T) {
	storage := setupStorage(t, 1)
	handler := setupCategoriesHandler(storage)
	category := setupCategory(t, storage, "category/path")

	category.Path = "/category/path"

	reqBody, _ := json.Marshal(category)
	req := httptest.NewRequest(http.MethodPut, "/categories", bytes.NewReader(reqBody))
	rec := httptest.NewRecorder()

	handler.DeleteCategory(rec, req)

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
	req = httptest.NewRequest(http.MethodPut, "/categories", bytes.NewReader(reqBody))
	rec = httptest.NewRecorder()

	handler.DeleteCategory(rec, req)

	res = rec.Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)

	err = json.NewDecoder(res.Body).Decode(&responseData)
	assert.NoError(t, err)
	assert.Equal(t, response.MessaggeInvalidData, responseData.Message)

	// Repeat check with wrong model data
	reqBody, _ = json.Marshal("{entirely:\"Wrong Field\"}")
	req = httptest.NewRequest(http.MethodPut, "/categories", bytes.NewReader(reqBody))
	rec = httptest.NewRecorder()

	handler.DeleteCategory(rec, req)

	res = rec.Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)

	err = json.NewDecoder(res.Body).Decode(&responseData)
	assert.NoError(t, err)
	assert.Equal(t, response.MessageInvalidJsonFormat, responseData.Message)
}

func TestDeleteCategory_NotFound(t *testing.T) {
	storage := setupStorage(t, 1)
	handler := setupCategoriesHandler(storage)
	
	category := &models.Category{
		Path: "random/path",
	}

	reqBody, _ := json.Marshal(category)
	req := httptest.NewRequest(http.MethodDelete, "/categories", bytes.NewReader(reqBody))
	rec := httptest.NewRecorder()

	handler.DeleteCategory(rec, req)

	res := rec.Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusNotFound, res.StatusCode)

	var responseData response.MessageResponse
	err := json.NewDecoder(res.Body).Decode(&responseData)
	assert.NoError(t, err)
	assert.Equal(t, response.MessageNotFound, responseData.Message)
}

func TestDeleteCategory_NotEmpty(t *testing.T) {
	storage := setupStorage(t, 1)
	handler := setupCategoriesHandler(storage)
	category := setupCategory(t, storage, "category/path")
	_ = setupCategory(t, storage, "category/path/contents")

	reqBody, _ := json.Marshal(category)
	req := httptest.NewRequest(http.MethodDelete, "/categories", bytes.NewReader(reqBody))
	rec := httptest.NewRecorder()

	handler.DeleteCategory(rec, req)

	res := rec.Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusForbidden, res.StatusCode)

	var responseData response.MessageResponse
	err := json.NewDecoder(res.Body).Decode(&responseData)
	assert.NoError(t, err)
	assert.Equal(t, response.MessageDirectoryNotEmpty, responseData.Message)
}
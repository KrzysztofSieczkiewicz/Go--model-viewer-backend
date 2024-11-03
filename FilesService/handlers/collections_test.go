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

func setupCollectionsHandler(storage files.Storage) *handlers.CollectionsHandler {
	return handlers.NewCollections(
		"http://localhost:3001",
		storage,
		slog.New(slog.NewJSONHandler(io.Discard, nil)),
		nil,
	)
}

func TestGetCollecion_Success(t *testing.T) {
	storage := setupStorage(t, 1)
	handler := setupCollectionsHandler(storage)
	category := setupCategory(t, storage, "category/path")
	collection := setupCollection(t, storage, category, "test")

	reqBody, _ := json.Marshal(collection)
	req := httptest.NewRequest(http.MethodGet, "/collections", bytes.NewReader(reqBody))
	rec := httptest.NewRecorder()

	handler.GetCollection(rec, req)

	res := rec.Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusOK, res.StatusCode)

	var responseData response.DirectoryContentsResponse
	err := json.NewDecoder(res.Body).Decode(&responseData)
	require.NoError(t, err)
}

func TestGetCollecion_BadRequest(t *testing.T) {
	storage := setupStorage(t, 1)
	handler := setupCollectionsHandler(storage)
	category := setupCategory(t, storage, "category/path")
	collection := setupCollection(t, storage, category, "test")

	// Malform the collection id before marshalling
	collection.ID = "_newID"

	reqBody, _ := json.Marshal(collection)
	req := httptest.NewRequest(http.MethodGet, "/collections", bytes.NewReader(reqBody))
	rec := httptest.NewRecorder()

	handler.GetCollection(rec, req)

	res := rec.Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)

	var responseData response.MessageResponse
	err := json.NewDecoder(res.Body).Decode(&responseData)
	assert.NoError(t, err)
	assert.Equal(t, response.MessaggeInvalidData, responseData.Message)

	// Repeat check with empty path
	collection.ID = ""

	reqBody, _ = json.Marshal(collection)
	req = httptest.NewRequest(http.MethodGet, "/collections", bytes.NewReader(reqBody))
	rec = httptest.NewRecorder()

	handler.GetCollection(rec, req)

	res = rec.Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)

	err = json.NewDecoder(res.Body).Decode(&responseData)
	assert.NoError(t, err)
	assert.Equal(t, response.MessaggeInvalidData, responseData.Message)

	// Repeat check with wrong model data
	reqBody, _ = json.Marshal("{entirely:\"Wrong Field\"}")
	req = httptest.NewRequest(http.MethodGet, "/collections", bytes.NewReader(reqBody))
	rec = httptest.NewRecorder()

	handler.GetCollection(rec, req)

	res = rec.Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)

	err = json.NewDecoder(res.Body).Decode(&responseData)
	assert.NoError(t, err)
	assert.Equal(t, response.MessageInvalidJsonFormat, responseData.Message)
}

func TestGetCollection_NotFound(t *testing.T) {
	storage := setupStorage(t, 1)
	handler := setupCollectionsHandler(storage)
	category := setupCategory(t, storage, "category/path")
	
	collection := &models.Collection{
		Category: *category,
		ID: "nonExistingCollection",
	}

	reqBody, _ := json.Marshal(collection)
	req := httptest.NewRequest(http.MethodGet, "/collections", bytes.NewReader(reqBody))
	rec := httptest.NewRecorder()

	handler.GetCollection(rec, req)

	res := rec.Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusNotFound, res.StatusCode)

	var responseData response.MessageResponse
	err := json.NewDecoder(res.Body).Decode(&responseData)
	assert.NoError(t, err)
	assert.Equal(t, response.MessageNotFound, responseData.Message)
}
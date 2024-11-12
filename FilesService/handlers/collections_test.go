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
	)
}

func TestGetCollecion_Success(t *testing.T) {
	var responseData models.DirectoryContentsResponse

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

	err := json.NewDecoder(res.Body).Decode(&responseData)
	require.NoError(t, err)
}

func TestGetCollecion_BadRequest(t *testing.T) {
	var responseData models.MessageResponse

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

	err := json.NewDecoder(res.Body).Decode(&responseData)
	assert.NoError(t, err)
	assert.Equal(t, response.MessageInvalidData, responseData.Message)

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
	assert.Equal(t, response.MessageInvalidData, responseData.Message)

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
	var responseData models.MessageResponse

	storage := setupStorage(t, 1)
	handler := setupCollectionsHandler(storage)
	category := setupCategory(t, storage, "category/path")
	
	collection := &models.Collection{
		Category: category,
		ID: "nonExistingCollection",
	}

	reqBody, _ := json.Marshal(collection)
	req := httptest.NewRequest(http.MethodGet, "/collections", bytes.NewReader(reqBody))
	rec := httptest.NewRecorder()

	handler.GetCollection(rec, req)

	res := rec.Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusNotFound, res.StatusCode)

	err := json.NewDecoder(res.Body).Decode(&responseData)
	assert.NoError(t, err)
	assert.Equal(t, response.MessageNotFound, responseData.Message)

	// repeat with non existing category
	category.Path = "some/random/path"

	reqBody, _ = json.Marshal(collection)
	req = httptest.NewRequest(http.MethodGet, "/collections", bytes.NewReader(reqBody))
	rec = httptest.NewRecorder() 

	handler.GetCollection(rec, req)

	res = rec.Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusNotFound, res.StatusCode)
	
	err = json.NewDecoder(res.Body).Decode(&responseData)
	assert.NoError(t, err)
	assert.Equal(t, response.MessageNotFound, responseData.Message)
}

func TestPostCollection_Success(t *testing.T) {
	storage := setupStorage(t, 1)
	handler := setupCollectionsHandler(storage)
	category := setupCategory(t, storage, "category/path")
	
	collection := &models.Collection{
		Category: category,
		ID: "newCollection",
	}

	reqBody, _ := json.Marshal(collection)
	req := httptest.NewRequest(http.MethodPost, "/collections", bytes.NewReader(reqBody))
	rec := httptest.NewRecorder()

	handler.PostCollection(rec, req)

	res := rec.Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusNoContent, res.StatusCode)

	contents, err := storage.ListCategoryContents(category)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(contents))
}

func TestPostCollecion_BadRequest(t *testing.T) {
	var responseData models.MessageResponse

	storage := setupStorage(t, 1)
	handler := setupCollectionsHandler(storage)
	category := setupCategory(t, storage, "category/path")
	
	collection := &models.Collection{
		Category: category,
		ID: "_newID",
	}

	reqBody, _ := json.Marshal(collection)
	req := httptest.NewRequest(http.MethodPost, "/collections", bytes.NewReader(reqBody))
	rec := httptest.NewRecorder()

	handler.PostCollection(rec, req)

	res := rec.Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)

	err := json.NewDecoder(res.Body).Decode(&responseData)
	assert.NoError(t, err)
	assert.Equal(t, response.MessageInvalidData, responseData.Message)

	// Repeat check with empty id
	collection.ID = ""

	reqBody, _ = json.Marshal(collection)
	req = httptest.NewRequest(http.MethodPost, "/collections", bytes.NewReader(reqBody))
	rec = httptest.NewRecorder()

	handler.PostCollection(rec, req)

	res = rec.Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)

	err = json.NewDecoder(res.Body).Decode(&responseData)
	assert.NoError(t, err)
	assert.Equal(t, response.MessageInvalidData, responseData.Message)

	// Repeat check with illegal id
	collection.ID = "_collection"

	reqBody, _ = json.Marshal(collection)
	req = httptest.NewRequest(http.MethodPost, "/collections", bytes.NewReader(reqBody))
	rec = httptest.NewRecorder()

	handler.PostCollection(rec, req)

	res = rec.Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)

	err = json.NewDecoder(res.Body).Decode(&responseData)
	assert.NoError(t, err)
	assert.Equal(t, response.MessageInvalidData, responseData.Message)

	// Repeat check with wrong model data
	reqBody, _ = json.Marshal("{entirely:\"Wrong Field\"}")
	req = httptest.NewRequest(http.MethodPost, "/collections", bytes.NewReader(reqBody))
	rec = httptest.NewRecorder()

	handler.PostCollection(rec, req)

	res = rec.Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)

	err = json.NewDecoder(res.Body).Decode(&responseData)
	assert.NoError(t, err)
	assert.Equal(t, response.MessageInvalidJsonFormat, responseData.Message)
}

func TestPostCollection_NotFound(t *testing.T) {
	storage := setupStorage(t, 1)
	handler := setupCollectionsHandler(storage)

	category := &models.Category{
		Path: "category/path",
	}
	
	collection := &models.Collection{
		Category: category,
		ID: "someCollection",
	}

	reqBody, _ := json.Marshal(collection)
	req := httptest.NewRequest(http.MethodPost, "/collections", bytes.NewReader(reqBody))
	rec := httptest.NewRecorder()

	handler.PostCollection(rec, req)

	res := rec.Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusNotFound, res.StatusCode)

	var responseData models.MessageResponse
	err := json.NewDecoder(res.Body).Decode(&responseData)
	assert.NoError(t, err)
	assert.Equal(t, response.MessageNotFound, responseData.Message)
}

func TestPostCollection_AlreadyExists(t *testing.T) {
	var responseData models.MessageResponse

	storage := setupStorage(t, 1)
	handler := setupCollectionsHandler(storage)
	category := setupCategory(t, storage, "category/path")
	_ = setupCollection(t, storage, category, "collection")
	
	collection := &models.Collection{
		Category: category,
		ID: "collection",
	}

	reqBody, _ := json.Marshal(collection)
	req := httptest.NewRequest(http.MethodPut, "/collections", bytes.NewReader(reqBody))
	rec := httptest.NewRecorder()

	handler.PostCollection(rec, req)

	res := rec.Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)

	err := json.NewDecoder(res.Body).Decode(&responseData)
	assert.NoError(t, err)
	assert.Equal(t, response.MessageAlreadyExists, responseData.Message)
}

func TestPutCollection_Success(t *testing.T) {
	var responseData models.MessageResponse

	storage := setupStorage(t, 1)
	handler := setupCollectionsHandler(storage)
	category := setupCategory(t, storage, "category/path")
	initCollection := setupCollection(t, storage, category, "initialCollection")
	
	putCollection := &models.PutRequest[models.Collection]{
		Existing: *initCollection,
		New : models.Collection{
			Category: category,
			ID: "newCollection",
		},
	}

	reqBody, _ := json.Marshal(putCollection)
	req := httptest.NewRequest(http.MethodPut, "/collections", bytes.NewReader(reqBody))
	rec := httptest.NewRecorder()

	handler.PutCollection(rec, req)

	res := rec.Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusNoContent, res.StatusCode)

	err := json.NewDecoder(res.Body).Decode(&responseData)
	assert.Error(t, err)
}

func TestPutCollection_BadRequest(t *testing.T) {
	var responseData models.MessageResponse

	storage := setupStorage(t, 1)
	handler := setupCollectionsHandler(storage)
	category := setupCategory(t, storage, "category/path")

	// Initial collection invalid/illegal name
	initCollection := setupCollection(t, storage, category, "_initialCollection")
	putCollection := &models.PutRequest[models.Collection]{
		Existing: *initCollection,
		New : models.Collection{
			Category: category,
			ID: "newCollection",
		},
	}

	reqBody, _ := json.Marshal(putCollection)
	req := httptest.NewRequest(http.MethodPut, "/collections", bytes.NewReader(reqBody))
	rec := httptest.NewRecorder()

	handler.PutCollection(rec, req)

	res := rec.Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)

	err := json.NewDecoder(res.Body).Decode(&responseData)
	assert.NoError(t, err)
	assert.Equal(t, response.MessageInvalidData, responseData.Message)

	// New collection invalid/illegal name
	initCollection = setupCollection(t, storage, category, "initialCollection")
	putCollection = &models.PutRequest[models.Collection]{
		Existing: *initCollection,
		New: models.Collection{
			Category: category,
			ID: "_newCollection",
		},
	}

	reqBody, _ = json.Marshal(putCollection)
	req = httptest.NewRequest(http.MethodPut, "/collections", bytes.NewReader(reqBody))
	rec = httptest.NewRecorder()

	handler.PutCollection(rec, req)

	res = rec.Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)

	err = json.NewDecoder(res.Body).Decode(&responseData)
	assert.NoError(t, err)
	assert.Equal(t, response.MessageInvalidData, responseData.Message)

	// New collection malformed category path
	putCollection = &models.PutRequest[models.Collection]{
		New: models.Collection{
			Category: &models.Category{
				Path: "/test/",
			},
			ID: "_newCollection",
		},
	}

	reqBody, _ = json.Marshal(putCollection)
	req = httptest.NewRequest(http.MethodPut, "/collections", bytes.NewReader(reqBody))
	rec = httptest.NewRecorder()

	handler.PutCollection(rec, req)

	res = rec.Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)

	err = json.NewDecoder(res.Body).Decode(&responseData)
	assert.NoError(t, err)
	assert.Equal(t, response.MessageInvalidData, responseData.Message)

	// New collection with returning category path
	putCollection = &models.PutRequest[models.Collection]{
		New: models.Collection{
			Category: &models.Category{
				Path: "/test/",
			},
			ID: "./../test",
		},
	}

	reqBody, _ = json.Marshal(putCollection)
	req = httptest.NewRequest(http.MethodPut, "/collections", bytes.NewReader(reqBody))
	rec = httptest.NewRecorder()

	handler.PutCollection(rec, req)

	res = rec.Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)

	err = json.NewDecoder(res.Body).Decode(&responseData)
	assert.NoError(t, err)
	assert.Equal(t, response.MessageInvalidData, responseData.Message)

	reqBody, _ = json.Marshal("{entirely:\"Wrong Field\"}")
	req = httptest.NewRequest(http.MethodPut, "/collections", bytes.NewReader(reqBody))
	rec = httptest.NewRecorder()

	handler.PutCollection(rec, req)

	res = rec.Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)

	err = json.NewDecoder(res.Body).Decode(&responseData)
	assert.NoError(t, err)
	assert.Equal(t, response.MessageInvalidJsonFormat, responseData.Message)
}

func TestPutCollection_NotFound(t *testing.T) {
	var responseData models.MessageResponse

	storage := setupStorage(t, 1)
	handler := setupCollectionsHandler(storage)
	category := setupCategory(t, storage, "category/path")

	initCollection := models.Collection{
		Category: category,
		ID: "existingCollection",
	}
	newCollection := models.Collection{
		Category: category,
		ID: "newCollection",
	}
	putCollection := &models.PutRequest[models.Collection]{
		Existing: initCollection,
		New : newCollection,
	}

	reqBody, _ := json.Marshal(putCollection)
	req := httptest.NewRequest(http.MethodPut, "/collections", bytes.NewReader(reqBody))
	rec := httptest.NewRecorder()

	handler.PutCollection(rec, req)

	res := rec.Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusNotFound, res.StatusCode)

	err := json.NewDecoder(res.Body).Decode(&responseData)
	assert.NoError(t, err)
	assert.Equal(t, response.MessageNotFound, responseData.Message)
}

func TestPutCollection_AlreadyExists(t *testing.T) {
	var responseData models.MessageResponse

	storage := setupStorage(t, 1)
	handler := setupCollectionsHandler(storage)
	category := setupCategory(t, storage, "category/path")
	initCollection := setupCollection(t, storage, category, "initCollection")
	existingCollection := setupCollection(t, storage, category, "existingCollection")

	putCollection := &models.PutRequest[models.Collection]{
		Existing: *initCollection,
		New : *existingCollection,
	}

	reqBody, _ := json.Marshal(putCollection)
	req := httptest.NewRequest(http.MethodPut, "/collections", bytes.NewReader(reqBody))
	rec := httptest.NewRecorder()

	handler.PutCollection(rec, req)

	res := rec.Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)

	err := json.NewDecoder(res.Body).Decode(&responseData)
	assert.NoError(t, err)
	assert.Equal(t, response.MessageAlreadyExists, responseData.Message)
}

func TestDeleteCollection_Success(t *testing.T) {
	var responseData models.MessageResponse

	storage := setupStorage(t, 1)
	handler := setupCollectionsHandler(storage)
	category := setupCategory(t, storage, "category/path")
	collection := setupCollection(t, storage, category, "existingCollection")

	reqBody, _ := json.Marshal(collection)
	req := httptest.NewRequest(http.MethodDelete, "/collections", bytes.NewReader(reqBody))
	rec := httptest.NewRecorder()

	handler.DeleteCollection(rec, req)

	res := rec.Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusNoContent, res.StatusCode)

	err := json.NewDecoder(res.Body).Decode(&responseData)
	assert.Error(t, err)
}

func TestDeleteCollection_BadRequest(t *testing.T) {
	var responseData models.MessageResponse

	storage := setupStorage(t, 1)
	handler := setupCollectionsHandler(storage)
	category := setupCategory(t, storage, "category/path")
	collection := setupCollection(t, storage, category, "existingCollection")

	// malform ID - illegal id
	collection.ID = "/id"

	reqBody, _ := json.Marshal(collection)
	req := httptest.NewRequest(http.MethodDelete, "/collections", bytes.NewReader(reqBody))
	rec := httptest.NewRecorder()

	handler.DeleteCollection(rec, req)

	res := rec.Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)

	err := json.NewDecoder(res.Body).Decode(&responseData)
	assert.NoError(t, err)
	assert.Equal(t, response.MessageInvalidData, responseData.Message)

	// repeat - illegal id with 'returns'
	collection.ID = "id/.."

	reqBody, _ = json.Marshal(collection)
	req = httptest.NewRequest(http.MethodDelete, "/collections", bytes.NewReader(reqBody))
	rec = httptest.NewRecorder()

	handler.DeleteCollection(rec, req)

	res = rec.Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)

	err = json.NewDecoder(res.Body).Decode(&responseData)
	assert.NoError(t, err)
	assert.Equal(t, response.MessageInvalidData, responseData.Message)

	// repeat - empty id
	collection.ID = ""

	reqBody, _ = json.Marshal(collection)
	req = httptest.NewRequest(http.MethodDelete, "/collections", bytes.NewReader(reqBody))
	rec = httptest.NewRecorder()

	handler.DeleteCollection(rec, req)

	res = rec.Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)

	err = json.NewDecoder(res.Body).Decode(&responseData)
	assert.NoError(t, err)
	assert.Equal(t, response.MessageInvalidData, responseData.Message)

	// repeat - invalid model
	reqBody, _ = json.Marshal("{entirely:\"Wrong Field\"}")
	req = httptest.NewRequest(http.MethodDelete, "/collections", bytes.NewReader(reqBody))
	rec = httptest.NewRecorder()

	handler.DeleteCollection(rec, req)

	res = rec.Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)

	err = json.NewDecoder(res.Body).Decode(&responseData)
	assert.NoError(t, err)
	assert.Equal(t, response.MessageInvalidJsonFormat, responseData.Message)
}

func TestDeleteCollection_NotFound(t *testing.T) {
	var responseData models.MessageResponse

	storage := setupStorage(t, 1)
	handler := setupCollectionsHandler(storage)
	category := setupCategory(t, storage, "category/path")

	collection := &models.Collection{
		Category: category,
		ID: "collection",
	}

	reqBody, _ := json.Marshal(collection)
	req := httptest.NewRequest(http.MethodDelete, "/collections", bytes.NewReader(reqBody))
	rec := httptest.NewRecorder()

	handler.DeleteCollection(rec, req)

	res := rec.Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusNotFound, res.StatusCode)

	err := json.NewDecoder(res.Body).Decode(&responseData)
	assert.NoError(t, err)
	assert.Equal(t, response.MessageNotFound, responseData.Message)
}
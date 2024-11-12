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

// setupCategoriesHandler initializes the CategoriesHandler with mock storage
func setupCategoriesHandler(storage files.Storage) *handlers.CategoriesHandler {
	return handlers.NewCategories(
		"http://localhost:3001",
		storage,
		slog.New(slog.NewJSONHandler(io.Discard, nil)),
	)
}

// TestGetCategory_Success tests successful retrieval of category
func TestGetCategory_Success(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
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

		var responseData models.DirectoryContentsResponse
		err := json.NewDecoder(res.Body).Decode(&responseData)
		require.NoError(t, err)
		assert.Equal(t, 2, len(responseData.Contents))
	})
}

// TestGetCategory_BadRequest tests error handling when request is malformed
func TestGetCategory_BadRequest(t *testing.T) {
	storage := setupStorage(t, 1)
	handler := setupCategoriesHandler(storage)
	category := setupCategory(t, storage, "category/path")

	// Add two collections to the category
	_ = setupCollection(t, storage, category, "FirstCollection")
	_ = setupCollection(t, storage, category, "SecondCollection")

	// Run subtests for different bad requests
	t.Run("MalformCategory", func(t *testing.T) {
		category.Path = "/./"
		reqBody, _ := json.Marshal(category)
		req := httptest.NewRequest(http.MethodGet, "/categories", bytes.NewReader(reqBody))
		rec := httptest.NewRecorder()

		handler.GetCategory(rec, req)

		res := rec.Result()
		defer res.Body.Close()
		assert.Equal(t, http.StatusBadRequest, res.StatusCode)

		var responseData models.MessageResponse
		err := json.NewDecoder(res.Body).Decode(&responseData)
		assert.NoError(t, err)
		assert.Equal(t, response.MessageInvalidData, responseData.Message)
	})

	t.Run("EmptyPath", func(t *testing.T) {
		category.Path = ""
		reqBody, _ := json.Marshal(category)
		req := httptest.NewRequest(http.MethodGet, "/categories", bytes.NewReader(reqBody))
		rec := httptest.NewRecorder()

		handler.GetCategory(rec, req)

		res := rec.Result()
		defer res.Body.Close()
		assert.Equal(t, http.StatusBadRequest, res.StatusCode)

		var responseData models.MessageResponse
		err := json.NewDecoder(res.Body).Decode(&responseData)
		assert.NoError(t, err)
		assert.Equal(t, response.MessageInvalidData, responseData.Message)
	})

	t.Run("InvalidModelData", func(t *testing.T) {
		reqBody, _ := json.Marshal("{entirely:\"Wrong Field\"}")
		req := httptest.NewRequest(http.MethodGet, "/categories", bytes.NewReader(reqBody))
		rec := httptest.NewRecorder()

		handler.GetCategory(rec, req)

		res := rec.Result()
		defer res.Body.Close()
		assert.Equal(t, http.StatusBadRequest, res.StatusCode)

		var responseData models.MessageResponse
		err := json.NewDecoder(res.Body).Decode(&responseData)
		assert.NoError(t, err)
		assert.Equal(t, response.MessageInvalidJsonFormat, responseData.Message)
	})
}

// TestGetCategory_NotFound tests category not found error
func TestGetCategory_NotFound(t *testing.T) {
	t.Run("CategoryNotFound", func(t *testing.T) {
		storage := setupStorage(t, 1)
		handler := setupCategoriesHandler(storage)
		category := setupCategory(t, storage, "category/path")

		// Simulate a non-existing category
		category.Path = "NonExistingCategory"

		reqBody, _ := json.Marshal(category)
		req := httptest.NewRequest(http.MethodGet, "/categories", bytes.NewReader(reqBody))
		rec := httptest.NewRecorder()

		handler.GetCategory(rec, req)

		res := rec.Result()
		defer res.Body.Close()
		assert.Equal(t, http.StatusNotFound, res.StatusCode)

		var responseData models.MessageResponse
		err := json.NewDecoder(res.Body).Decode(&responseData)
		assert.NoError(t, err)
		assert.Equal(t, response.MessageNotFound, responseData.Message)
	})
}

// TestPostCategory_Success tests successful creation of category
func TestPostCategory_Success(t *testing.T) {
	storage := setupStorage(t, 1)
	handler := setupCategoriesHandler(storage)

	category := &models.Category{Path: "example/category/path"}

	t.Run("Successfully creates category", func(t *testing.T) {
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
	})
}

func TestPostCategory_BadRequest(t *testing.T) {
	storage := setupStorage(t, 1)
	handler := setupCategoriesHandler(storage)

	// Define the common bad requests to reduce repetition
	badRequests := []struct {
		path    	string
		expectedMsg string
	}{
		{"/./", response.MessageInvalidData}, // Invalid path
		{ "", response.MessageInvalidData},    // Empty path
		{"wrong\"path", response.MessageInvalidData}, // Invalid JSON format
	}

	for _, badRequest := range badRequests {
		t.Run("BadRequest for "+badRequest.path, func(t *testing.T) {
			category := &models.Category{Path: badRequest.path}
			reqBody, _ := json.Marshal(category)
			req := httptest.NewRequest(http.MethodPost, "/categories", bytes.NewReader(reqBody))
			rec := httptest.NewRecorder()

			handler.PostCategory(rec, req)

			res := rec.Result()
			defer res.Body.Close()
			assert.Equal(t, http.StatusBadRequest, res.StatusCode)

			var responseData models.MessageResponse
			err := json.NewDecoder(res.Body).Decode(&responseData)
			require.NoError(t, err)
			assert.Equal(t, badRequest.expectedMsg, responseData.Message)
		})
	}
}

func TestPostCategory_AlreadyExists(t *testing.T) {
	storage := setupStorage(t, 1)
	handler := setupCategoriesHandler(storage)
	category := setupCategory(t, storage, "category/path")

	t.Run("Category already exists", func(t *testing.T) {
		reqBody, _ := json.Marshal(category)
		req := httptest.NewRequest(http.MethodPost, "/categories", bytes.NewReader(reqBody))
		rec := httptest.NewRecorder()

		handler.PostCategory(rec, req)

		res := rec.Result()
		defer res.Body.Close()
		assert.Equal(t, http.StatusBadRequest, res.StatusCode)

		var responseData models.MessageResponse
		err := json.NewDecoder(res.Body).Decode(&responseData)
		require.NoError(t, err)
		assert.Equal(t, response.MessageAlreadyExists, responseData.Message)
	})
}

func TestPutCategory_Success(t *testing.T) {
	storage := setupStorage(t, 1)
	handler := setupCategoriesHandler(storage)
	category := setupCategory(t, storage, "category/path")

	putCategory := &models.PutRequest[models.Category]{
		Existing: *category,
		New: models.Category{Path: "category/new"},
	}

	t.Run("Successfully updates category", func(t *testing.T) {
		reqBody, _ := json.Marshal(putCategory)
		req := httptest.NewRequest(http.MethodPut, "/categories", bytes.NewReader(reqBody))
		rec := httptest.NewRecorder()

		handler.PutCategory(rec, req)

		res := rec.Result()
		defer res.Body.Close()
		assert.Equal(t, http.StatusOK, res.StatusCode)

		var responseData models.MessageResponse
		err := json.NewDecoder(res.Body).Decode(&responseData)
		require.NoError(t, err)
		assert.Equal(t, response.MessageUpdateSuccessful, responseData.Message)
	})
}

func TestPutCategory_BadRequest(t *testing.T) {
	storage := setupStorage(t, 1)
	handler := setupCategoriesHandler(storage)
	category := setupCategory(t, storage, "category/path")

	putCategory := &models.PutRequest[models.Category]{
		Existing: *category,
		New: models.Category{Path: "category/new"},
	}

	t.Run("BadRequest - Invalid data", func(t *testing.T) {
		reqBody, _ := json.Marshal(putCategory)
		req := httptest.NewRequest(http.MethodPut, "/categories", bytes.NewReader(reqBody))
		rec := httptest.NewRecorder()

		handler.PostCategory(rec, req)

		res := rec.Result()
		defer res.Body.Close()
		assert.Equal(t, http.StatusBadRequest, res.StatusCode)

		var responseData models.MessageResponse
		err := json.NewDecoder(res.Body).Decode(&responseData)
		require.NoError(t, err)
		assert.Equal(t, response.MessageInvalidData, responseData.Message)
	})

	t.Run("BadRequest - Empty path", func(t *testing.T) {
		putCategory.New.Path = ""
		reqBody, _ := json.Marshal(putCategory)
		req := httptest.NewRequest(http.MethodPut, "/categories", bytes.NewReader(reqBody))
		rec := httptest.NewRecorder()

		handler.PostCategory(rec, req)

		res := rec.Result()
		defer res.Body.Close()
		assert.Equal(t, http.StatusBadRequest, res.StatusCode)

		var responseData models.MessageResponse
		err := json.NewDecoder(res.Body).Decode(&responseData)
		require.NoError(t, err)
		assert.Equal(t, response.MessageInvalidData, responseData.Message)
	})

	t.Run("BadRequest - Invalid JSON", func(t *testing.T) {
		reqBody := []byte("{entirely:\"Wrong Field\"}")
		req := httptest.NewRequest(http.MethodPut, "/categories", bytes.NewReader(reqBody))
		rec := httptest.NewRecorder()

		handler.PostCategory(rec, req)

		res := rec.Result()
		defer res.Body.Close()
		assert.Equal(t, http.StatusBadRequest, res.StatusCode)

		var responseData models.MessageResponse
		err := json.NewDecoder(res.Body).Decode(&responseData)
		require.NoError(t, err)
		assert.Equal(t, response.MessageInvalidJsonFormat, responseData.Message)
	})
}

func TestPutCategory_NotFound(t *testing.T) {
	storage := setupStorage(t, 1)
	handler := setupCategoriesHandler(storage)

	putCategory := &models.PutRequest[models.Category]{
		Existing: models.Category{Path: "test/nonExisting/path"},
		New: models.Category{Path: "test/nonExisting/path"},
	}

	t.Run("Category not found", func(t *testing.T) {
		reqBody, _ := json.Marshal(putCategory)
		req := httptest.NewRequest(http.MethodPut, "/categories", bytes.NewReader(reqBody))
		rec := httptest.NewRecorder()

		handler.PutCategory(rec, req)

		res := rec.Result()
		defer res.Body.Close()
		assert.Equal(t, http.StatusNotFound, res.StatusCode)

		var responseData models.MessageResponse
		err := json.NewDecoder(res.Body).Decode(&responseData)
		require.NoError(t, err)
		assert.Equal(t, response.MessageNotFound, responseData.Message)
	})
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

	t.Run("Category already exists", func(t *testing.T) {
		reqBody, _ := json.Marshal(putCategory)
		req := httptest.NewRequest(http.MethodPut, "/categories", bytes.NewReader(reqBody))
		rec := httptest.NewRecorder()

		handler.PutCategory(rec, req)

		res := rec.Result()
		defer res.Body.Close()
		assert.Equal(t, http.StatusBadRequest, res.StatusCode)

		var responseData models.MessageResponse
		err := json.NewDecoder(res.Body).Decode(&responseData)
		require.NoError(t, err)
		assert.Equal(t, response.MessageAlreadyExists, responseData.Message)
	})
}

func TestPutCategory_UpdateFailed(t *testing.T) {
	storage := setupStorage(t, 1)
	handler := setupCategoriesHandler(storage)
	category := setupCategory(t, storage, "category/path")

	putCategory := &models.PutRequest[models.Category]{
		Existing: *category,
		New: models.Category{Path: "test/nonExisting/path"},
	}

	t.Run("Update failed due to server error", func(t *testing.T) {
		reqBody, _ := json.Marshal(putCategory)
		req := httptest.NewRequest(http.MethodPut, "/categories", bytes.NewReader(reqBody))
		rec := httptest.NewRecorder()

		handler.PutCategory(rec, req)

		res := rec.Result()
		defer res.Body.Close()
		assert.Equal(t, http.StatusInternalServerError, res.StatusCode)

		var responseData models.MessageResponse
		err := json.NewDecoder(res.Body).Decode(&responseData)
		require.NoError(t, err)
		assert.Equal(t, response.MessageFailedUpdate, responseData.Message)
	})
}

func TestDeleteCategory_Success(t *testing.T) {
	storage := setupStorage(t, 1)
	handler := setupCategoriesHandler(storage)
	category := setupCategory(t, storage, "category/to/delete")

	t.Run("Successfully deletes category by path", func(t *testing.T) {
		reqBody, _ := json.Marshal(category)
		req := httptest.NewRequest(http.MethodDelete, "/categories/", bytes.NewReader(reqBody))
		rec := httptest.NewRecorder()

		handler.DeleteCategory(rec, req)

		res := rec.Result()
		defer res.Body.Close()
		assert.Equal(t, http.StatusNoContent, res.StatusCode)

		// Verify the category was deleted
		deletedCategory, err := storage.ListCategoryContents(category)
		require.Error(t, err)
		assert.Nil(t, deletedCategory)
	})
}

func TestDeleteCategory_NotFound(t *testing.T) {
	storage := setupStorage(t, 1)
	handler := setupCategoriesHandler(storage)
	category := &models.Category{Path: "path/to/category"}

	t.Run("Returns NotFound when category does not exist", func(t *testing.T) {
		reqBody, _ := json.Marshal(category)
		req := httptest.NewRequest(http.MethodDelete, "/categories/nonexistent", bytes.NewReader(reqBody))
		rec := httptest.NewRecorder()

		handler.DeleteCategory(rec, req)

		res := rec.Result()
		defer res.Body.Close()
		assert.Equal(t, http.StatusNotFound, res.StatusCode)

		var responseData models.MessageResponse
		err := json.NewDecoder(res.Body).Decode(&responseData)
		require.NoError(t, err)
		assert.Equal(t, response.MessageNotFound, responseData.Message)
	})
}

func TestDeleteCategory_BadRequest(t *testing.T) {
	storage := setupStorage(t, 1)
	handler := setupCategoriesHandler(storage)
	category := setupCategory(t, storage, "category/path")

	badRequests := []struct {
		path string
		expectedMsg string
	}{
		{path: "/./", 		 	 expectedMsg: response.MessageInvalidData}, // Invalid path
		{path: "", 				 expectedMsg: response.MessageInvalidData}, // Empty path
		{path: "/category/path", expectedMsg: response.MessageInvalidData}, // Invalid prefix in path
		{path: "wrong\"path", 	 expectedMsg: response.MessageInvalidData}, // Invalid JSON format
	}

	for _, badRequest := range badRequests {
		t.Run("Delete Category with path: "+badRequest.path, func(t *testing.T) {
			category.Path = badRequest.path
			reqBody, _ := json.Marshal(category)
			
			req := httptest.NewRequest(http.MethodPost, "/categories", bytes.NewReader(reqBody))
			rec := httptest.NewRecorder()

			handler.PostCategory(rec, req)

			res := rec.Result()
			defer res.Body.Close()
			assert.Equal(t, http.StatusBadRequest, res.StatusCode)

			var responseData models.MessageResponse
			err := json.NewDecoder(res.Body).Decode(&responseData)
			require.NoError(t, err)
			assert.Equal(t, badRequest.expectedMsg, responseData.Message)
		})
	}
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
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)

	var responseData models.MessageResponse
	err := json.NewDecoder(res.Body).Decode(&responseData)
	assert.NoError(t, err)
	assert.Equal(t, response.MessageDirectoryNotEmpty, responseData.Message)
}
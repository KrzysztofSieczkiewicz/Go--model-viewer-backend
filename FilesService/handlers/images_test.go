package handlers_test

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/KrzysztofSieczkiewicz/go--model-viewer-backend/FilesService/files"
	"github.com/KrzysztofSieczkiewicz/go--model-viewer-backend/FilesService/handlers"
	"github.com/KrzysztofSieczkiewicz/go--model-viewer-backend/FilesService/models"
	"github.com/KrzysztofSieczkiewicz/go--model-viewer-backend/FilesService/response"
	"github.com/KrzysztofSieczkiewicz/go--model-viewer-backend/FilesService/signedurl"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func setupImagesHandler(storage files.Storage, mockCache *MockCache) *handlers.ImagesHandler {
	return handlers.NewImages(
		"http://localhost:3001",
		storage,
		slog.New(slog.NewJSONHandler(io.Discard, nil)),
		mockCache,
	)

}

func TestGetImageUrl_Success(t *testing.T) {
	storage := setupStorage(t, 1)
	mockCache := new(MockCache)
	handler := setupImagesHandler(storage, mockCache)

	category := setupCategory(t, storage, "category/path")
	collection := setupCollection(t, storage, category, "collection")
	image := setupImage(t, storage, collection, "albedo", "2048x2048", "png")

	mockCache.On("Set", mock.Anything, mock.AnythingOfType("*models.Image")).Return()
	mockCache.On("Get", mock.Anything).Return("mocked_file_path", nil)
	
	reqBody, err := json.Marshal(image)
	require.NoError(t, err)
	req := httptest.NewRequest(http.MethodGet, "/images", bytes.NewReader(reqBody))
	rec := httptest.NewRecorder()

	handler.GetImageUrl(rec, req)

	res := rec.Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusOK, res.StatusCode)

	var responseData response.FileUrlResponse
	err = json.NewDecoder(res.Body).Decode(&responseData)
	assert.NoError(t, err)
}

func TestGetImageUrl_BadRequest(t *testing.T) {
	var responseData response.MessageResponse

	storage := setupStorage(t, 1)
	mockCache := new(MockCache)
	handler := setupImagesHandler(storage, mockCache)

	category := setupCategory(t, storage, "category/path")
	collection := setupCollection(t, storage, category, "collection")
	image := setupImage(t, storage, collection, "albedo", "2048x2048", "png")

	// Malform image type
	image.ImgType = "Image12"

	mockCache.On("Set", mock.Anything, mock.AnythingOfType("*models.Image")).Return()
	mockCache.On("Get", mock.Anything).Return("mocked_file_path", nil)
	
	reqBody, err := json.Marshal(image)
	require.NoError(t, err)
	req := httptest.NewRequest(http.MethodGet, "/images", bytes.NewReader(reqBody))
	rec := httptest.NewRecorder()

	handler.GetImageUrl(rec, req)

	res := rec.Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)

	err = json.NewDecoder(res.Body).Decode(&responseData)
	assert.NoError(t, err)
	assert.Equal(t, response.MessageInvalidData, responseData.Message)

	// Retry with malformed image resolution
	image.Resolution = "25x25"

	mockCache.On("Set", mock.Anything, mock.AnythingOfType("*models.Image")).Return()
	mockCache.On("Get", mock.Anything).Return("mocked_file_path", nil)
	
	reqBody, err = json.Marshal(image)
	require.NoError(t, err)
	req = httptest.NewRequest(http.MethodGet, "/images", bytes.NewReader(reqBody))
	rec = httptest.NewRecorder()

	handler.GetImageUrl(rec, req)

	res = rec.Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)

	err = json.NewDecoder(res.Body).Decode(&responseData)
	assert.NoError(t, err)
	assert.Equal(t, response.MessageInvalidData, responseData.Message)

	// Retry with malformed image resolution 2
	image.Resolution = "25000x25000"

	mockCache.On("Set", mock.Anything, mock.AnythingOfType("*models.Image")).Return()
	mockCache.On("Get", mock.Anything).Return("mocked_file_path", nil)
	
	reqBody, err = json.Marshal(image)
	require.NoError(t, err)
	req = httptest.NewRequest(http.MethodGet, "/images", bytes.NewReader(reqBody))
	rec = httptest.NewRecorder()

	handler.GetImageUrl(rec, req)

	res = rec.Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)

	err = json.NewDecoder(res.Body).Decode(&responseData)
	assert.NoError(t, err)
	assert.Equal(t, response.MessageInvalidData, responseData.Message)

	// Retry with invalid image extension
	image.FileExtension = ".png"

	mockCache.On("Set", mock.Anything, mock.AnythingOfType("*models.Image")).Return()
	mockCache.On("Get", mock.Anything).Return("mocked_file_path", nil)
	
	reqBody, err = json.Marshal(image)
	require.NoError(t, err)
	req = httptest.NewRequest(http.MethodGet, "/images", bytes.NewReader(reqBody))
	rec = httptest.NewRecorder()

	handler.GetImageUrl(rec, req)

	res = rec.Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)

	err = json.NewDecoder(res.Body).Decode(&responseData)
	assert.NoError(t, err)
	assert.Equal(t, response.MessageInvalidData, responseData.Message)

	// Retry with malformed image extension 2
	image.FileExtension = "random"

	mockCache.On("Set", mock.Anything, mock.AnythingOfType("*models.Image")).Return()
	mockCache.On("Get", mock.Anything).Return("mocked_file_path", nil)
	
	reqBody, err = json.Marshal(image)
	require.NoError(t, err)
	req = httptest.NewRequest(http.MethodGet, "/images", bytes.NewReader(reqBody))
	rec = httptest.NewRecorder()

	handler.GetImageUrl(rec, req)

	res = rec.Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)

	err = json.NewDecoder(res.Body).Decode(&responseData)
	assert.NoError(t, err)
	assert.Equal(t, response.MessageInvalidData, responseData.Message)
}

func TestGetImageUrl_NotFound(t *testing.T) {
	storage := setupStorage(t, 1)
	mockCache := new(MockCache)
	handler := setupImagesHandler(storage, mockCache)

	category := setupCategory(t, storage, "category/path")
	collection := setupCollection(t, storage, category, "collection")
	
	image := &models.Image{
		Collection: collection,
		ImgType: "albedo",
		Resolution: "2048x2048",
		FileExtension: "png",
	}
	
	mockCache.On("Set", mock.Anything, mock.AnythingOfType("*models.Image")).Return()
	mockCache.On("Get", mock.Anything).Return("mocked_file_path", nil)
	
	reqBody, err := json.Marshal(image)
	require.NoError(t, err)
	req := httptest.NewRequest(http.MethodGet, "/images", bytes.NewReader(reqBody))
	rec := httptest.NewRecorder()

	handler.GetImageUrl(rec, req)

	res := rec.Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusNotFound, res.StatusCode)

	var responseData response.MessageResponse
	err = json.NewDecoder(res.Body).Decode(&responseData)
	assert.NoError(t, err)
	assert.Equal(t, response.MessageNotFound, responseData.Message)
}

func TestGetImage_Success(t *testing.T) {
	storage := setupStorage(t, 1)
	mockCache := new(MockCache)
	handler := setupImagesHandler(storage, mockCache)

	category := setupCategory(t, storage, "category/path")
	collection := setupCollection(t, storage, category, "collection")
	image := setupImage(t, storage, collection, "albedo", "2048x2048", "png")

	mockCache.On("Set", mock.Anything, mock.AnythingOfType("*models.Image")).Return()
	mockCache.On("Get", mock.Anything).Return(image.ConstructFilepath(), nil)

	mockCache.Set("tempId", image)

	signedUrl := signedurl.NewSignedUrl(
        "Secret key my boy",
        "http://localhost:3001/images",
        time.Duration(5*time.Minute),
    )
	url := signedUrl.GenerateSignedUrl("tempId")

	req := httptest.NewRequest(http.MethodGet, url, nil)
	rec := httptest.NewRecorder()

	handler.GetImage(rec, req)

	res := rec.Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusOK, res.StatusCode)

	assert.Equal(t, "application/octet-stream", res.Header.Get("Content-Type"))
}

func TestGetImage_InvalidUrl(t *testing.T) {
	var responseData response.MessageResponse

	storage := setupStorage(t, 1)
	mockCache := new(MockCache)
	handler := setupImagesHandler(storage, mockCache)

	category := setupCategory(t, storage, "category/path")
	collection := setupCollection(t, storage, category, "collection")
	_ = setupImage(t, storage, collection, "albedo", "2048x2048", "png")

	// Create signed url with invalid secret
	signedUrl := signedurl.NewSignedUrl(
        "Invalid secret key",
        "http://localhost:3001/images",
        time.Duration(5*time.Minute),
    )
	url := signedUrl.GenerateSignedUrl("tempId")

	req := httptest.NewRequest(http.MethodGet, url, nil)
	rec := httptest.NewRecorder()

	handler.GetImage(rec, req)

	res := rec.Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)

	err := json.NewDecoder(res.Body).Decode(&responseData)
	assert.NoError(t, err)
	assert.Equal(t, response.MessageInvalidSignature, responseData.Message)

	// Repeat with expired timestamp
	// Create signed url with invalid secret
	signedUrl = signedurl.NewSignedUrl(
        "Secret key my boy",
        "http://localhost:3001/images",
        time.Duration(1*time.Microsecond),
    )

	url = signedUrl.GenerateSignedUrl("tempId")

	// Make sure url expired
	time.Sleep(time.Duration(1*time.Second))

	req = httptest.NewRequest(http.MethodGet, url, nil)
	rec = httptest.NewRecorder()

	handler.GetImage(rec, req)

	res = rec.Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusForbidden, res.StatusCode)

	err = json.NewDecoder(res.Body).Decode(&responseData)
	assert.NoError(t, err)
	assert.Equal(t, response.MessageExpiredUrl, responseData.Message)
}

func TestGetImage_NoAsset(t *testing.T) {
	var responseData response.MessageResponse

	storage := setupStorage(t, 1)
	mockCache := new(MockCache)
	handler := setupImagesHandler(storage, mockCache)

	category := setupCategory(t, storage, "category/path")
	_ = setupCollection(t, storage, category, "collection")
	//image := setupImage(t, storage, collection, "albedo", "2048x2048", "png")

	mockCache.On("Set", mock.Anything, mock.AnythingOfType("*models.Image")).Return()
	mockCache.On("Get", mock.Anything).Return("filepath", nil)

	// Create signed url with invalid secret
	signedUrl := signedurl.NewSignedUrl(
        "Secret key my boy",
        "http://localhost:3001/images",
        time.Duration(5*time.Minute),
    )
	url := signedUrl.GenerateSignedUrl("tempId")

	req := httptest.NewRequest(http.MethodGet, url, nil)
	rec := httptest.NewRecorder()

	handler.GetImage(rec, req)

	res := rec.Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusInternalServerError, res.StatusCode)

	err := json.NewDecoder(res.Body).Decode(&responseData)
	assert.NoError(t, err)
	assert.Equal(t, response.MessageFailedRead, responseData.Message)
}


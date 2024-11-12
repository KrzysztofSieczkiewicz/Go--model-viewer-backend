package handlers_test

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"mime/multipart"
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

func setupModelHandler(storage files.Storage, mockCache *MockCache) *handlers.ModelsHandler {
	return handlers.NewModels(
		"http://localhost:3001",
		storage,
		slog.New(slog.NewJSONHandler(io.Discard, nil)),
		mockCache,
	)

}

func TestGetModelUrl_Success(t *testing.T) {
	storage := setupStorage(t, 1)
	mockCache := new(MockCache)
	handler := setupModelHandler(storage, mockCache)

	category := setupCategory(t, storage, "category/path")
	collection := setupCollection(t, storage, category, "collection")
	model := setupModel(t, storage, collection, "scan", "LOD1", "gltf")

	mockCache.On("Set", mock.Anything, mock.AnythingOfType("*models.Model")).Return()
	mockCache.On("Get", mock.Anything).Return("mocked_file_path", nil)
	
	reqBody, err := json.Marshal(model)
	require.NoError(t, err)
	req := httptest.NewRequest(http.MethodGet, "/models", bytes.NewReader(reqBody))
	rec := httptest.NewRecorder()

	handler.GetModelUrl(rec, req)

	res := rec.Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusOK, res.StatusCode)

	var responseData response.FileUrlResponse
	err = json.NewDecoder(res.Body).Decode(&responseData)
	assert.NoError(t, err)
}

func TestGetModelUrl_BadRequest(t *testing.T) {
	var responseData response.MessageResponse

	storage := setupStorage(t, 1)
	mockCache := new(MockCache)
	handler := setupModelHandler(storage, mockCache)

	category := setupCategory(t, storage, "category/path")
	collection := setupCollection(t, storage, category, "collection")
	model := setupModel(t, storage, collection, "scan", "LOD1", "gltf")

	// Malform image type
	model.ModelType = "unwrapped12"

	mockCache.On("Set", mock.Anything, mock.AnythingOfType("*models.Model")).Return()
	mockCache.On("Get", mock.Anything).Return("mocked_file_path", nil)
	
	reqBody, err := json.Marshal(model)
	require.NoError(t, err)
	req := httptest.NewRequest(http.MethodGet, "/images", bytes.NewReader(reqBody))
	rec := httptest.NewRecorder()

	handler.GetModelUrl(rec, req)

	res := rec.Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)

	err = json.NewDecoder(res.Body).Decode(&responseData)
	assert.NoError(t, err)
	assert.Equal(t, response.MessageInvalidData, responseData.Message)

	// Retry with malformed image resolution
	model.LOD = "NOT_LOD1"
	
	reqBody, err = json.Marshal(model)
	require.NoError(t, err)
	req = httptest.NewRequest(http.MethodGet, "/models", bytes.NewReader(reqBody))
	rec = httptest.NewRecorder()

	handler.GetModelUrl(rec, req)

	res = rec.Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)

	err = json.NewDecoder(res.Body).Decode(&responseData)
	assert.NoError(t, err)
	assert.Equal(t, response.MessageInvalidData, responseData.Message)

	// Retry with malformed image resolution 2
	model.LOD = "LOD1234"
	
	reqBody, err = json.Marshal(model)
	require.NoError(t, err)
	req = httptest.NewRequest(http.MethodGet, "/models", bytes.NewReader(reqBody))
	rec = httptest.NewRecorder()

	handler.GetModelUrl(rec, req)

	res = rec.Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)

	err = json.NewDecoder(res.Body).Decode(&responseData)
	assert.NoError(t, err)
	assert.Equal(t, response.MessageInvalidData, responseData.Message)

	// Retry with invalid image extension
	model.FileExtension = ".gltf"
	
	reqBody, err = json.Marshal(model)
	require.NoError(t, err)
	req = httptest.NewRequest(http.MethodGet, "/models", bytes.NewReader(reqBody))
	rec = httptest.NewRecorder()

	handler.GetModelUrl(rec, req)

	res = rec.Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)

	err = json.NewDecoder(res.Body).Decode(&responseData)
	assert.NoError(t, err)
	assert.Equal(t, response.MessageInvalidData, responseData.Message)

	// Retry with malformed image extension 2
	model.FileExtension = "random"
	
	reqBody, err = json.Marshal(model)
	require.NoError(t, err)
	req = httptest.NewRequest(http.MethodGet, "/models", bytes.NewReader(reqBody))
	rec = httptest.NewRecorder()

	handler.GetModelUrl(rec, req)

	res = rec.Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)

	err = json.NewDecoder(res.Body).Decode(&responseData)
	assert.NoError(t, err)
	assert.Equal(t, response.MessageInvalidData, responseData.Message)
}

func TestGetModelUrl_NotFound(t *testing.T) {
	storage := setupStorage(t, 1)
	mockCache := new(MockCache)
	handler := setupModelHandler(storage, mockCache)

	category := setupCategory(t, storage, "category/path")
	collection := setupCollection(t, storage, category, "collection")
	
	model := &models.Model{
		Collection: collection,
		ModelType: "scan",
		LOD: "LOD2",
		FileExtension: "gltf",
	}
	
	mockCache.On("Set", mock.Anything, mock.AnythingOfType("*models.Model")).Return()
	mockCache.On("Get", mock.Anything).Return("mocked_file_path", nil)
	
	reqBody, err := json.Marshal(model)
	require.NoError(t, err)
	req := httptest.NewRequest(http.MethodGet, "/models", bytes.NewReader(reqBody))
	rec := httptest.NewRecorder()

	handler.GetModelUrl(rec, req)

	res := rec.Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusNotFound, res.StatusCode)

	var responseData response.MessageResponse
	err = json.NewDecoder(res.Body).Decode(&responseData)
	assert.NoError(t, err)
	assert.Equal(t, response.MessageNotFound, responseData.Message)
}

func TestGetModel_Success(t *testing.T) {
	storage := setupStorage(t, 1)
	mockCache := new(MockCache)
	handler := setupModelHandler(storage, mockCache)

	category := setupCategory(t, storage, "category/path")
	collection := setupCollection(t, storage, category, "collection")
	model := setupModel(t, storage, collection, "scan", "LOD1", "gltf")

	mockCache.On("Set", mock.Anything, mock.AnythingOfType("*models.Model")).Return()
	mockCache.On("Get", mock.Anything).Return(model.ConstructFilepath(), nil)

	mockCache.Set("tempId", model)

	signedUrl := signedurl.NewSignedUrl(
        "Secret key my boy",
        "http://localhost:3001/images",
        time.Duration(5*time.Minute),
    )
	url := signedUrl.GenerateSignedUrl("tempId")

	req := httptest.NewRequest(http.MethodGet, url, nil)
	rec := httptest.NewRecorder()

	handler.GetModel(rec, req)

	res := rec.Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusOK, res.StatusCode)

	assert.Equal(t, "application/octet-stream", res.Header.Get("Content-Type"))
}

func TestGetModel_InvalidUrl(t *testing.T) {
	var responseData response.MessageResponse

	storage := setupStorage(t, 1)
	mockCache := new(MockCache)
	handler := setupModelHandler(storage, mockCache)

	category := setupCategory(t, storage, "category/path")
	collection := setupCollection(t, storage, category, "collection")
	_ = setupModel(t, storage, collection, "scan", "LOD1", "gltf")

	// Create signed url with invalid secret
	signedUrl := signedurl.NewSignedUrl(
        "Invalid secret key",
        "http://localhost:3001/images",
        time.Duration(5*time.Minute),
    )
	url := signedUrl.GenerateSignedUrl("tempId")

	req := httptest.NewRequest(http.MethodGet, url, nil)
	rec := httptest.NewRecorder()

	handler.GetModel(rec, req)

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

	handler.GetModel(rec, req)

	res = rec.Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusForbidden, res.StatusCode)

	err = json.NewDecoder(res.Body).Decode(&responseData)
	assert.NoError(t, err)
	assert.Equal(t, response.MessageExpiredUrl, responseData.Message)
}

func TestGetModel_NoFile(t *testing.T) {
	var responseData response.MessageResponse

	storage := setupStorage(t, 1)
	mockCache := new(MockCache)
	handler := setupModelHandler(storage, mockCache)

	category := setupCategory(t, storage, "category/path")
	_ = setupCollection(t, storage, category, "collection")

	mockCache.On("Set", mock.Anything, mock.AnythingOfType("*models.Model")).Return()
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

	handler.GetModel(rec, req)

	res := rec.Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusInternalServerError, res.StatusCode)

	err := json.NewDecoder(res.Body).Decode(&responseData)
	assert.NoError(t, err)
	assert.Equal(t, response.MessageFailedRead, responseData.Message)
}

func TestPostModel_Success(t *testing.T) {
	storage := setupStorage(t, 1)
	mockCache := new(MockCache)
	handler := setupModelHandler(storage, mockCache)

	category := setupCategory(t, storage, "category/path")

	collection := setupCollection(t, storage, category, "collection")
	model := &models.Model{
		Collection: collection,
		ModelType: "scan",
		LOD: "LOD2",
		FileExtension: "gltf",
	}
	imageJson, err := json.Marshal(model)
	require.NoError(t, err)

	// Create image metadata part
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	err = writer.WriteField("metadata", string(imageJson))
	require.NoError(t, err)

	// Create file part
	file := []byte("dummy image data")
	part, err := writer.CreateFormFile("file", model.ConstructName())
	require.NoError(t, err)

	_, err = part.Write(file)
	require.NoError(t, err)
	writer.Close()

	// Send the request
	req := httptest.NewRequest(http.MethodPost, "/models", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rec := httptest.NewRecorder()

	handler.PostModel(rec, req)

	res := rec.Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusCreated, res.StatusCode)

	var responseData response.MessageResponse
	err = json.NewDecoder(res.Body).Decode(&responseData)
	assert.NoError(t, err)
	assert.Equal(t, response.MessageUploadSuccessful, responseData.Message)
}

func TestPostModel_InvalidMetadataPart(t *testing.T) {
	var responseData response.MessageResponse

	storage := setupStorage(t, 1)
	mockCache := new(MockCache)
	handler := setupModelHandler(storage, mockCache)

	category := setupCategory(t, storage, "category/path")
	collection := setupCollection(t, storage, category, "collection")
	model := &models.Model{
		Collection: collection,
		ModelType: "scan",
		LOD: "LOD2",
		FileExtension: "gltf",
	}
	modelJson, err := json.Marshal(model)
	require.NoError(t, err)

	// Create image metadata part
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	err = writer.WriteField("data", string(modelJson))
	require.NoError(t, err)

	// Provide invalid file tag
	file := []byte("dummy image data")
	part, err := writer.CreateFormFile("file", model.ConstructName())
	require.NoError(t, err)

	_, err = part.Write(file)
	require.NoError(t, err)
	writer.Close()

	// Send the request
	req := httptest.NewRequest(http.MethodPost, "/models", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rec := httptest.NewRecorder()

	handler.PostModel(rec, req)

	res := rec.Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)

	err = json.NewDecoder(res.Body).Decode(&responseData)
	assert.NoError(t, err)
	assert.Equal(t, response.MessageInvalidMultipartJson, responseData.Message)

	// Repeat with no file tag
	// Create image metadata part
	body = new(bytes.Buffer)
	writer = multipart.NewWriter(body)

	// Provide invalid file tag
	file = []byte("dummy image data")
	part, err = writer.CreateFormFile("file", model.ConstructName())
	require.NoError(t, err)
	_, err = part.Write(file)
	require.NoError(t, err)

	writer.Close()

	// Send the request
	req = httptest.NewRequest(http.MethodPost, "/models", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rec = httptest.NewRecorder()

	handler.PostModel(rec, req)

	res = rec.Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)

	err = json.NewDecoder(res.Body).Decode(&responseData)
	assert.NoError(t, err)
	assert.Equal(t, response.MessageInvalidMultipartJson, responseData.Message)

	// Repeat with malformed data
	// Create image metadata part
	body = new(bytes.Buffer)
	writer = multipart.NewWriter(body)
	err = writer.WriteField("metadata", "{entirely:\"Wrong Field\"}")
	require.NoError(t, err)

	// Create file part
	file = []byte("dummy image data")
	part, err = writer.CreateFormFile("file", model.ConstructName())
	require.NoError(t, err)

	_, err = part.Write(file)
	require.NoError(t, err)
	writer.Close()

	// Send the request
	req = httptest.NewRequest(http.MethodPost, "/models", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rec = httptest.NewRecorder()

	handler.PostModel(rec, req)

	res = rec.Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)

	err = json.NewDecoder(res.Body).Decode(&responseData)
	assert.NoError(t, err)
	assert.Equal(t, response.MessageInvalidJsonFormat, responseData.Message)
}

func TestPostModel_InvalidFilePart(t *testing.T) {
	var responseData response.MessageResponse

	storage := setupStorage(t, 1)
	mockCache := new(MockCache)
	handler := setupModelHandler(storage, mockCache)

	category := setupCategory(t, storage, "category/path")
	collection := setupCollection(t, storage, category, "collection")
	model := &models.Model{
		Collection: collection,
		ModelType: "scan",
		LOD: "LOD2",
		FileExtension: "gltf",
	}
	imageModel, err := json.Marshal(model)
	require.NoError(t, err)

	// Create image metadata part
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	err = writer.WriteField("metadata", string(imageModel))
	require.NoError(t, err)

	// Provide invalid file tag
	file := []byte("dummy image data")
	part, err := writer.CreateFormFile("files", model.ConstructName())
	require.NoError(t, err)

	_, err = part.Write(file)
	require.NoError(t, err)
	writer.Close()

	// Send the request
	req := httptest.NewRequest(http.MethodPost, "/models", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rec := httptest.NewRecorder()

	handler.PostModel(rec, req)

	res := rec.Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)

	err = json.NewDecoder(res.Body).Decode(&responseData)
	assert.NoError(t, err)
	assert.Equal(t, response.MessageInvalidMultipartFile, responseData.Message)

	// Repeat with no file tag
	// Create image metadata part
	body = new(bytes.Buffer)
	writer = multipart.NewWriter(body)
	err = writer.WriteField("metadata", string(imageModel))
	require.NoError(t, err)

	writer.Close()

	// Send the request
	req = httptest.NewRequest(http.MethodPost, "/models", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rec = httptest.NewRecorder()

	handler.PostModel(rec, req)

	res = rec.Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)

	err = json.NewDecoder(res.Body).Decode(&responseData)
	assert.NoError(t, err)
	assert.Equal(t, response.MessageInvalidMultipartFile, responseData.Message)
}

func TestPostModel_AlreadyExists(t *testing.T) {
	storage := setupStorage(t, 1)
	mockCache := new(MockCache)
	handler := setupModelHandler(storage, mockCache)

	category := setupCategory(t, storage, "category/path")
	collection := setupCollection(t, storage, category, "collection")
	models := setupModel(t, storage, collection, "scan", "LOD1", "gltf")

	mockCache.On("Get", mock.Anything).Return(models.ConstructFilepath(), nil)

	modelsJson, err := json.Marshal(models)
	require.NoError(t, err)

	// Create image metadata part
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	err = writer.WriteField("metadata", string(modelsJson))
	require.NoError(t, err)

	// Create file part
	file := []byte("dummy image data")
	part, err := writer.CreateFormFile("file", models.ConstructName())
	require.NoError(t, err)

	_, err = part.Write(file)
	require.NoError(t, err)
	writer.Close()

	// Send the request
	req := httptest.NewRequest(http.MethodPost, "/models", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rec := httptest.NewRecorder()

	handler.PostModel(rec, req)

	res := rec.Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)

	var responseData response.MessageResponse
	err = json.NewDecoder(res.Body).Decode(&responseData)
	assert.NoError(t, err)
	assert.Equal(t, response.MessageAlreadyExists, responseData.Message)
}

func TestPostModel_NotFound(t *testing.T) {
	storage := setupStorage(t, 1)
	mockCache := new(MockCache)
	handler := setupModelHandler(storage, mockCache)

	category := setupCategory(t, storage, "category/path")

	collection := &models.Collection{
		Category: category,
		ID: "collection",
	}
	model := &models.Model{
		Collection: collection,
		ModelType: "scan",
		LOD: "LOD2",
		FileExtension: "gltf",
	}
	mockCache.On("Get", mock.Anything).Return(model.ConstructFilepath(), nil)

	modelJson, err := json.Marshal(model)
	require.NoError(t, err)

	// Create image metadata part
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	err = writer.WriteField("metadata", string(modelJson))
	require.NoError(t, err)

	// Create file part
	file := []byte("dummy model data")
	part, err := writer.CreateFormFile("file", model.ConstructName())
	require.NoError(t, err)

	_, err = part.Write(file)
	require.NoError(t, err)
	writer.Close()

	// Send the request
	req := httptest.NewRequest(http.MethodPost, "/models", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rec := httptest.NewRecorder()

	handler.PostModel(rec, req)

	res := rec.Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusNotFound, res.StatusCode)

	var responseData response.MessageResponse
	err = json.NewDecoder(res.Body).Decode(&responseData)
	assert.NoError(t, err)
	assert.Equal(t, response.MessageNotFound, responseData.Message)
}

func TestPutModel_Success(t *testing.T) {
	storage := setupStorage(t, 1)
	mockCache := new(MockCache)
	handler := setupModelHandler(storage, mockCache)

	category := setupCategory(t, storage, "category/path")
	collection := setupCollection(t, storage, category, "collection")
	model := setupModel(t, storage, collection, "scan", "LOD1", "gltf")
	modelJson, err := json.Marshal(model)
	require.NoError(t, err)
	
	// Create image metadata part
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	err = writer.WriteField("metadata", string(modelJson))
	require.NoError(t, err)

	// Create file part
	file := []byte("dummy image data")
	part, err := writer.CreateFormFile("file", model.ConstructName())
	require.NoError(t, err)

	_, err = part.Write(file)
	require.NoError(t, err)
	writer.Close()

	// Send the request
	req := httptest.NewRequest(http.MethodPost, "/models", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rec := httptest.NewRecorder()

	handler.PutModel(rec, req)

	res := rec.Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusOK, res.StatusCode)

	var responseData response.MessageResponse
	err = json.NewDecoder(res.Body).Decode(&responseData)
	assert.NoError(t, err)
	assert.Equal(t, response.MessageUpdateSuccessful, responseData.Message)
}

func TestPutModel_InvalidFilePart(t *testing.T) {
	var responseData response.MessageResponse

	storage := setupStorage(t, 1)
	mockCache := new(MockCache)
	handler := setupModelHandler(storage, mockCache)

	category := setupCategory(t, storage, "category/path")
	collection := setupCollection(t, storage, category, "collection")
	model := &models.Model{
		Collection: collection,
		ModelType: "scan",
		LOD: "LOD2",
		FileExtension: "gltf",
	}
	modelJson, err := json.Marshal(model)
	require.NoError(t, err)

	// Create image metadata part
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	err = writer.WriteField("metadata", string(modelJson))
	require.NoError(t, err)

	// Provide invalid file tag
	file := []byte("dummy model data")
	part, err := writer.CreateFormFile("files", model.ConstructName())
	require.NoError(t, err)

	_, err = part.Write(file)
	require.NoError(t, err)
	writer.Close()

	// Send the request
	req := httptest.NewRequest(http.MethodPost, "/models", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rec := httptest.NewRecorder()

	handler.PutModel(rec, req)

	res := rec.Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)

	err = json.NewDecoder(res.Body).Decode(&responseData)
	assert.NoError(t, err)
	assert.Equal(t, response.MessageInvalidMultipartFile, responseData.Message)

	// Repeat with no file tag
	// Create image metadata part
	body = new(bytes.Buffer)
	writer = multipart.NewWriter(body)
	err = writer.WriteField("metadata", string(modelJson))
	require.NoError(t, err)

	writer.Close()

	// Send the request
	req = httptest.NewRequest(http.MethodPost, "/models", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rec = httptest.NewRecorder()

	handler.PutModel(rec, req)

	res = rec.Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)

	err = json.NewDecoder(res.Body).Decode(&responseData)
	assert.NoError(t, err)
	assert.Equal(t, response.MessageInvalidMultipartFile, responseData.Message)
}

func TestPutModel_InvalidMetadataPart(t *testing.T) {
	var responseData response.MessageResponse

	storage := setupStorage(t, 1)
	mockCache := new(MockCache)
	handler := setupModelHandler(storage, mockCache)
	category := setupCategory(t, storage, "category/path")
	collection := setupCollection(t, storage, category, "collection")
	model := &models.Model{
		Collection: collection,
		ModelType: "scan",
		LOD: "LOD2",
		FileExtension: "gltf",
	}
	modelJson, err := json.Marshal(model)
	require.NoError(t, err)

	// Create image metadata part
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	err = writer.WriteField("data", string(modelJson))
	require.NoError(t, err)

	// Provide invalid file tag
	file := []byte("dummy model data")
	part, err := writer.CreateFormFile("file", model.ConstructName())
	require.NoError(t, err)

	_, err = part.Write(file)
	require.NoError(t, err)
	writer.Close()

	// Send the request
	req := httptest.NewRequest(http.MethodPost, "/models", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rec := httptest.NewRecorder()

	handler.PutModel(rec, req)

	res := rec.Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)

	err = json.NewDecoder(res.Body).Decode(&responseData)
	assert.NoError(t, err)
	assert.Equal(t, response.MessageInvalidMultipartJson, responseData.Message)

	// Repeat with no file tag
	// Create image metadata part
	body = new(bytes.Buffer)
	writer = multipart.NewWriter(body)

	// Provide invalid file tag
	file = []byte("dummy model data")
	part, err = writer.CreateFormFile("file", model.ConstructName())
	require.NoError(t, err)
	_, err = part.Write(file)
	require.NoError(t, err)

	writer.Close()

	// Send the request
	req = httptest.NewRequest(http.MethodPost, "/models", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rec = httptest.NewRecorder()

	handler.PutModel(rec, req)

	res = rec.Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)

	err = json.NewDecoder(res.Body).Decode(&responseData)
	assert.NoError(t, err)
	assert.Equal(t, response.MessageInvalidMultipartJson, responseData.Message)

	// Repeat with malformed data
	// Create image metadata part
	body = new(bytes.Buffer)
	writer = multipart.NewWriter(body)
	err = writer.WriteField("metadata", "{entirely:\"Wrong Field\"}")
	require.NoError(t, err)

	// Create file part
	file = []byte("dummy model data")
	part, err = writer.CreateFormFile("file", model.ConstructName())
	require.NoError(t, err)

	_, err = part.Write(file)
	require.NoError(t, err)
	writer.Close()

	// Send the request
	req = httptest.NewRequest(http.MethodPost, "/models", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rec = httptest.NewRecorder()

	handler.PutModel(rec, req)

	res = rec.Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)

	err = json.NewDecoder(res.Body).Decode(&responseData)
	assert.NoError(t, err)
	assert.Equal(t, response.MessageInvalidJsonFormat, responseData.Message)
}

func TestPutModel_NotFound(t *testing.T) {
	storage := setupStorage(t, 1)
	mockCache := new(MockCache)
	handler := setupModelHandler(storage, mockCache)

	category := setupCategory(t, storage, "category/path")

	collection := &models.Collection{
		Category: category,
		ID: "collection",
	}
	model := &models.Model{
		Collection: collection,
		ModelType: "scan",
		LOD: "LOD2",
		FileExtension: "gltf",
	}

	mockCache.On("Get", mock.Anything).Return(model.ConstructFilepath(), nil)

	modelJson, err := json.Marshal(model)
	require.NoError(t, err)

	// Create image metadata part
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	err = writer.WriteField("metadata", string(modelJson))
	require.NoError(t, err)

	// Create file part
	file := []byte("dummy model data")
	part, err := writer.CreateFormFile("file", model.ConstructName())
	require.NoError(t, err)

	_, err = part.Write(file)
	require.NoError(t, err)
	writer.Close()

	// Send the request
	req := httptest.NewRequest(http.MethodPost, "/models", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rec := httptest.NewRecorder()

	handler.PutModel(rec, req)

	res := rec.Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusNotFound, res.StatusCode)

	var responseData response.MessageResponse
	err = json.NewDecoder(res.Body).Decode(&responseData)
	assert.NoError(t, err)
	assert.Equal(t, response.MessageNotFound, responseData.Message)
}

func TestPutModelData_Success(t *testing.T) {
	storage := setupStorage(t, 1)
	mockCache := new(MockCache)
	handler := setupModelHandler(storage, mockCache)

	category := setupCategory(t, storage, "category/path")
	collection := setupCollection(t, storage, category, "collection")
	model := setupModel(t, storage, collection, "scan", "LOD1", "gltf")
	
	putModel := &models.PutRequest[models.Model]{
		Existing: *model,
		New: models.Model{
			Collection: collection,
			ModelType: "scan",
			LOD: "LOD2",
			FileExtension: "gltf",
		},
	}

	reqBody, _ := json.Marshal(putModel)
	req := httptest.NewRequest(http.MethodGet, "/models", bytes.NewReader(reqBody))
	rec := httptest.NewRecorder()

	handler.PutModelData(rec, req)

	res := rec.Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusOK, res.StatusCode)

	var responseData response.MessageResponse
	err := json.NewDecoder(res.Body).Decode(&responseData)
	assert.NoError(t, err)
	assert.Equal(t, response.MessageUpdateSuccessful, responseData.Message)
}

func TestPutModelData_BadRequest(t *testing.T) {
	var responseData response.MessageResponse

	storage := setupStorage(t, 1)
	mockCache := new(MockCache)
	handler := setupModelHandler(storage, mockCache)

	reqBody, _ := json.Marshal("{entirely:\"Wrong Field\"}")
	req := httptest.NewRequest(http.MethodGet, "/models", bytes.NewReader(reqBody))
	rec := httptest.NewRecorder()

	handler.PutModelData(rec, req)

	res := rec.Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)

	err := json.NewDecoder(res.Body).Decode(&responseData)
	assert.NoError(t, err)
	assert.Equal(t, response.MessageInvalidJsonFormat, responseData.Message)

	// repeat with invalid data
	category := setupCategory(t, storage, "category/path")
	collection := setupCollection(t, storage, category, "collection")
	model := setupModel(t, storage, collection, "scan", "LOD1", "gltf")
	invalidModel := setupModel(t, storage, collection, "_", " ", "")

	putModel := &models.PutRequest[models.Model]{
		Existing: *model,
		New: *invalidModel,
	}

	reqBody, _ = json.Marshal(putModel)
	req = httptest.NewRequest(http.MethodGet, "/models", bytes.NewReader(reqBody))
	rec = httptest.NewRecorder()

	handler.PutModelData(rec, req)

	res = rec.Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)

	err = json.NewDecoder(res.Body).Decode(&responseData)
	assert.NoError(t, err)
	assert.Equal(t, response.MessageInvalidData, responseData.Message)

	// repeat with wrong model
	image := &models.Image{
		Collection: collection,
		ImgType: "albedo",
		Resolution: "2048x2048",
		FileExtension: "png",
	}

	reqBody, _ = json.Marshal(image)
	req = httptest.NewRequest(http.MethodGet, "/models", bytes.NewReader(reqBody))
	rec = httptest.NewRecorder()

	handler.PutModelData(rec, req)

	res = rec.Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)

	err = json.NewDecoder(res.Body).Decode(&responseData)
	assert.NoError(t, err)
	assert.Equal(t, response.MessageInvalidData, responseData.Message)
}

func TestPutModelData_NotFound(t *testing.T) {
	storage := setupStorage(t, 1)
	mockCache := new(MockCache)
	handler := setupModelHandler(storage, mockCache)

	category := setupCategory(t, storage, "category/path")
	collection := setupCollection(t, storage, category, "collection")

	model := models.Model{
		Collection: collection,
		ModelType: "scan",
		LOD: "LOD1",
		FileExtension: "gltf",
	}
	putModel := &models.PutRequest[models.Model]{
		Existing: model,
		New: models.Model{
			Collection: collection,
			ModelType: "model",
			LOD: "LOD2",
			FileExtension: "stl",
		},
	}

	reqBody, _ := json.Marshal(putModel)
	req := httptest.NewRequest(http.MethodGet, "/models", bytes.NewReader(reqBody))
	rec := httptest.NewRecorder()

	handler.PutModelData(rec, req)

	res := rec.Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusNotFound, res.StatusCode)

	var responseData response.MessageResponse
	err := json.NewDecoder(res.Body).Decode(&responseData)
	assert.NoError(t, err)
	assert.Equal(t, response.MessageNotFound, responseData.Message)
}

func TestDeleteModel_Success(t *testing.T) {
	storage := setupStorage(t, 1)
	mockCache := new(MockCache)
	handler := setupModelHandler(storage, mockCache)

	category := setupCategory(t, storage, "category/path")
	collection := setupCollection(t, storage, category, "collection")
	model := setupModel(t, storage, collection, "scan", "LOD1", "gltf")

	reqBody, _ := json.Marshal(model)
	req := httptest.NewRequest(http.MethodGet, "/models", bytes.NewReader(reqBody))
	rec := httptest.NewRecorder()

	handler.DeleteModel(rec, req)

	res := rec.Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusNoContent, res.StatusCode)

	var responseData response.MessageResponse
	err := json.NewDecoder(res.Body).Decode(&responseData)
	assert.Error(t, err)
	assert.Empty(t, responseData.Message)
}

func TestDeleteModel_BadRequest(t *testing.T) {
	var responseData response.MessageResponse

	storage := setupStorage(t, 1)
	mockCache := new(MockCache)
	handler := setupModelHandler(storage, mockCache)

	category := setupCategory(t, storage, "category/path")
	collection := setupCollection(t, storage, category, "collection")
	model := models.Model{
		Collection: collection,
		ModelType: "_scan",
		LOD: "LOD1",
		FileExtension: "gltf",
	}

	reqBody, _ := json.Marshal(model)
	req := httptest.NewRequest(http.MethodGet, "/models", bytes.NewReader(reqBody))
	rec := httptest.NewRecorder()

	handler.DeleteModel(rec, req)

	res := rec.Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
	
	err := json.NewDecoder(res.Body).Decode(&responseData)
	assert.NoError(t, err)
	assert.Equal(t, response.MessageInvalidData, responseData.Message)

	// repeat with malformed payload
	reqBody, _ = json.Marshal("{entirely:\"Wrong Field\"}")
	req = httptest.NewRequest(http.MethodGet, "/models", bytes.NewReader(reqBody))
	rec = httptest.NewRecorder()

	handler.DeleteModel(rec, req)

	res = rec.Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)

	err = json.NewDecoder(res.Body).Decode(&responseData)
	assert.NoError(t, err)
	assert.Equal(t, response.MessageInvalidJsonFormat, responseData.Message)
}

func TestDeleteModel_NotFound(t *testing.T) {
	storage := setupStorage(t, 1)
	mockCache := new(MockCache)
	handler := setupModelHandler(storage, mockCache)

	category := setupCategory(t, storage, "category/path")
	collection := setupCollection(t, storage, category, "collection")
	model := models.Model{
		Collection: collection,
		ModelType: "scan",
		LOD: "LOD1",
		FileExtension: "glTF",
	}

	reqBody, _ := json.Marshal(model)
	req := httptest.NewRequest(http.MethodGet, "/models", bytes.NewReader(reqBody))
	rec := httptest.NewRecorder()

	handler.DeleteModel(rec, req)

	res := rec.Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusNotFound, res.StatusCode)

	var responseData response.MessageResponse
	err := json.NewDecoder(res.Body).Decode(&responseData)
	assert.NoError(t, err)
	assert.Equal(t, response.MessageNotFound, responseData.Message)
}
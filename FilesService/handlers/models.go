package handlers

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/KrzysztofSieczkiewicz/go--model-viewer-backend/FilesService/caches"
	"github.com/KrzysztofSieczkiewicz/go--model-viewer-backend/FilesService/files"
	"github.com/KrzysztofSieczkiewicz/go--model-viewer-backend/FilesService/models"
	"github.com/KrzysztofSieczkiewicz/go--model-viewer-backend/FilesService/response"
	"github.com/KrzysztofSieczkiewicz/go--model-viewer-backend/FilesService/signedurl"
	"github.com/KrzysztofSieczkiewicz/go--model-viewer-backend/FilesService/utils"
)

// curl -v -i -X POST http://localhost:9090/models -H "Content-Type: multipart/form-data" -F "metadata={\"collection\":{\"category\":{\"path\":\"random/test\"},\"id\":\"1\"},\"type\":\"Asset\",\"lod\":\"LOD0\",\"extension\":\"png\"}" -F "file=@FilesService/thumbnail.png;type=image/png"

// curl -v -i -X POST http://localhost:9090/models -H "Content-Type: multipart/form-data" -F "metadata={\"path\":\"random/test\",\"id\":\"1\",\"type\":\"Asset\",\"lod\":\"LOD0\",\"extension\":\"png\"}" -F "file=@FilesService/thumbnail.png;type=image/png"

// Handler for managing models
type ModelsHandler struct {
	baseUrl		string
	logger		*slog.Logger
	store		files.Storage
	cache		caches.Cache
	signedUrl	signedurl.SignedUrl
}

func NewModels(baseUrl string, storage files.Storage, slogger *slog.Logger, cache caches.Cache) *ModelsHandler {
	logger := slogger.With(slog.String("handler", "files")) // TODO: do this when initializing logger in the main (you can pass the same logger to the store then)

	return &ModelsHandler{
		baseUrl: baseUrl,
		store:   storage,
		logger:  logger,
		cache:   cache,
		signedUrl: *signedurl.NewSignedUrl(
			"Secret key my boy",
			baseUrl+"/files", // TODO: accept as parameter from main.go
			time.Duration(5*int(time.Minute)),
		),
	}
}

// swagger:route GET /models models getModelUrl
//
// Return a signed url pointing to the requested model
//
// consumes:
//	- application/json
//
// produces:
//	- application/json
//
// Responses:
// 	200: fileUrl
//  400: message
//	404: message
//	500: message
func (h *ModelsHandler) GetModelUrl(rw http.ResponseWriter, r *http.Request) {
	h.logger.Info("Processing GET Model URL request")

	model := &models.Model{}
	err := utils.FromJSON(model, r.Body)
	if err != nil {
		response.RespondWithMessage(rw, http.StatusBadRequest, response.MessageInvalidJsonFormat)
		return
	}

	err = model.Validate()
	if err != nil {
		response.RespondWithMessage(rw, http.StatusBadRequest, response.MessaggeInvalidData)
		return
	}

	err = h.store.CheckAsset(model)
	if err != nil {
		if err == files.ErrNotFound {
			http.Error(rw, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(rw, err.Error(), http.StatusInternalServerError)
	}

	tmpId := caches.GenerateUUID()
	h.cache.Set(tmpId, model)
	url := h.signedUrl.GenerateSignedUrl(tmpId)

    urlResponse := response.FileUrlResponse{URL: url}
	response.RespondWithJSON(rw, http.StatusOK, urlResponse)
}

// swagger:route GET /{id}&{expires}&{signature} models getModel
//
// Returns a model from collection. Can only be accessed by signed URLs from getModelUrl request
//
// produces:
//  - application/octet-stream
//	- application/json
//
// Responses:
// 	200: fileByteStream
//	400: message
//	403: message
//	404: message
//	500: message
func (h *ModelsHandler) GetModel(rw http.ResponseWriter, r *http.Request) {
	h.logger.Info("Processing GET Model request")

	id := r.URL.Query().Get("id")
	exp := r.URL.Query().Get("expires")
	sign := r.URL.Query().Get("signature")

	err := h.signedUrl.ValidateSignedUrl(id, exp, sign)
	if err != nil {
		if err == signedurl.ErrUrlExpired {
			response.RespondWithMessage(rw, http.StatusForbidden, "URL has expired")
			return
		}
		if err == signedurl.ErrInvalidSignature {
			response.RespondWithMessage(rw, http.StatusBadRequest, "Invalid signature")
			return
		}
		if err == signedurl.ErrInvalidTimestamp {
			response.RespondWithMessage(rw, http.StatusBadRequest, "Invalid timestamp")
			return
		}
		response.RespondWithMessage(rw, http.StatusBadRequest, "Invalid request")
		return
	}

	fp, err := h.cache.Get(id)
	if err != nil {
		response.RespondWithMessage(rw, http.StatusInternalServerError, "Request doesn't match cache")
		return
	}

	err = h.store.GetAsset(fp, rw)
	if err != nil {
		response.RespondWithMessage(rw, http.StatusInternalServerError, "Failed to retrieve requested file")
		return
	}

	rw.Header().Set("Content-Type", "application/octet-stream")
	rw.WriteHeader(http.StatusOK)
}

// swagger:route POST /models models postModel
//
// Add an model file to the existing collection
//
// consumes:
//  - multipart/form-data
//
// produces:
//	- application/json
//
// Responses:
// 	201: message
//  400: message
// 	403: message
// 	500: message
func (h *ModelsHandler) PostModel(rw http.ResponseWriter, r *http.Request) {
	h.logger.Info("Processing POST Model request")

	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		response.RespondWithMessage(rw, http.StatusBadRequest, "Unable to parse form data")
		return
	}

	model := &models.Model{}
	json := r.FormValue("metadata")
	if json == "" {
		response.RespondWithMessage(rw, http.StatusBadRequest, "Invalid JSON part of the request")
		return
	}

	err = utils.FromJSONString(model, json)
	if err != nil {
		response.RespondWithMessage(rw, http.StatusBadRequest, response.MessageInvalidJsonFormat)
		return
	}

	err = model.Validate()
	if err != nil {
		response.RespondWithMessage(rw, http.StatusBadRequest, response.MessaggeInvalidData)
		return
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		response.RespondWithMessage(rw, http.StatusBadRequest, "Error reading file from request")
		return
	}
	defer file.Close()

	err = h.store.AddAsset(model, file)
	if err != nil {
		if err == files.ErrAlreadyExists {
			response.RespondWithMessage(rw, http.StatusForbidden, "Asset already exists")
			return
		}
		if err == files.ErrNotFound {
			response.RespondWithMessage(rw, http.StatusBadRequest, "Collection doesn't exist")
			return
		}
		response.RespondWithMessage(rw, http.StatusInternalServerError, "Failed to create the file")
		return
	}

	response.RespondWithMessage(rw, http.StatusCreated, "Asset uploaded sucessfully")
}

// swagger:route PUT /models models putModel
//
// Overwrites the model file in the collection
//
// consumes:
//  - multipart/form-data
//
// produces:
//	- application/json
//
// Responses:
// 	200: message
//  400: message
// 	404: message
// 	500: message
func (h *ModelsHandler) PutModel(rw http.ResponseWriter, r *http.Request) {
	h.logger.Info("Processing PUT Model request")

	model := &models.Model{}
	json := r.FormValue("metadata")
	if json == "" {
		response.RespondWithMessage(rw, http.StatusBadRequest, "Invalid JSON part of the request")
		return
	}

	err := utils.FromJSONString(model, json)
	if err != nil {
		response.RespondWithMessage(rw, http.StatusBadRequest, response.MessageInvalidJsonFormat)
		return
	}
	
	err = model.Validate()
	if err != nil {
		response.RespondWithMessage(rw, http.StatusBadRequest, response.MessaggeInvalidData)
		return
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		response.RespondWithMessage(rw, http.StatusBadRequest, "Error reading file from request")
		return
	}
	defer file.Close()

	err = h.store.OverwriteAsset(model, file)
	if err != nil {
		if err == files.ErrNotFound {
			response.RespondWithMessage(rw, http.StatusNotFound, "File does not exist")
			return
		}
		response.RespondWithMessage(rw, http.StatusInternalServerError, "Failed to update the file")
		return
	}

	response.RespondWithMessage(rw, http.StatusOK, "Asset updated sucessfully")
}

// swagger:route PUT /models/update models putModelData
//
// Updates the model data without changing the file contents
//
// consumes:
//  - application/json
//
// produces:
//	- application/json
//
// Responses:
// 	200: message
//  400: message
// 	404: message
// 	500: message
func (h *ModelsHandler) PutModelData(rw http.ResponseWriter, r *http.Request) {
	h.logger.Info("Processing PUT Model request")

	request := &models.PutRequest[models.Model]{}
	err := utils.FromJSON(request, r.Body)
	if err != nil {
		response.RespondWithMessage(rw, http.StatusBadRequest, response.MessageInvalidJsonFormat)
		return
	}

	err = request.Existing.Validate()
	if err != nil {
		response.RespondWithMessage(rw, http.StatusBadRequest, response.MessaggeInvalidData)
		return
	}

	err = request.New.Validate()
	if err != nil {
		response.RespondWithMessage(rw, http.StatusBadRequest, response.MessaggeInvalidData)
		return
	}

	err = h.store.UpdateAsset(&request.Existing, &request.New)
	if err != nil {
		if err == files.ErrNotFound {
			response.RespondWithMessage(rw, http.StatusNotFound, "File does not exist")
			return
		}
		response.RespondWithMessage(rw, http.StatusInternalServerError, "Failed to update the file")
		return
	}

	response.RespondWithMessage(rw, http.StatusOK, "Asset updated sucessfully")
}

// swagger:route DELETE /models models deleteModel
//
// Remove model from the collection
//
// consumes:
//  - application/json
//
// produces:
//	- application/json
//
// Responses:
// 	204: empty
//  400: message
//	404: message
//	500: message
func (h *ModelsHandler) DeleteModel(rw http.ResponseWriter, r *http.Request) {
	h.logger.Info("Processing DELETE Model request")

	model := &models.Model{}
	err := utils.FromJSON(model, r.Body)
	if err != nil {
		response.RespondWithMessage(rw, http.StatusBadRequest, response.MessageInvalidJsonFormat)
		return
	}

	err = model.Validate()
	if err != nil {
		response.RespondWithMessage(rw, http.StatusBadRequest, response.MessaggeInvalidData)
		return
	}

	err = h.store.DeleteAsset(model)
	if err != nil {
		if err == files.ErrNotFound {
			response.RespondWithMessage(rw, http.StatusNotFound, "Model was not found")
			return
		}
		response.RespondWithMessage(rw, http.StatusBadRequest, "Failed to delete the model")
		return
	}

	rw.WriteHeader(http.StatusNoContent)
}
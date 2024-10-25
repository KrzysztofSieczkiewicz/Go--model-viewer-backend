package handlers

import (
	"log/slog"
	"net/http"
	"path/filepath"
	"time"

	"github.com/KrzysztofSieczkiewicz/go--model-viewer-backend/FilesService/caches"
	"github.com/KrzysztofSieczkiewicz/go--model-viewer-backend/FilesService/files"
	"github.com/KrzysztofSieczkiewicz/go--model-viewer-backend/FilesService/models"
	"github.com/KrzysztofSieczkiewicz/go--model-viewer-backend/FilesService/response"
	"github.com/KrzysztofSieczkiewicz/go--model-viewer-backend/FilesService/signedurl"
	"github.com/KrzysztofSieczkiewicz/go--model-viewer-backend/FilesService/utils"
)

// Handler for managing assets
type AssetsHandler struct {
	baseUrl		string
	logger		*slog.Logger
	store		files.Storage
	cache		caches.Cache
	signedUrl	signedurl.SignedUrl
}

func NewAssets(baseUrl string, storage files.Storage, slogger *slog.Logger, cache caches.Cache) *AssetsHandler {
	logger := slogger.With(slog.String("handler", "files")) // TODO: do this when initializing logger in the main (you can pass the same logger to the store then)

	return &AssetsHandler{
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

// swagger:route GET /assets assets getAssetUrl
//
// Return a signed url pointing to the requested asset
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
func (h *AssetsHandler) GetAssetUrl(rw http.ResponseWriter, r *http.Request) {
	h.logger.Info("Processing GET Asset URL request")

	a := &models.Asset{}
	err := utils.FromJSON(a, r.Body)
	if err != nil {
		response.RespondWithMessage(rw, http.StatusBadRequest, response.MessageInvalidJsonFormat)
		return
	}

	err = a.Validate()
	if err != nil {
		response.RespondWithMessage(rw, http.StatusBadRequest, response.MessaggeInvalidData)
		return
	}

	fn := a.ConstructName()
	fp := filepath.Join(
		a.Category,
		a.ID,
		fn,
	)

	err = h.store.IfExists(fp)
	if err != nil {
		if err == files.ErrNotFound {
			http.Error(rw, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(rw, err.Error(), http.StatusInternalServerError)
	}

	tmpId := caches.GenerateUUID()
	h.cache.Set(tmpId, fp)
	url := h.signedUrl.GenerateSignedUrl(tmpId)

    urlResponse := response.FileUrlResponse{
        Filename: fn,
        URL:      url,
    }

	response.RespondWithJSON(rw, http.StatusOK, urlResponse)
}

// swagger:route GET /{id}&{expires}&{signature} assets getAsset
//
// Return an asset from collection. Can only be accessed by signed URLs from getAssetUrl request
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
func (h *AssetsHandler) GetAsset(rw http.ResponseWriter, r *http.Request) {
	h.logger.Info("Processing GET Asset request")

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

	err = h.store.ReadFile(fp, rw)
	if err != nil {
		response.RespondWithMessage(rw, http.StatusInternalServerError, "Failed to retrieve requested file")
		return
	}

	rw.Header().Set("Content-Type", "application/octet-stream")
	rw.WriteHeader(http.StatusOK)
}

// swagger:route POST /assets assets postAsset
//
// Add an asset file to the existing collection
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
func (h *AssetsHandler) PostAsset(rw http.ResponseWriter, r *http.Request) {
	h.logger.Info("Processing POST Asset request")

	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		response.RespondWithMessage(rw, http.StatusBadRequest, "Unable to parse form data")
		return
	}

	a := &models.Asset{}
	json := r.FormValue("metadata")
	if json == "" {
		response.RespondWithMessage(rw, http.StatusBadRequest, "Invalid JSON part of the request")
		return
	}

	err = utils.FromJSONString(a, json)
	if err != nil {
		response.RespondWithMessage(rw, http.StatusBadRequest, response.MessageInvalidJsonFormat)
		return
	}

	err = a.Validate()
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
	
	fp := filepath.Join(
		a.Category,
		a.ID,
		a.ConstructName(),
	)

	err = h.store.WriteFile(fp, file)
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

// swagger:route PUT /assets assets putAsset
//
// Update an asset in the collection
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
func (h *AssetsHandler) PutAsset(rw http.ResponseWriter, r *http.Request) {
	h.logger.Info("Processing PUT Asset request")

	a := &models.Asset{}
	json := r.FormValue("metadata")
	if json == "" {
		response.RespondWithMessage(rw, http.StatusBadRequest, "Invalid JSON part of the request")
		return
	}

	err := utils.FromJSONString(a, json)
	if err != nil {
		response.RespondWithMessage(rw, http.StatusBadRequest, response.MessageInvalidJsonFormat)
		return
	}
	
	err = a.Validate()
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
	
	fn := a.ConstructName()
	fp := filepath.Join(
		a.Category,
		a.ID,
		fn,
	)

	err = h.store.OverwriteFile(fp, file)
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

// swagger:route DELETE /assets assets deleteAsset
//
// Remove Asset from the collection
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
func (h *AssetsHandler) DeleteAsset(rw http.ResponseWriter, r *http.Request) {
	h.logger.Info("Processing DELETE Asset request")

	i := &models.Asset{}
	err := utils.FromJSON(i, r.Body)
	if err != nil {
		response.RespondWithMessage(rw, http.StatusBadRequest, response.MessageInvalidJsonFormat)
		return
	}

	err = i.Validate()
	if err != nil {
		response.RespondWithMessage(rw, http.StatusBadRequest, response.MessaggeInvalidData)
		return
	}

	fn := i.ConstructName()
	fp := filepath.Join(
		i.Category,
		i.ID,
		fn,
	)

	err = h.store.DeleteFile(fp)
	if err != nil {
		if err == files.ErrNotFound {
			response.RespondWithMessage(rw, http.StatusNotFound, "Asset was not found")
			return
		}
		response.RespondWithMessage(rw, http.StatusBadRequest, "Failed to delete the asset")
		return
	}

	rw.WriteHeader(http.StatusNoContent)
}
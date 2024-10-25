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

type AssetHandler[T models.Asset] struct {
    baseUrl   string
    logger    *slog.Logger
    store     files.Storage
    cache     caches.Cache
    signedUrl signedurl.SignedUrl
}

// NewAssetHandler creates a new AssetHandler for a specific type.
func NewAssetHandler[T models.Asset](baseUrl string, s files.Storage, l *slog.Logger, c caches.Cache) *AssetHandler[T] {
    return &AssetHandler[T]{
        baseUrl: baseUrl,
        store:   s,
        logger:  l,
        cache:   c,
        signedUrl: *signedurl.NewSignedUrl(
            "Secret key my boy",
            baseUrl+"/assets",
            time.Duration(5*int(time.Minute)),
        ),
    }
}


// swagger:route GET /files files getFileUrl
//
// Return a signed url to requested file
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
func (h *AssetHandler[T]) GetFileUrl(rw http.ResponseWriter, r *http.Request) {
	h.logger.Info("Processing GET File URL request")
	
	var asset T
	err := utils.FromJSON(asset, r.Body)
	if err != nil {
		response.RespondWithMessage(rw, http.StatusBadRequest, "Invalid JSON data")
		return
	}

	err = asset.Validate()
	if err != nil {
		response.RespondWithMessage(rw, http.StatusBadRequest, "Invalid file data")
		return
	}

	err = h.store.CheckAsset(asset)
	if err != nil {
		if err == files.ErrNotFound {
			http.Error(rw, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(rw, err.Error(), http.StatusInternalServerError)
	}

	tmpId := caches.GenerateUUID()
	h.cache.Set(tmpId, asset)
	url := h.signedUrl.GenerateSignedUrl(tmpId)

    urlResponse := response.FileUrlResponse{
        URL:      url,
    }

	response.RespondWithJSON(rw, http.StatusOK, urlResponse)
}

// swagger:route GET /{id}&{expires}&{signature} files getFile
//
// Return an file from collection. Can only be accessed by signed URLs
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
func (h *AssetHandler[T]) GetFile(rw http.ResponseWriter, r *http.Request) {
	h.logger.Info("Processing GET File request")

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

// swagger:route POST /files files postFile
//
// Add an file to the existing collection
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
func (h *AssetHandler[T]) PostFile(rw http.ResponseWriter, r *http.Request) {
	h.logger.Info("Processing POST File request")

	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		response.RespondWithMessage(rw, http.StatusBadRequest, "Unable to parse form data")
		return
	}

	json := r.FormValue("metadata")
	if json == "" {
		response.RespondWithMessage(rw, http.StatusBadRequest, "Invalid JSON part of the request")
		return
	}

	var file T
	err = utils.FromJSONString(file, json)
	if err != nil {
		response.RespondWithMessage(rw, http.StatusBadRequest, "Invalid data format")
		return
	}

	err = file.Validate()
	if err != nil {
		response.RespondWithMessage(rw, http.StatusBadRequest, "Invalid file data")
		return
	}

	f, _, err := r.FormFile("file")
	if err != nil {
		response.RespondWithMessage(rw, http.StatusBadRequest, "Error reading file from request")
		return
	}
	defer f.Close()

	err = h.store.AddAsset(file, f)
	if err != nil {
		if err == files.ErrAlreadyExists {
			response.RespondWithMessage(rw, http.StatusForbidden, "File already exists")
			return
		}
		if err == files.ErrNotFound {
			response.RespondWithMessage(rw, http.StatusBadRequest, "Collection doesn't exist")
			return
		}
		response.RespondWithMessage(rw, http.StatusInternalServerError, "Failed to create the file")
		return
	}

	response.RespondWithMessage(rw, http.StatusCreated, "File uploaded sucessfully")
}

// swagger:route PUT /files files putFile
//
// Update an file in the collection
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
func (h *AssetHandler[T]) PutFile(rw http.ResponseWriter, r *http.Request) {
	h.logger.Info("Processing PUT File request")

	json := r.FormValue("metadata")
	if json == "" {
		response.RespondWithMessage(rw, http.StatusBadRequest, "Invalid JSON part of the request")
		return
	}

	var file T
	err := utils.FromJSONString(file, json)
	if err != nil {
		response.RespondWithMessage(rw, http.StatusBadRequest, "Invalid data format")
		return
	}
	
	err = file.Validate()
	if err != nil {
		response.RespondWithMessage(rw, http.StatusBadRequest, "Invalid file data")
		return
	}

	f, _, err := r.FormFile("file")
	if err != nil {
		response.RespondWithMessage(rw, http.StatusBadRequest, "Error reading file from request")
		return
	}
	defer f.Close()

	err = h.store.OverwriteAsset(file, f)
	if err != nil {
		if err == files.ErrNotFound {
			response.RespondWithMessage(rw, http.StatusNotFound, "File does not exist")
			return
		}
		response.RespondWithMessage(rw, http.StatusInternalServerError, "Failed to update the file")
		return
	}

	response.RespondWithMessage(rw, http.StatusOK, "File updated sucessfully")
}

// swagger:route PUT /files/update files putFileData
//
// Updates the file data without changing the file contents
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
func (h *AssetHandler[T]) PutFileData(rw http.ResponseWriter, r *http.Request) {
	h.logger.Info("Processing PUT File request")

	request := &models.PutRequest[T]{}
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

	err = h.store.UpdateAsset(request.Existing, request.New)
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

// swagger:route DELETE /files files deleteFile
//
// Remove file from the collection
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
func (h *AssetHandler[T]) DeleteFile(rw http.ResponseWriter, r *http.Request) {
	h.logger.Info("Processing DELETE File request")

	var file T
	err := utils.FromJSON(file, r.Body)
	if err != nil {
		response.RespondWithMessage(rw, http.StatusBadRequest, "Invalid data format")
		return
	}

	err = file.Validate()
	if err != nil {
		response.RespondWithMessage(rw, http.StatusBadRequest, "Invalid file data")
		return
	}

	err = h.store.DeleteAsset(file)
	if err != nil {
		if err == files.ErrNotFound {
			response.RespondWithMessage(rw, http.StatusNotFound, "File was not found")
			return
		}
		response.RespondWithMessage(rw, http.StatusBadRequest, "Failed to delete the file")
		return
	}

	rw.WriteHeader(http.StatusNoContent)
}
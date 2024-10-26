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

// Handler for managing collections
type CollectionsHandler struct {
	baseUrl		string
	logger		*slog.Logger
	store		files.Storage
	cache		caches.Cache
	signedUrl	signedurl.SignedUrl
}

func NewCollections(baseUrl string, s files.Storage, l *slog.Logger, c caches.Cache) *CollectionsHandler {
	logger := l.With(slog.String("handler", "collections")) // TODO: do this when initializing logger in the main (you can pass the same logger to the store then)

	return &CollectionsHandler{
		baseUrl: baseUrl,
		store:   s,
		logger:  logger,
		cache:   c,
		signedUrl: *signedurl.NewSignedUrl(
			"Secret key my boy",
			baseUrl+"/files", // TODO: accept as parameter from main.go
			time.Duration(5*int(time.Minute)),
		),
	}
}

// swagger:route GET /collections collections getCollection
//
// Returns Collection contents
//
// consumes:
//	- application/json
//
// produces:
//	- application/json
//
// Responses:
// 	200: getCollectionResponse
//  400: message
//	404: message
// 	500: message
func (h *CollectionsHandler) GetCollection(rw http.ResponseWriter, r *http.Request) {
	h.logger.Info("Processing GET Collection request")

	collection := &models.AssetsCollection{}
	err := utils.FromJSON(collection, r.Body)
	if err != nil {
		response.RespondWithMessage(rw, http.StatusBadRequest, response.MessageInvalidJsonFormat)
		return
	}

	err = collection.Validate()
	if err != nil {
		response.RespondWithMessage(rw, http.StatusBadRequest, response.MessaggeInvalidData)
		return
	}

	f, err := h.store.ListCollectionContents(collection)
	if err != nil {
		if err == files.ErrNotFound {
			response.RespondWithMessage(rw, http.StatusNotFound, response.MessageNotFound)
			return
		}
		response.RespondWithMessage(rw, http.StatusInternalServerError, response.MessageFailedRead)
		return
	}

	cr := &response.DirectoryContentsResponse{Contents: f}

	response.RespondWithJSON(rw, http.StatusOK, cr)
}

// swagger:route POST /collections collections postCollection
//
// Create a new collection
//
// consumes:
//	- application/json
//
// produces:
//	- application/json
//
// Responses:
// 	204: empty
//  400: message
// 	403: message
//	404: message
// 	500: message
func (h *CollectionsHandler) PostCollection(rw http.ResponseWriter, r *http.Request) {
	h.logger.Info("Processing POST Collection request")

	collection := &models.AssetsCollection{}
	err := utils.FromJSON(collection, r.Body)
	if err != nil {
		response.RespondWithMessage(rw, http.StatusBadRequest, response.MessageInvalidJsonFormat)
		return
	}

	err = collection.Validate()
	if err != nil {
		response.RespondWithMessage(rw, http.StatusBadRequest, response.MessaggeInvalidData)
		return
	}

	err = h.store.CreateCollection(collection)
	if err != nil {
		if err == files.ErrNotFound {
			response.RespondWithMessage(rw, http.StatusNotFound, response.MessageNotFound)
			return
		}
		if err == files.ErrAlreadyExists {
			response.RespondWithMessage(rw, http.StatusForbidden, response.MessageAlreadyExists)
			return
		}
		response.RespondWithMessage(rw, http.StatusInternalServerError, response.MessageFailedCreate)
		return
	}

	response.RespondWithNoContent(rw)
}

// swagger:route PUT /collections collections putCollection
//
// Update existing collection id or category. Allows moving to the different category, but it won't create any new categories
//
// consumes:
//	- application/json
//
// produces:
//	- application/json
//
// Responses:
// 	204: empty
//  400: message
//	404: message
// 	500: message
func (h *CollectionsHandler) PutCollection(rw http.ResponseWriter, r *http.Request) {
	h.logger.Info("Processing PUT Collection request")
	
	collection := &models.PutRequest[models.AssetsCollection]{}
	err := utils.FromJSON(collection, r.Body)
	if err != nil {
		response.RespondWithMessage(rw, http.StatusBadRequest, response.MessageInvalidJsonFormat)
		return
	}

	err = collection.Existing.Validate()
	if err != nil {
		response.RespondWithMessage(rw, http.StatusBadRequest, response.MessaggeInvalidData)
		return
	}

	err = collection.New.Validate()
	if err != nil {
		response.RespondWithMessage(rw, http.StatusBadRequest, response.MessaggeInvalidData)
		return
	}

	err = h.store.UpdateCollection(&collection.Existing, &collection.New)
	if err != nil {
		if err == files.ErrNotFound {
			response.RespondWithMessage(rw, http.StatusNotFound, response.MessageNotFound)
			return
		}
		response.RespondWithMessage(rw, http.StatusInternalServerError, response.MessageFailedUpdate)
		return
	}

	response.RespondWithNoContent(rw)
}

// swagger:route DELETE /collections collections deleteCollection
//
// Delete existing collection
//
// consumes:
//	- application/json
//
// produces:
//	- application/json
//
// Responses:
// 	204: message
//  400: message
//  404: message
// 	500: message
func (h *CollectionsHandler) DeleteCollection(rw http.ResponseWriter, r *http.Request) {
	h.logger.Info("Processing DELETE Collection request")

	collection := &models.AssetsCollection{}
	err := utils.FromJSON(collection, r.Body)
	if err != nil {
		response.RespondWithMessage(rw, http.StatusBadRequest, response.MessageInvalidJsonFormat)
		return
	}

	err = collection.Validate()
	if err != nil {
		response.RespondWithMessage(rw, http.StatusBadRequest, response.MessaggeInvalidData)
		return
	}

	err = h.store.DeleteCollection(collection)
	if err != nil {
		if err == files.ErrNotFound {
			response.RespondWithMessage(rw, http.StatusNotFound, response.MessageNotFound)
			return
		}
		response.RespondWithMessage(rw, http.StatusInternalServerError, response.MessageFailedDelete)
		return
	}

	response.RespondWithNoContent(rw)
}
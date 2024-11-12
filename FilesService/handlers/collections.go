package handlers

import (
	"log/slog"
	"net/http"

	"github.com/KrzysztofSieczkiewicz/go--model-viewer-backend/FilesService/files"
	"github.com/KrzysztofSieczkiewicz/go--model-viewer-backend/FilesService/models"
	"github.com/KrzysztofSieczkiewicz/go--model-viewer-backend/FilesService/response"
	"github.com/KrzysztofSieczkiewicz/go--model-viewer-backend/FilesService/utils"
)

/*
POST COLLECTION
curl -v -X POST http://localhost:9090/models/collections -H "Content-Type: application/json" -d "{\"category\":{\"path\":\"random/testNew\"}, \"id\":\"collection\"}"

GET COLLECTION
curl -v -X GET http://localhost:9090/models/collections -H "Content-Type: application/json" -d "{\"category\":{\"path\":\"random/testNew\"}, \"id\":\"collection\"}"

PUT COLLECTION
curl -v -X PUT http://localhost:9090/models/collections -H "Content-Type: application/json" -d "{\"existing\":{\"category\":{\"path\":\"random/testNew\"}, \"id\":\"collection\"},\"new\":{\"category\":{\"path\":\"random/testNew\"}, \"id\":\"collectionNew\"}}"

DELETE COLLECTION
curl -v -X DELETE http://localhost:9090/models/collections -H "Content-Type: application/json" -d "{\"category\":{\"path\":\"random/testNew\"}, \"id\":\"collectionNew\"}"
*/

// Handler for managing collections
type CollectionsHandler struct {
	baseUrl		string
	logger		*slog.Logger
	store		files.Storage
}

func NewCollections(baseUrl string, s files.Storage, slogger *slog.Logger) *CollectionsHandler {
	logger := slogger.With(slog.String("handler", "collections"))

	return &CollectionsHandler{
		baseUrl: baseUrl,
		store:   s,
		logger:  logger,
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

	collection := &models.Collection{}
	err := utils.FromJSON(collection, r.Body)
	if err != nil {
		response.RespondWithMessage(rw, http.StatusBadRequest, response.MessageInvalidJsonFormat)
		return
	}

	err = collection.Validate()
	if err != nil {
		response.RespondWithMessage(rw, http.StatusBadRequest, response.MessageInvalidData)
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

	cr := &models.DirectoryContentsResponse{Contents: f}

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
//	404: message
// 	500: message
func (h *CollectionsHandler) PostCollection(rw http.ResponseWriter, r *http.Request) {
	h.logger.Info("Processing POST Collection request")

	collection := &models.Collection{}
	err := utils.FromJSON(collection, r.Body)
	if err != nil {
		response.RespondWithMessage(rw, http.StatusBadRequest, response.MessageInvalidJsonFormat)
		return
	}

	err = collection.Validate()
	if err != nil {
		response.RespondWithMessage(rw, http.StatusBadRequest, response.MessageInvalidData)
		return
	}

	err = h.store.CreateCollection(collection)
	if err != nil {
		if err == files.ErrNotFound {
			response.RespondWithMessage(rw, http.StatusNotFound, response.MessageNotFound)
			return
		}
		if err == files.ErrAlreadyExists {
			response.RespondWithMessage(rw, http.StatusBadRequest, response.MessageAlreadyExists)
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
	
	collection := &models.PutRequest[models.Collection]{}
	err := utils.FromJSON(collection, r.Body)
	if err != nil {
		response.RespondWithMessage(rw, http.StatusBadRequest, response.MessageInvalidJsonFormat)
		return
	}

	err = collection.Existing.Validate()
	if err != nil {
		response.RespondWithMessage(rw, http.StatusBadRequest, response.MessageInvalidData)
		return
	}

	err = collection.New.Validate()
	if err != nil {
		response.RespondWithMessage(rw, http.StatusBadRequest, response.MessageInvalidData)
		return
	}

	err = h.store.UpdateCollection(&collection.Existing, &collection.New)
	if err != nil {
		if err == files.ErrNotFound {
			response.RespondWithMessage(rw, http.StatusNotFound, response.MessageNotFound)
			return
		}
		if err == files.ErrAlreadyExists {
			response.RespondWithMessage(rw, http.StatusBadRequest, response.MessageAlreadyExists)
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

	collection := &models.Collection{}
	err := utils.FromJSON(collection, r.Body)
	if err != nil {
		response.RespondWithMessage(rw, http.StatusBadRequest, response.MessageInvalidJsonFormat)
		return
	}

	err = collection.Validate()
	if err != nil {
		response.RespondWithMessage(rw, http.StatusBadRequest, response.MessageInvalidData)
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
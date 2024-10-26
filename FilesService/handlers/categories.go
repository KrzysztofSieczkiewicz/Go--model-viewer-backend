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

// Handler for managing categories
type CategoriesHandler struct {
	baseUrl		string
	logger		*slog.Logger
	store		files.Storage
	cache		caches.Cache
	signedUrl	signedurl.SignedUrl
}

func NewCategories(baseUrl string, s files.Storage, l *slog.Logger, c caches.Cache) *CategoriesHandler {
	logger := l.With(slog.String("handler", "categories")) // TODO: do this when initializing logger in the main (you can pass the same logger to the store then)

	return &CategoriesHandler{
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

// swagger:route GET /categories categories getCategory
//
// List files available in the category
//
// consumes:
//	- application/json
//
// produces:
//	- application/json
//
// Responses:
// 	200: directoryContents
//  400: message
//	404: message
// 	500: message
func (h *CategoriesHandler) GetCategory(rw http.ResponseWriter, r *http.Request) {
	h.logger.Info("Processing GET Category request")

	category := &models.Category{}
	err := utils.FromJSON(category, r.Body)
	if err != nil {
		response.RespondWithMessage(rw, http.StatusBadRequest, response.MessageInvalidJsonFormat)
		return
	}

	err = category.Validate()
	if err != nil {
		response.RespondWithMessage(rw, http.StatusBadRequest, response.MessaggeInvalidData)
		return
	}

	f, err := h.store.ListCategoryContents(category)
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

// swagger:route POST /categories categories postCategory
//
// Creates a category or categories path
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
// 	403: message
// 	500: message
func (h *CategoriesHandler) PostCategory(rw http.ResponseWriter, r *http.Request) {
	h.logger.Info("Processing POST Category request")

	category := &models.Category{}
	err := utils.FromJSON(category, r.Body)
	if err != nil {
		response.RespondWithMessage(rw, http.StatusBadRequest, response.MessageInvalidJsonFormat)
	}

	err = category.Validate()
	if err != nil {
		response.RespondWithMessage(rw, http.StatusBadRequest, response.MessaggeInvalidData)
		return
	}

	err = h.store.CreateCategory(category)
	if err != nil {
		if err == files.ErrAlreadyExists {
			response.RespondWithMessage(rw, http.StatusForbidden, response.MessageAlreadyExists)
			return
		}
		response.RespondWithMessage(rw, http.StatusInternalServerError, response.MessageFailedCreate)
		return
	}

	response.RespondWithNoContent(rw)
}

// swagger:route PUT /categories categories putCategory
//
// Updates existing Category, allows for moving. Doesn't create new filepaths
//
// consumes:
//	- application/json
//
// produces:
//	- application/json
//
// Responses:
// 	200: message
//  400: message
//	404: message
// 	500: message
func (h *CategoriesHandler) PutCategory(rw http.ResponseWriter, r *http.Request) {
	h.logger.Info("Processing PUT Category request")

	category := &models.PutRequest[models.Category]{}
	err := utils.FromJSON(category, r.Body)
	if err != nil {
		response.RespondWithMessage(rw, http.StatusBadRequest, response.MessageInvalidJsonFormat)
	}

	err = category.Existing.Validate()
	if err != nil {
		response.RespondWithMessage(rw, http.StatusBadRequest, response.MessaggeInvalidData)
		return
	}

	err = category.New.Validate()
	if err != nil {
		response.RespondWithMessage(rw, http.StatusBadRequest, response.MessaggeInvalidData)
		return
	}

	err = h.store.UpdateCategory(&category.Existing, &category.New)
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

	response.RespondWithMessage(rw, http.StatusOK, response.MessageUpdateSuccessful)
}

// swagger:route DELETE /categories categories deleteCategory
//
// Deletes the Category, requires being empty beforehand
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
//	403: message
//	404: message
// 	500: message
func (h *CategoriesHandler) DeleteCategory(rw http.ResponseWriter, r *http.Request) {
	h.logger.Info("Processing DELETE Category request")

	category := &models.Category{}
	err := utils.FromJSON(category, r.Body)
	if err != nil {
		response.RespondWithMessage(rw, http.StatusBadRequest, response.MessageInvalidJsonFormat)
		return
	}

	err = category.Validate()
	if err != nil {
		response.RespondWithMessage(rw, http.StatusBadRequest, response.MessaggeInvalidData)
		return
	}

	err = h.store.DeleteCategory(category)
	if err != nil {
		if err == files.ErrDirNotEmpty {
			response.RespondWithMessage(rw, http.StatusForbidden, response.MessageDirectoryNotEmpty)
			return
		}
		response.RespondWithMessage(rw, http.StatusInternalServerError, response.MessageFailedDelete)
		return
	}

	response.RespondWithNoContent(rw)
}
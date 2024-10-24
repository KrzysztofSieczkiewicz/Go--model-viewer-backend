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

// Handler for managing collections and categories
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
// Return ImageSet details and available Images
//
// consumes:
//	- application/json
//
// produces:
//	- application/json
//
// Responses:
// 	200: getImageSet
//  400: message
//	404: message
// 	500: message
func (h *CollectionsHandler) GetCollection(rw http.ResponseWriter, r *http.Request) {
	h.logger.Info("Processing GET Collection request")

	c := &models.Collection{}
	err := utils.FromJSON(c, r.Body)
	if err != nil {
		response.RespondWithMessage(rw, http.StatusBadRequest, response.MessageInvalidJsonFormat)
		return
	}

	is := &models.ImageSet{}
	err = utils.FromJSON(is, r.Body)
	if err != nil {
		response.RespondWithMessage(rw, http.StatusBadRequest, response.MessageInvalidJsonFormat)
		return
	}

	err = is.Validate()
	if err != nil {
		response.RespondWithMessage(rw, http.StatusBadRequest, response.MessaggeInvalidData)
		return
	}

	fp := filepath.Join(is.Category, is.ID)

	f, err := h.store.ListFiles(fp)
	if err != nil {
		if err == files.ErrNotFound {
			response.RespondWithMessage(rw, http.StatusNotFound, "Requested image set doesn't exist")
			return
		}
		response.RespondWithMessage(rw, http.StatusInternalServerError, "Unable to retrieve ImageSet data")
		return
	}

	i := &models.Images{}
	err = i.DeconstructImageNames(f)
	if err != nil {
		response.RespondWithMessage(rw, http.StatusInternalServerError, "Unable to retrieve images list")
		return
	}

	response.RespondWithJSON(rw, http.StatusOK, i)
}
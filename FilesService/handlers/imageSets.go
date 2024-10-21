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

/*
Example curls:
GET IMAGESET:
curl -v -X GET http://localhost:9090/imageSets -H "Content-Type: application/json" -d "{\"category\":\"random/test\",\"id\":\"1\"}"

POST IMAGESET:
curl -v -i -X POST http://localhost:9090/imageSets -H "Content-Type: application/json" -d "{\"category\":\"random/test\",\"id\":\"1\"}"

PUT IMAGESET:
curl -v -i -X PUT http://localhost:9090/imageSets -H "Content-Type: application/json" -d "{\"existing\":{\"category\":\"random/test\",\"id\":\"1\"}, \"new\":{\"category\":\"random/test\",\"id\":\"2\"}}"

DELETE IMAGESET:
curl -v -i -X DELETE http://localhost:9090/imageSets -H "Content-Type: application/json" -d "{\"category\":\"random/test\",\"id\":\"1\"}"

GET CATEGORY:
curl -v -X GET http://localhost:9090/imageCategories -H "Content-Type: application/json" -d "{\"filepath\":\"random/test\"}"

POST CATEGORY:
curl -v -X POST http://localhost:9090/imageCategories -H "Content-Type: application/json" -d "{\"filepath\":\"random/test\"}"

PUT CATEGORY:
curl -v -X PUT http://localhost:9090/imageCategories -H "Content-Type: application/json" -d "{\"existing\":{\"filepath\":\"random/test\"}, \"new\":{\"filepath\":\"random/test3\"}}"

DELETE CATEGORY:
curl -v -i -X DELETE http://localhost:9090/imageCategories -H "Content-Type: application/json" -d "{\"filepath\":\"random/test\"}"
*/

// Handler for managing imageSets and categories
type ImageSetsHandler struct {
	baseUrl		string
	logger		*slog.Logger
	store		files.Storage
	cache		caches.Cache
	signedUrl	signedurl.SignedUrl
}

func NewImageSets(baseUrl string, s files.Storage, l *slog.Logger, c caches.Cache) *ImageSetsHandler {
	logger := l.With(slog.String("endpoint", "imageSets"))

	return &ImageSetsHandler{
		baseUrl: baseUrl,
		store: s, 
		logger: logger,
		cache: c,
		signedUrl: *signedurl.NewSignedUrl(
			"Secret key my boy",
			baseUrl + "/files",
			time.Duration(5 * int(time.Minute)),
		),
	}
}

// swagger:route GET /imageSets imageSets getImageSet
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
func (h *ImageSetsHandler) GetImageSet(rw http.ResponseWriter, r *http.Request) {
	h.logger.Info("Processing GET ImageSet request")

	is := &models.ImageSet{}
	err := utils.FromJSON(is, r.Body)
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

// swagger:route POST /imageSets imageSets postImageSet
//
// Create a new image set
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
// 	500: message
func (h *ImageSetsHandler) PostImageSet(rw http.ResponseWriter, r *http.Request) {
	h.logger.Info("Processing POST ImageSet request")

	is := &models.ImageSet{}
	err := utils.FromJSON(is, r.Body)
	if err != nil {
		response.RespondWithMessage(rw, http.StatusBadRequest, response.MessageInvalidJsonFormat)
		return
	}

	err = is.Validate()
	if err != nil {
		response.RespondWithMessage(rw, http.StatusBadRequest, response.MessaggeInvalidData)
		return
	}

	err = h.store.IfExists(is.Category)
	if err != nil {
		response.RespondWithMessage(rw, http.StatusBadRequest, "Invalid category")
		return
	}

	fp := filepath.Join(is.Category, is.ID)

	err = h.store.CreateDirectory(fp)
	if err != nil {
		if err == files.ErrAlreadyExists {
			response.RespondWithMessage(rw, http.StatusForbidden, "ImageSet already exists")
			return
		}
		response.RespondWithMessage(rw, http.StatusInternalServerError, "Unable to create ImageSet")
		return
	}

	response.RespondWithNoContent(rw)
}

// swagger:route PUT /imageSets imageSets putImageSet
//
// Update existing imageset id or category. Allows to move imageset to the different category, but it must be initialized beforehand
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
func (h *ImageSetsHandler) PutImageSet(rw http.ResponseWriter, r *http.Request) {
	h.logger.Info("Processing PUT ImageSet request")
	
	is := &models.PutImageSetRequest{}
	err := utils.FromJSON(is, r.Body)
	if err != nil {
		response.RespondWithMessage(rw, http.StatusBadRequest, response.MessageInvalidJsonFormat)
		return
	}

	err = is.Existing.Validate()
	if err != nil {
		response.RespondWithMessage(rw, http.StatusBadRequest, response.MessaggeInvalidData)
		return
	}

	err = is.New.Validate()
	if err != nil {
		response.RespondWithMessage(rw, http.StatusBadRequest, response.MessaggeInvalidData)
		return
	}

	ofp := filepath.Join(is.Existing.Category, is.Existing.ID)
	nfp := filepath.Join(is.New.Category, is.New.ID)

	err = h.store.ChangeDirectory(ofp, nfp)
	if err != nil {
		if err == files.ErrNotFound {
			response.RespondWithMessage(rw, http.StatusNotFound, "Unable to find ImageSet")
			return
		}
		response.RespondWithMessage(rw, http.StatusInternalServerError, "Unable to update ImageSet")
		return
	}

	response.RespondWithNoContent(rw)
}

// swagger:route DELETE /imageSets imageSets deleteImageSet
//
// Delete existing imageset
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
func (h *ImageSetsHandler) DeleteImageSet(rw http.ResponseWriter, r *http.Request) {
	h.logger.Info("Processing DELETE ImageSet request")

	is := &models.ImageSet{}
	err := utils.FromJSON(is, r.Body)
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

	err = h.store.DeleteFiles(fp)
	if err != nil {
		if err == files.ErrNotFound {
			response.RespondWithMessage(rw, http.StatusNotFound, "Unable to find ImageSet")
			return
		}
		response.RespondWithMessage(rw, http.StatusInternalServerError, "Unable to clear ImageSet contents")
		return
	}

	err = h.store.DeleteDirectory(fp)
	if err != nil {
		response.RespondWithMessage(rw, http.StatusInternalServerError, "Unable to remove ImageSet directory")
		return
	}

	response.RespondWithNoContent(rw)
}

// swagger:route GET /imageCategories imageSets getCategory
//
// List subdirectories available in the category
//
// consumes:
//	- application/json
//
// produces:
//	- application/json
//
// Responses:
// 	200: categoryContents
//  400: message
//	404: message
// 	500: message
func (h *ImageSetsHandler) GetCategory(rw http.ResponseWriter, r *http.Request) {
	h.logger.Info("Processing GET ImageSet Category request")

	c := &models.Category{}
	err := utils.FromJSON(c, r.Body)
	if err != nil {
		response.RespondWithMessage(rw, http.StatusBadRequest, response.MessageInvalidJsonFormat)
	}

	// TODO: PUT VALIDATION HERE

	f, err := h.store.ListDirectories(c.Filepath)
	if err != nil {
		if err == files.ErrNotFound {
			response.RespondWithMessage(rw, http.StatusNotFound, "Category doesn't exist")
			return
		}
		response.RespondWithMessage(rw, http.StatusInternalServerError, "Unable to retrieve Category")
		return
	}

	is := &response.CategoryResponse{Directories: f}

	response.RespondWithJSON(rw, http.StatusOK, is)
}

// swagger:route POST /imageCategories imageSets postCategory
//
// Creates a requested directory path
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
func (h *ImageSetsHandler) PostCategory(rw http.ResponseWriter, r *http.Request) {
	h.logger.Info("Processing POST ImageSet Category request")

	c := &models.Category{}
	err := utils.FromJSON(c, r.Body)
	if err != nil {
		response.RespondWithMessage(rw, http.StatusBadRequest, response.MessageInvalidJsonFormat)
	}

	err = h.store.CreateDirectory(c.Filepath)
	if err != nil {
		if err == files.ErrAlreadyExists {
			response.RespondWithMessage(rw, http.StatusForbidden, "Directory already exists")
			return
		}
		response.RespondWithMessage(rw, http.StatusInternalServerError, "Unable to create Category")
		return
	}

	response.RespondWithNoContent(rw)
}

// swagger:route PUT /imageCategories imageSets putCategory
//
// Update existing Category
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
func (h *ImageSetsHandler) PutCategory(rw http.ResponseWriter, r *http.Request) {
	h.logger.Info("Processing PUT ImageSet Category request")

	c := &models.PutCategoryRequest{}
	err := utils.FromJSON(c, r.Body)
	if err != nil {
		response.RespondWithMessage(rw, http.StatusBadRequest, response.MessageInvalidJsonFormat)
	}

	err = h.store.ChangeDirectory(c.Existing.Filepath, c.New.Filepath)
	if err != nil {
		if err == files.ErrNotFound {
			response.RespondWithMessage(rw, http.StatusNotFound, "Unable to find Category")
			return
		}
		if err == files.ErrAlreadyExists {
			response.RespondWithMessage(rw, http.StatusBadRequest, "Category already exists")
			return
		}
		response.RespondWithMessage(rw, http.StatusInternalServerError, "Unable to update Category")
		return
	}

	response.RespondWithMessage(rw, http.StatusOK, "Category updated successfully")
}

// swagger:route DELETE /imageCategories imageSets deleteCategory
//
// Delete existing category, requires being emptied beforehand
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
//	403: message
//	404: message
// 	500: message
func (h *ImageSetsHandler) DeleteCategory(rw http.ResponseWriter, r *http.Request) {
	h.logger.Info("Processing DELETE ImageSet Category request")

	c := &models.Category{}
	err := utils.FromJSON(c, r.Body)
	if err != nil {
		response.RespondWithMessage(rw, http.StatusBadRequest, response.MessageInvalidJsonFormat)
		return
	}

	err = h.store.DeleteDirectory(c.Filepath)
	if err != nil {
		if err == files.ErrDirNotEmpty {
			response.RespondWithMessage(rw, http.StatusForbidden, "Category is not empty")
			return
		}
		response.RespondWithMessage(rw, http.StatusInternalServerError, "Unable to remove Category")
		return
	}

	response.RespondWithMessage(rw, http.StatusOK, "Category removed successfully")
}
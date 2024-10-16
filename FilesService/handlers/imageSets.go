package handlers

import (
	"log/slog"
	"net/http"
	"net/url"
	"path/filepath"
	"time"

	"github.com/KrzysztofSieczkiewicz/go--model-viewer-backend/FilesService/caches"
	"github.com/KrzysztofSieczkiewicz/go--model-viewer-backend/FilesService/data"
	"github.com/KrzysztofSieczkiewicz/go--model-viewer-backend/FilesService/files"
	"github.com/KrzysztofSieczkiewicz/go--model-viewer-backend/FilesService/response"
	"github.com/KrzysztofSieczkiewicz/go--model-viewer-backend/FilesService/signedurl"
	"github.com/KrzysztofSieczkiewicz/go--model-viewer-backend/FilesService/utils"
)

/*
Example curls:
GET IMAGESET URL:
curl -v -X GET http://localhost:9090/imageSets/random%2Ftest%2F/1

POST IMAGESET:
curl -v -i -X POST http://localhost:9090/imageSets/random%2Ftest%2F/1

PUT IMAGESET:
curl -v -i -X PUT http://localhost:9090/imageSets/random%2Ftest%2F/1 -H "Content-Type: application/json" -d "{\"category\":\"random/test\",\"id\":\"4\"}"

DELETE IMAGESET:
curl -v -i -X DELETE http://localhost:9090/imageSets/random%2Ftest%2F/1

GET CATEGORY:
curl -v -X GET http://localhost:9090/imageCategories/random%2Ftest%2F

POST CATEGORY:
curl -v -X POST http://localhost:9090/imageCategories/random%2Ftest%2F1

PUT CATEGORY:
curl -v -X PUT http://localhost:9090/imageCategories/random%2Ftest%2F1 -H "Content-Type: application/json" -d "{\"category\":\"random/test2\"}"

DELETE CATEGORY:
curl -v -i -X DELETE http://localhost:9090/imageCategories/random%2Ftest%2F/1
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

// swagger:route GET /imageSets/{category}/{id} imageSets getImageSet
//
// Return ImageSet details and available Images
//
// produces:
//	- application/json
//
// Responses:
// 	200: imageSet
//  400: message
// 	403: message
//	404: message
// 	500: message
func (h *ImageSetsHandler) GetImageSet(rw http.ResponseWriter, r *http.Request) {
	h.logger.Info("Processing GET ImageSet request")

	c, err := url.QueryUnescape( r.PathValue("category") )
	if err != nil {
		response.RespondWithMessage(rw, http.StatusBadRequest, "Cannot decode the category from url")
		return
	}

	id, err := url.QueryUnescape( r.PathValue("id") )
	if err != nil {
		response.RespondWithMessage(rw, http.StatusBadRequest, "Cannot decode the id from url")
		return
	}

	fp := filepath.Join(c, id)

	f, err := h.store.ListFiles(fp)
	if err != nil {
		if err == files.ErrNotFound {
			response.RespondWithMessage(rw, http.StatusNotFound, "ImageSet doesn't exist")
			return
		}
		if err == files.ErrNotDirectory {
			response.RespondWithMessage(rw, http.StatusForbidden, "Requested path is not a directory")
			return
		}
		response.RespondWithMessage(rw, http.StatusInternalServerError, "Unable to retrieve ImageSet data")
		return
	}

	i := &data.Images{}
	err = i.DeconstructImageNames(f)
	if err != nil {
		response.RespondWithMessage(rw, http.StatusInternalServerError, "Unable to deconstruct image names:")
		return
	}

	is := &data.ImageSet{
		ID: id,
		Category: c,
		Images: *i,
	}

	response.RespondWithJSON(rw, http.StatusOK, is)
}

// swagger:route POST /imageSets/{category}/{id} imageSets postImageSet
//
// Create a new image set
//
// produces:
//	- application/json
//
// Responses:
// 	201: message
//  400: message
// 	403: message
// 	500: message
func (h *ImageSetsHandler) PostImageSet(rw http.ResponseWriter, r *http.Request) {
	h.logger.Info("Processing POST ImageSet request")

	c, err := url.QueryUnescape( r.PathValue("category") )
	if err != nil {
		response.RespondWithMessage(rw, http.StatusBadRequest, "Cannot decode the category from url")
		return
	}
	id, err := url.QueryUnescape( r.PathValue("id") )
	if err != nil {
		response.RespondWithMessage(rw, http.StatusBadRequest, "Cannot decode the id from url")
		return
	}

	fp := filepath.Join(c, id)

	err = h.store.CreateDirectory(fp)
	if err != nil {
		if err == files.ErrAlreadyExists {
			response.RespondWithMessage(rw, http.StatusForbidden, "ImageSet already exists")
			return
		}
		response.RespondWithMessage(rw, http.StatusInternalServerError, "Unable to create ImageSet")
		return
	}

	response.RespondWithMessage(rw, http.StatusCreated, "ImageSet created successfully")
}

// swagger:route PUT /imageSets/{category}/{id} imageSets putImageSet
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
// 	200: message
//  400: message
// 	403: message
//	404: message
// 	500: message
func (h *ImageSetsHandler) PutImageSet(rw http.ResponseWriter, r *http.Request) {
	h.logger.Info("Processing PUT ImageSet request")

	c, err := url.QueryUnescape( r.PathValue("category") )
	if err != nil {
		response.RespondWithMessage(rw, http.StatusBadRequest, "Cannot decode the category from url")
		return
	}
	id, err := url.QueryUnescape( r.PathValue("id") )
	if err != nil {
		response.RespondWithMessage(rw, http.StatusBadRequest, "Cannot decode the id from url")
		return
	}

	ofp := filepath.Join(c, id)

	i := &data.ImageSet{}
	err = utils.FromJSON(i, r.Body)
	if err != nil {
		response.RespondWithMessage(rw, http.StatusBadRequest, "Invalid data format")
		return
	}

	err = i.Validate()
	if err != nil {
		response.RespondWithMessage(rw, http.StatusBadRequest, "Invalid image set data")
		return
	}
	nfp := filepath.Join(i.Category, i.ID)

	err = h.store.RenameDirectory(ofp, nfp)
	if err != nil {
		if err == files.ErrNotFound {
			response.RespondWithMessage(rw, http.StatusNotFound, "Unable to find ImageSet")
			return
		}
		response.RespondWithMessage(rw, http.StatusInternalServerError, "Unable to update ImageSet")
		return
	}

	response.RespondWithMessage(rw, http.StatusOK, "ImageSet updated successfully")
}

// swagger:route DELETE /imageSets/{category}/{id} imageSets deleteImageSet
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
// 	200: message
//  400: message
//	403: message
//	404: message
// 	500: message
func (h *ImageSetsHandler) DeleteImageSet(rw http.ResponseWriter, r *http.Request) {
	h.logger.Info("Processing DELETE ImageSet request")

	c, err := url.QueryUnescape( r.PathValue("category") )
	if err != nil {
		response.RespondWithMessage(rw, http.StatusBadRequest, "Cannot decode the category from url")
		return
	}

	id, err := url.QueryUnescape( r.PathValue("id") )
	if err != nil {
		response.RespondWithMessage(rw, http.StatusBadRequest, "Cannot decode the id from url")
		return
	}

	fp := filepath.Join(c, id)

	err = h.store.DeleteFiles(fp)
	if err != nil {
		if err == files.ErrNotFound {
			response.RespondWithMessage(rw, http.StatusNotFound, "ImageSet doesn't exist")
			return
		}
		response.RespondWithMessage(rw, http.StatusInternalServerError, "Unable to remove ImageSet")
		return
	}

	err = h.store.DeleteDirectory(fp)
	if err != nil {
		if err == files.ErrDirNotEmpty {
			response.RespondWithMessage(rw, http.StatusForbidden, "ImageSet contains subdirectories")
			return
		}
		response.RespondWithMessage(rw, http.StatusInternalServerError, "Unable to remove ImageSet")
		return
	}

	response.RespondWithMessage(rw, http.StatusOK, "ImageSet removed successfully")
}

// swagger:route GET /imageCategories/{category} imageSets getCategory
//
// List subdirectories available in the category
//
// produces:
//	- application/json
//
// Responses:
// 	200: categoryContents
//  400: message
// 	403: message
//	404: message
// 	500: message
func (h *ImageSetsHandler) GetCategory(rw http.ResponseWriter, r *http.Request) {
	h.logger.Info("Processing GET ImageSet Category request")
	
	c := r.PathValue("category")
	if c == "" {
		response.RespondWithMessage(rw, http.StatusBadRequest, "Category is required")
		return
	}

	fp, err := url.QueryUnescape( r.PathValue("category") )
	if err != nil {
		response.RespondWithMessage(rw, http.StatusBadRequest, "Cannot decode the category from url")
		return
	}

	f, err := h.store.ListDirectories(fp)
	if err != nil {
		if err == files.ErrNotFound {
			response.RespondWithMessage(rw, http.StatusNotFound, "Category doesn't exist")
			return
		}
		if err == files.ErrNotDirectory {
			response.RespondWithMessage(rw, http.StatusForbidden, "Requested path is not a directory")
			return
		}
		response.RespondWithMessage(rw, http.StatusInternalServerError, "Unable to retrieve Category")
		return
	}

	is := &response.CategoryResponse{Directories: f}

	response.RespondWithJSON(rw, http.StatusOK, is)
}

// swagger:route POST /imageCategories/{category} imageSets postCategory
//
// Creates a requested directory path
//
// produces:
//	- application/json
//
// Responses:
// 	200: message
//  400: message
// 	403: message
// 	500: message
func (h *ImageSetsHandler) PostCategory(rw http.ResponseWriter, r *http.Request) {
	h.logger.Info("Processing POST ImageSet Category request")

	c := r.PathValue("category")
	if c == "" {
		response.RespondWithMessage(rw, http.StatusBadRequest, "Category is required")
		return
	}
	
	fp, err := url.QueryUnescape( r.PathValue("category") )
	if err != nil {
		response.RespondWithMessage(rw, http.StatusBadRequest, "Cannot decode category from url")
		return
	}

	err = h.store.CreateDirectory(fp)
	if err != nil {
		if err == files.ErrAlreadyExists {
			response.RespondWithMessage(rw, http.StatusForbidden, "Directory already exists")
			return
		}
		response.RespondWithMessage(rw, http.StatusInternalServerError, "Unable to create Category")
		return
	}

	response.RespondWithMessage(rw, http.StatusOK, "Category created successfully")
}

// swagger:route PUT /imageCategories/{category} imageSets putCategory
//
// Update existing Category
//
// produces:
//	- application/json
//
// Responses:
// 	200: message
//  400: message
// 	403: message
//	404: message
// 	500: message
func (h *ImageSetsHandler) PutCategory(rw http.ResponseWriter, r *http.Request) {
	h.logger.Info("Processing PUT ImageSet Category request")

	c := r.PathValue("category")
	if c == "" {
		response.RespondWithMessage(rw, http.StatusBadRequest, "Category is required")
		return
	}

	ofp, err := url.QueryUnescape( r.PathValue("category") )
	if err != nil {
		response.RespondWithMessage(rw, http.StatusBadRequest, "Cannot decode the category from url")
		return
	}

	// TODO: replace this with proper request/response structs
	i := &data.ImageSet{}
	err = utils.FromJSON(i, r.Body)
	if err != nil {
		response.RespondWithMessage(rw, http.StatusBadRequest, "Invalid data format")
		return
	}

	err = h.store.MoveDirectory(ofp, i.Category)
	if err != nil {
		if err == files.ErrNotFound {
			response.RespondWithMessage(rw, http.StatusNotFound, "Unable to find Category")
			return
		}
		if err == files.ErrAlreadyExists {
			response.RespondWithMessage(rw, http.StatusBadRequest, "Category already exists")
			return
		}
		if err == files.ErrDirContainsFiles {
			response.RespondWithMessage(rw, http.StatusForbidden, "Category contains illegal files")
			return
		}
		response.RespondWithMessage(rw, http.StatusInternalServerError, "Unable to update Category")
		return
	}

	response.RespondWithMessage(rw, http.StatusOK, "Category updated successfully")
}

// swagger:route DELETE /imageCategories/{category} imageSets deleteCategory
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

	c := r.PathValue("category")
	if c == "" {
		response.RespondWithMessage(rw, http.StatusBadRequest, "Category is required")
		return
	}

	fp, err := url.QueryUnescape( r.PathValue("category") )
	if err != nil {
		response.RespondWithMessage(rw, http.StatusBadRequest, "Cannot decode category from url")
		return
	}

	err = h.store.DeleteSubdirectories(fp)
	if err != nil {
		if err == files.ErrNotFound {
			response.RespondWithMessage(rw, http.StatusNotFound, "Category doesn't exist")
			return
		}
		response.RespondWithMessage(rw, http.StatusInternalServerError, "Unable to remove Category")
		return
	}

	err = h.store.DeleteDirectory(fp)
	if err != nil {
		if err == files.ErrDirNotEmpty {
			response.RespondWithMessage(rw, http.StatusForbidden, "Category contains files")
			return
		}
		response.RespondWithMessage(rw, http.StatusInternalServerError, "Unable to remove Category")
		return
	}

	response.RespondWithMessage(rw, http.StatusOK, "Category removed successfully")
}
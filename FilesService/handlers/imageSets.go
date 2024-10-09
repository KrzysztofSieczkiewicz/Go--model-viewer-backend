package handlers

import (
	"log"
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
curl -v -X GET http://localhost:9090/imageSets/random/1

POST IMAGESET:
curl -v -i -X POST http://localhost:9090/imageSets/random/1

PUT IMAGESET:
curl -v -i -X PUT http://localhost:9090/imageSets/random/1 -H "Content-Type: application/json" -d "{\"category\":\"random\",\"id\":\"4\"}"

DELETE IMAGESET:
curl -v -i -X DELETE http://localhost:9090/imageSets/random/1

GET CATEGORY:
curl -v -X GET http://localhost:9090/imageCategories/random

POST CATEGORY:
curl -v -X POST http://localhost:9090/imageCategories/random%2Ftest%2F1

PUT CATEGORY:
curl -v -X PUT http://localhost:9090/imageCategories/random%2Ftest%2F1 -H "Content-Type: application/json" -d "{\"FilePath\":\"random/test2\"}"

*/

// Handler for managing imageSets and categories
type ImageSetsHandler struct {
	baseUrl		string
	logger		*log.Logger
	store		files.Storage
	cache		caches.Cache
	signedUrl	signedurl.SignedUrl
}

func NewImageSets(baseUrl string, s files.Storage, l *log.Logger, c caches.Cache) *ImageSetsHandler {
	return &ImageSetsHandler{
		baseUrl: baseUrl,
		store: s, 
		logger: l,
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
// Returns ImageSet details and available images.
//
// produces:
//	- application/json
//
// Responses:
// 	200: imageSetJson
//  400: messageJson
// 	403: messageJson
//	404: messageJson
// 	500: messageJson
func (h *ImageSetsHandler) GetImageSet(rw http.ResponseWriter, r *http.Request) {
	c := r.PathValue("category")
	id := r.PathValue("id")
	if c == "" || id == "" {
		utils.RespondWithMessage(rw, http.StatusBadRequest, "Category and ID are required")
		return
	}

	fp := filepath.Join(c, id)

	f, err := h.store.ListFiles(fp)
	if err != nil {
		if err == files.ErrDirectoryNotFound {
			utils.RespondWithMessage(rw, http.StatusNotFound, "ImageSet doesn't exist")
			return
		}
		if err == files.ErrNotDirectory {
			utils.RespondWithMessage(rw, http.StatusForbidden, "Requested path is not a directory")
			return
		}
		utils.RespondWithMessage(rw, http.StatusInternalServerError, "Unable to retrieve ImageSet data")
		return
	}

	i := &data.Images{}
	err = i.DeconstructImageNames(f)
	if err != nil {
		utils.RespondWithMessage(rw, http.StatusInternalServerError, "Unable to deconstruct image names:")
		return
	}

	is := &response.ImageSetResponse{
		ID: id,
		Category: c,
		Images: *i,
	}

	utils.RespondWithJSON(rw, http.StatusOK, is)
}

// swagger:route POST /imageSets/{category}/{id} imageSets postImageSet
//
// Create a new image set.
//
// produces:
//	- application/json
//
// Responses:
// 	201: messageJson
//  400: messageJson
// 	403: messageJson
// 	500: messageJson
func (h *ImageSetsHandler) PostImageSet(rw http.ResponseWriter, r *http.Request) {
	c := r.PathValue("category")
	id := r.PathValue("id")
	if c == "" || id == "" {
		utils.RespondWithMessage(rw, http.StatusBadRequest, "Category and ID are required")
		return
	}

	fp := filepath.Join(c, id)

	err := h.store.MakeDirectory(fp)
	if err != nil {
		if err == files.ErrDirectoryAlreadyExists {
			utils.RespondWithMessage(rw, http.StatusForbidden, "ImageSet already exists")
			return
		}
		utils.RespondWithMessage(rw, http.StatusInternalServerError, "Failed to create ImageSet")
		return
	}

	utils.RespondWithMessage(rw, http.StatusCreated, "ImageSet created successfully")
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
// 	200: messageJson
//  400: messageJson
// 	403: messageJson
//	404: messageJson
// 	500: messageJson
func (h *ImageSetsHandler) PutImageSet(rw http.ResponseWriter, r *http.Request) {
	c := r.PathValue("category")
	id := r.PathValue("id")
	if c == "" || id == "" {
		utils.RespondWithMessage(rw, http.StatusBadRequest, "Category and ID are required")
		return
	}
	ofp := filepath.Join(c, id)

	i := &data.ImageSet{}
	err := utils.FromJSON(i, r.Body)
	if err != nil {
		utils.RespondWithMessage(rw, http.StatusBadRequest, "Invalid data format")
		return
	}
	nfp := filepath.Join(i.Category, i.ID)

	// Rename imageSet
	err = h.store.RenameDirectory(ofp, nfp)
	if err != nil {
		if err == files.ErrDirectoryNotFound {
			utils.RespondWithMessage(rw, http.StatusNotFound, "Unable to find ImageSet")
			return
		}
		utils.RespondWithMessage(rw, http.StatusInternalServerError, "Failed to update ImageSet")
		return
	}

	utils.RespondWithMessage(rw, http.StatusOK, "ImageSet updated successfully")
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
// 	200: messageJson
//  400: messageJson
//	403: messageJson
//	404: messageJson
// 	500: messageJson
func (h *ImageSetsHandler) DeleteImageSet(rw http.ResponseWriter, r *http.Request) {
	c := r.PathValue("category")
	id := r.PathValue("id")
	if c == "" || id == "" {
		utils.RespondWithMessage(rw, http.StatusBadRequest, "Category and ID are required")
		return
	}
	fp := filepath.Join(c, id)

	err := h.store.DeleteDirectory(fp)
	if err != nil {
		if err == files.ErrDirectoryNotFound {
			utils.RespondWithMessage(rw, http.StatusNotFound, "ImageSet doesn't exist")
			return
		}
		if err == files.ErrDirectorySubdirectoryFound {
			utils.RespondWithMessage(rw, http.StatusForbidden, "ImageSet contains subdirectories")
			return
		}
		utils.RespondWithMessage(rw, http.StatusInternalServerError, "Failed to remove ImageSet")
		return
	}

	utils.RespondWithMessage(rw, http.StatusOK, "ImageSet removed successfully")
}

// swagger:route GET /imageCategories/{category} imageSets getCategory
//
// List subdirectories available in the category
//
// produces:
//	- application/json
//
// Responses:
// 	200: categoryContentsJson
//  400: messageJson
// 	403: messageJson
//	404: messageJson
// 	500: messageJson
func (h *ImageSetsHandler) GetCategory(rw http.ResponseWriter, r *http.Request) {
	c := r.PathValue("category")
	if c == "" {
		utils.RespondWithMessage(rw, http.StatusBadRequest, "Category is required")
		return
	}

	fp, err := url.QueryUnescape( r.PathValue("category") )
	if err != nil {
		utils.RespondWithMessage(rw, http.StatusBadRequest, "Cannot decode the category")
		return
	}

	f, err := h.store.ListDirectories(fp)
	if err != nil {
		if err == files.ErrDirectoryNotFound {
			utils.RespondWithMessage(rw, http.StatusNotFound, "Category doesn't exist")
			return
		}
		if err == files.ErrNotDirectory {
			utils.RespondWithMessage(rw, http.StatusForbidden, "Requested path is not a directory")
			return
		}
		utils.RespondWithMessage(rw, http.StatusInternalServerError, "Unable to retrieve Category")
		return
	}

	is := &response.CategoryResponse{ImageSets: f}

	utils.RespondWithJSON(rw, http.StatusOK, is)
}

// swagger:route POST /imageCategories/{category} imageSets postCategory
//
// Creates a requested directory path
//
// produces:
//	- application/json
//
// Responses:
// 	200: messageJson
//  400: messageJson
// 	403: messageJson
// 	500: messageJson
func (h *ImageSetsHandler) PostCategory(rw http.ResponseWriter, r *http.Request) {
	c := r.PathValue("category")
	if c == "" {
		utils.RespondWithMessage(rw, http.StatusBadRequest, "Category is required")
		return
	}
	
	fp, err := url.QueryUnescape( r.PathValue("category") )
	if err != nil {
		utils.RespondWithMessage(rw, http.StatusBadRequest, "Cannot decode category")
		return
	}

	err = h.store.MakeDirectory(fp)
	if err != nil {
		if err == files.ErrDirectoryAlreadyExists {
			utils.RespondWithMessage(rw, http.StatusForbidden, "Directory already exists")
			return
		}
		utils.RespondWithMessage(rw, http.StatusInternalServerError, "Failed to create Category")
		return
	}

	utils.RespondWithMessage(rw, http.StatusOK, "Category created successfully")
}

// swagger:route PUT /imageCategories/{category} imageSets putCategory
//
// Update existing Category.
//
// produces:
//	- application/json
//
// Responses:
// 	200: messageJson
//  400: messageJson
// 	403: messageJson
//	404: messageJson
// 	500: messageJson
func (h *ImageSetsHandler) PutCategory(rw http.ResponseWriter, r *http.Request) {
	c := r.PathValue("category")
	if c == "" {
		utils.RespondWithMessage(rw, http.StatusBadRequest, "Category is required")
		return
	}

	ofp, err := url.QueryUnescape( r.PathValue("category") )
	if err != nil {
		utils.RespondWithMessage(rw, http.StatusBadRequest, "Cannot decode the category")
		return
	}

	// TODO: replace this with proper request/response structs
	i := &struct {
		FilePath string
	}{
		FilePath: "",
	}
	err = utils.FromJSON(i, r.Body)
	if err != nil {
		utils.RespondWithMessage(rw, http.StatusBadRequest, "Invalid data format")
		return
	}

	err = h.store.MoveDirectory(ofp, i.FilePath)
	if err != nil {
		if err == files.ErrDirectoryNotFound {
			utils.RespondWithMessage(rw, http.StatusNotFound, "Unable to find Category")
			return
		}
		if err == files.ErrDirectoryAlreadyExists {
			utils.RespondWithMessage(rw, http.StatusBadRequest, "Category already exists")
			return
		}
		if err == files.ErrDirectoryNonDirectoryFound {
			utils.RespondWithMessage(rw, http.StatusForbidden, "Category contains illegal files")
			return
		}
		utils.RespondWithMessage(rw, http.StatusInternalServerError, "Failed to update Category")
		return
	}

	utils.RespondWithMessage(rw, http.StatusOK, "Category updated successfully")
}

// swagger:route DELETE /imageCategories/{category} imageSets deleteImageSet
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
// 	200: messageJson
//  400: messageJson
//	403: messageJson
//	404: messageJson
// 	500: messageJson
func (h *ImageSetsHandler) DeleteCategory(rw http.ResponseWriter, r *http.Request) {
	c := r.PathValue("category")
	if c == "" {
		utils.RespondWithMessage(rw, http.StatusBadRequest, "Category is required")
		return
	}

	fp, err := url.QueryUnescape( r.PathValue("category") )
	if err != nil {
		utils.RespondWithMessage(rw, http.StatusBadRequest, "Cannot decode category")
		return
	}

	// TODO: Add DeleteFiles and DeleteDirectories functions to clear the file beforehand
	// Limit delete directory to not empty dirs

	err = h.store.DeleteDirectory(fp)
	if err != nil {
		if err == files.ErrDirectoryNotFound {
			utils.RespondWithMessage(rw, http.StatusNotFound, "Category doesn't exist")
			return
		}
		if err == files.ErrDirectorySubdirectoryFound {
			utils.RespondWithMessage(rw, http.StatusForbidden, "Category contains subdirectories")
			return
		}
		utils.RespondWithMessage(rw, http.StatusInternalServerError, "Failed to remove ImageSet")
		return
	}

	utils.RespondWithMessage(rw, http.StatusOK, "Category removed successfully")
}
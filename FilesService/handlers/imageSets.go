package handlers

import (
	"log"
	"net/http"
	"path/filepath"
	"time"

	"github.com/KrzysztofSieczkiewicz/go--model-viewer-backend/FilesService/caches"
	"github.com/KrzysztofSieczkiewicz/go--model-viewer-backend/FilesService/data"
	"github.com/KrzysztofSieczkiewicz/go--model-viewer-backend/FilesService/files"
	"github.com/KrzysztofSieczkiewicz/go--model-viewer-backend/FilesService/response"
	"github.com/KrzysztofSieczkiewicz/go--model-viewer-backend/FilesService/signedurl"
	"github.com/KrzysztofSieczkiewicz/go--model-viewer-backend/FilesService/utils"
)

// Example curls (OUTDATED):
// Get file: curl -v localhost:9090/files/random/1/thumbnail.png
// Post file: curl -v -X POST -H "Content-Type: image/png" --data-binary @FilesService/thumbnail.png localhost:9090/files/random/1/thumbnail.png
// Get url: curl -v localhost:9090/url/random/1/thumbnail.png

// Handler for reading and writing files to provided storage
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


// swagger:route GET /{category}/{id} imageSets getImageUrl
//
// Returns a signed url to requested resource.
//
// consumes:
//	- application/json
//
// produces:
//	- application/json
//
// Responses:
// 	200: imageUrlJson
//	404: messageJson
//	500: messageJson
func (h *ImageSetsHandler) GetImageUrl(rw http.ResponseWriter, r *http.Request) {
	c := r.PathValue("category")
	id := r.PathValue("id")
	if c == "" || id == "" {
		utils.RespondWithMessage(rw, http.StatusBadRequest, "Category and ID are required")
		return
	}

	i := &data.Image{}
	json := r.FormValue("json")
	if json == "" {
		utils.RespondWithMessage(rw, http.StatusBadRequest, "Missing JSON data")
		return
	}

	err := utils.FromJSONString(i, json)
	if err != nil {
		utils.RespondWithMessage(rw, http.StatusBadRequest, "Invalid data format")
		return
	}

	fn := i.ConstructImageName()
	fp := filepath.Join(c, id, fn)

	err = h.store.CheckFile(fp)
	if err != nil {
		if err == files.ErrFileNotFound {
			http.Error(rw, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(rw, err.Error(), http.StatusInternalServerError)
	}

	tmpId := caches.GenerateUUID()
	h.cache.Set(tmpId, fp)
	url := h.signedUrl.GenerateSignedUrl(tmpId)

    response := response.ImageUrlResponse{
        Filename: fn,
        URL:      url,
    }

	utils.RespondWithJSON(rw, http.StatusOK, response)
}

// swagger:route POST /{category}/{id} imageSets postImage
//
// Adds an image to the existing set.
//
// consumes:
//  - multipart/form-data
//
// produces:
//	- application/json
//
// Responses:
// 	201: messageJson
//  400: messageJson
// 	403: messageJson
// 	500: messageJson
func (h *ImageSetsHandler) PostImage(rw http.ResponseWriter, r *http.Request) {
	c := r.PathValue("category")
	id := r.PathValue("id")
	if c == "" || id == "" {
		utils.RespondWithMessage(rw, http.StatusBadRequest, "Category and ID are required")
		return
	}

	i := &data.Image{}
	json := r.FormValue("json")
	if json == "" {
		utils.RespondWithMessage(rw, http.StatusBadRequest, "Missing JSON part of the request")
		return
	}

	err := utils.FromJSONString(i, json)
	if err != nil {
		utils.RespondWithMessage(rw, http.StatusBadRequest, "Invalid data format")
		return
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		utils.RespondWithMessage(rw, http.StatusBadRequest, "Error reading file from request")
		return
	}
	defer file.Close()
	
	fn := i.ConstructImageName()
	fp := filepath.Join(c, id, fn)

	err = h.store.Write(fp, file)
	if err != nil {
		if err == files.ErrFileAlreadyExists {
			utils.RespondWithMessage(rw, http.StatusForbidden, "Image already exists")
			return
		}
		if err == files.ErrDirectoryNotFound {
			utils.RespondWithMessage(rw, http.StatusBadRequest, "Image set doesn't exist")
			return
		}
		utils.RespondWithMessage(rw, http.StatusInternalServerError, "Failed to create the file")
		return
	}

	utils.RespondWithMessage(rw, http.StatusCreated, "Image uploaded sucessfully")
}

// swagger:route PUT /{category}/{id} imageSets putImage
//
// Updates an image in the image set.
//
// consumes:
//  - multipart/form-data
//
// produces:
//	- application/json
//
// Responses:
// 	200: messageJson
// 	404: messageJson
// 	500: messageJson
func (h *ImageSetsHandler) PutImage(rw http.ResponseWriter, r *http.Request) {
	c := r.PathValue("category")
	id := r.PathValue("id")
	if c == "" || id == "" {
		utils.RespondWithMessage(rw, http.StatusBadRequest, "Category and ID are required")
		return
	}

	i := &data.Image{}
	json := r.FormValue("json")
	if json == "" {
		utils.RespondWithMessage(rw, http.StatusBadRequest, "Missing JSON part of the request")
		return
	}

	err := utils.FromJSONString(i, json)
	if err != nil {
		utils.RespondWithMessage(rw, http.StatusBadRequest, "Invalid data format")
		return
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		utils.RespondWithMessage(rw, http.StatusBadRequest, "Error reading file from request")
		return
	}
	defer file.Close()
	
	fn := i.ConstructImageName()
	fp := filepath.Join(c, id, fn)

	err = h.store.Overwrite(fp, file)
	if err != nil {
		if err == files.ErrFileNotFound {
			utils.RespondWithMessage(rw, http.StatusNotFound, "File does not exist")
			return
		}
		utils.RespondWithMessage(rw, http.StatusInternalServerError, "Failed to update the file")
		return
	}

	utils.RespondWithMessage(rw, http.StatusOK, "Image uploaded sucessfully")
}

// swagger:route DELETE /{category}/{id} imageSets deleteImage
//
// Removes image from the image set.
//
// consumes:
//  - application/json
//
// produces:
//	- application/json
//
// Responses:
// 	204: empty
//	404: messageJson
//	500: messageJson
func (f *Files) DeleteImage(rw http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	c := r.PathValue("category")
	if c == "" || id == "" {
		utils.RespondWithMessage(rw, http.StatusBadRequest, "Category and ID are required")
		return
	}

	i := &data.Image{}
	err := utils.FromJSON(i, r.Body)
	if err != nil {
		utils.RespondWithMessage(rw, http.StatusBadRequest, "Invalid data format")
		return
	}

	fn := i.ConstructImageName()
	fp := filepath.Join(c, id, fn)

	err = f.store.Delete(fp)
	if err != nil {
		if err == files.ErrFileNotFound {
			utils.RespondWithMessage(rw, http.StatusNotFound, "Image was not found")
			return
		}
		utils.RespondWithMessage(rw, http.StatusBadRequest, "Failed to delete the image")
		return
	}

	rw.WriteHeader(http.StatusNoContent)
}

// swagger:route GET /{category}/{id} imageSets getImageSet
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

// swagger:route POST /{category}/{id} imageSets postImageSet
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
	}

	utils.RespondWithMessage(rw, http.StatusCreated, "ImageSet created successfully")
}

// swagger:route PUT /{category}/{id} imageSets putImageSet
//
// Update existing imageset id or category
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

	err = h.store.RenameDirectory(ofp, nfp)
	if err != nil {
		if err == files.ErrDirectoryNotFound {
			utils.RespondWithMessage(rw, http.StatusNotFound, "Unable to find ImageSet")
			return
		}
		if err == files.ErrDirectoryAlreadyExists {
			utils.RespondWithMessage(rw, http.StatusForbidden, "ImageSet already exists")
			return
		}
		utils.RespondWithMessage(rw, http.StatusInternalServerError, "Failed to update ImageSet")
		return
	}

	utils.RespondWithMessage(rw, http.StatusOK, "ImageSet updated successfully")
}

// swagger:route DELETE /{category}/{id} imageSets deleteImageSet
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
		}
		utils.RespondWithMessage(rw, http.StatusInternalServerError, "Failed to remove ImageSet")
		return
	}

	utils.RespondWithMessage(rw, http.StatusOK, "ImageSet removed successfully")
}

// swagger:route GET /{category} imageSets getCategory
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

	f, err := h.store.ListDirectories(c)
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

	is := &response.CategoryResponse{Contents: f}

	utils.RespondWithJSON(rw, http.StatusOK, is)
}


// swagger:route GET /{id}&{expires}&{signature} imageSets getImage
//
// Returns an image from imageset. Handles signed URLs
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
func (f *Files) GetImage(rw http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	exp := r.URL.Query().Get("expires")
	sign := r.URL.Query().Get("signature")

	err := f.signedUrl.ValidateSignedUrl(id, exp, sign)
	if err != nil {
		if err == signedurl.ErrUrlExpired {
			utils.RespondWithMessage(rw, http.StatusForbidden, "URL has expired")
			return
		}
		if err == signedurl.ErrInvalidSignature {
			utils.RespondWithMessage(rw, http.StatusBadRequest, "Invalid signature")
			return
		}
		if err == signedurl.ErrInvalidTimestamp {
			utils.RespondWithMessage(rw, http.StatusBadRequest, "Invalid timestamp")
			return
		}
		utils.RespondWithMessage(rw, http.StatusBadRequest, "Invalid request")
		return
	}

	fp, err := f.cache.Get(id)
	if err != nil {
		utils.RespondWithMessage(rw, http.StatusInternalServerError, "Request doesn't match cache")
		return
	}

	err = f.store.Read(fp, rw)
	if err != nil {
		utils.RespondWithMessage(rw, http.StatusInternalServerError, "Failed to retrieve requested file")
		return
	}

	rw.Header().Set("Content-Type", "application/octet-stream")
	rw.WriteHeader(http.StatusOK)
}
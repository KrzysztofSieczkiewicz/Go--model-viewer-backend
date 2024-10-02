package handlers

import (
	"log"
	"net/http"
	"path/filepath"
	"time"

	"github.com/KrzysztofSieczkiewicz/go--model-viewer-backend/FilesService/caches"
	"github.com/KrzysztofSieczkiewicz/go--model-viewer-backend/FilesService/data"
	"github.com/KrzysztofSieczkiewicz/go--model-viewer-backend/FilesService/files"
	"github.com/KrzysztofSieczkiewicz/go--model-viewer-backend/FilesService/signedurl"
	"github.com/KrzysztofSieczkiewicz/go--model-viewer-backend/FilesService/utils"
)

/*
TODO: required functionalities:
1. Post Image
2. Post multiple images within set
3. Update image
4. Update multiple images within set
5. Delete image
6. Delete multiple images within set
7. Delete image set
8. Get image from set (url)
9. Get entire set ([]url)
10. Get resource from URL
11. List available under given cathegory

To handle these requests data must be moved from the url to the body - preferably as json
*/

// Example curls:
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
// 	201: message
//  400: message
// 	403: message
// 	500: message
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
	
	fn := i.GetImageName()
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
// 	200: message
// 	404: message
// 	500: message
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
	
	fn := i.GetImageName()
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
//	404: error
//	500: error
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

	fn := i.GetImageName()
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
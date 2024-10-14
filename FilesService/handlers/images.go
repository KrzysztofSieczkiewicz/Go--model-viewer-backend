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
GET IMAGE URL:
curl -v -X GET http://localhost:9090/images/random%2Ftest%2F/1 -H "Content-Type: application/json" -d "{\"type\":\"albedo\",\"resolution\":\"2048x2048\",\"extension\":\"png\"}"

POST IMAGE:
curl -v -i -X POST http://localhost:9090/images/random%2Ftest%2F/1 -H "Content-Type: multipart/form-data" -F "metadata={\"type\":\"albedo\",\"resolution\":\"2048x2048\",\"extension\":\"png\"};type=application/json" -F "file=@FilesService/thumbnail.png;type=image/png"

PUT IMAGE:
curl -v -i -X PUT http://localhost:9090/images/random%2Ftest%2F/1 -H "Content-Type: multipart/form-data" -F "metadata={\"type\":\"albedo\",\"resolution\":\"2048x2048\",\"extension\":\"png\"};type=application/json" -F "file=@FilesService/thumbnail.png;type=image/png"

DELETE IMAGE:
curl -v -i -X DELETE http://localhost:9090/images/random%2Ftest%2F/1 -H "Content-Type: application/json" -d "{\"type\":\"albedo\",\"resolution\":\"2048x2048\",\"extension\":\"png\"}"
*/

// Handler for reading and writing images into the imageSets in the storage
type ImagesHandler struct {
	baseUrl		string
	logger		*slog.Logger
	store		files.Storage
	cache		caches.Cache
	signedUrl	signedurl.SignedUrl
}

func NewImages(baseUrl string, s files.Storage, l *slog.Logger, c caches.Cache) *ImagesHandler {
	return &ImagesHandler{
		baseUrl: baseUrl,
		store: s, 
		logger: l,
		cache: c,
		signedUrl: *signedurl.NewSignedUrl(
			"Secret key my boy",
			baseUrl + "/images",
			time.Duration(5 * int(time.Minute)),
		),
	}
}

// swagger:route GET /images/{category}/{id} images getImageUrl
//
// Return a signed url to requested image
//
// consumes:
//	- application/json
//
// produces:
//	- application/json
//
// Responses:
// 	200: fileUrl
//	404: message
//	500: message
func (h *ImagesHandler) GetUrl(rw http.ResponseWriter, r *http.Request) {
	h.logger.Info("Processing GET Image URL request")

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

	i := &data.Image{}
	err = utils.FromJSON(i, r.Body)
	if err != nil {
		response.RespondWithMessage(rw, http.StatusBadRequest, "Invalid JSON data")
		return
	}

	fn := i.ConstructImageName()
	fp := filepath.Join(c, id, fn)

	err = h.store.IfExists(fp)
	if err != nil {
		if err == files.ErrNotFound {
			http.Error(rw, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(rw, err.Error(), http.StatusInternalServerError)
	}

	tmpId := caches.GenerateUUID()
	h.cache.Set(tmpId, fp)
	url := h.signedUrl.GenerateSignedUrl(tmpId)

    urlResponse := response.FileUrlResponse{
        Filename: fn,
        URL:      url,
    }

	response.RespondWithJSON(rw, http.StatusOK, urlResponse)
}

// swagger:route GET /{id}&{expires}&{signature} images getImage
//
// Return an image from imageset. Can only be accessed by signed URLs
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
func (h *ImagesHandler) GetImage(rw http.ResponseWriter, r *http.Request) {
	h.logger.Info("Processing GET Image request")

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

	err = h.store.ReadFile(fp, rw)
	if err != nil {
		response.RespondWithMessage(rw, http.StatusInternalServerError, "Failed to retrieve requested file")
		return
	}

	rw.Header().Set("Content-Type", "application/octet-stream")
	rw.WriteHeader(http.StatusOK)
}

// swagger:route POST /images/{category}/{id} images postImage
//
// Add an image to the existing set
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
func (h *ImagesHandler) PostImage(rw http.ResponseWriter, r *http.Request) {
	h.logger.Info("Processing POST Image request")

	c := r.PathValue("category")
	id := r.PathValue("id")
	if c == "" || id == "" {
		response.RespondWithMessage(rw, http.StatusBadRequest, "Category and ID are required")
		return
	}

	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		response.RespondWithMessage(rw, http.StatusBadRequest, "Unable to parse form data")
		return
	}

	i := &data.Image{}
	json := r.FormValue("metadata")
	if json == "" {
		response.RespondWithMessage(rw, http.StatusBadRequest, "Missing JSON part of the request")
		return
	}

	err = utils.FromJSONString(i, json)
	if err != nil {
		response.RespondWithMessage(rw, http.StatusBadRequest, "Invalid data format")
		return
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		response.RespondWithMessage(rw, http.StatusBadRequest, "Error reading file from request")
		return
	}
	defer file.Close()
	
	fn := i.ConstructImageName()
	fp := filepath.Join(c, id, fn)

	err = h.store.WriteFile(fp, file)
	if err != nil {
		if err == files.ErrFileAlreadyExists {
			response.RespondWithMessage(rw, http.StatusForbidden, "Image already exists")
			return
		}
		if err == files.ErrNotFound {
			response.RespondWithMessage(rw, http.StatusBadRequest, "ImageSet doesn't exist")
			return
		}
		response.RespondWithMessage(rw, http.StatusInternalServerError, "Failed to create the file")
		return
	}

	response.RespondWithMessage(rw, http.StatusCreated, "Image uploaded sucessfully")
}

// swagger:route PUT /images/{category}/{id} images putImage
//
// Update an image in the image set
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
func (h *ImagesHandler) PutImage(rw http.ResponseWriter, r *http.Request) {
	h.logger.Info("Processing PUT Image request")

	c := r.PathValue("category")
	id := r.PathValue("id")
	if c == "" || id == "" {
		response.RespondWithMessage(rw, http.StatusBadRequest, "Category and ID are required")
		return
	}

	i := &data.Image{}
	json := r.FormValue("metadata")
	if json == "" {
		response.RespondWithMessage(rw, http.StatusBadRequest, "Missing JSON part of the request")
		return
	}

	err := utils.FromJSONString(i, json)
	if err != nil {
		response.RespondWithMessage(rw, http.StatusBadRequest, "Invalid data format")
		return
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		response.RespondWithMessage(rw, http.StatusBadRequest, "Error reading file from request")
		return
	}
	defer file.Close()
	
	fn := i.ConstructImageName()
	fp := filepath.Join(c, id, fn)

	err = h.store.OverwriteFile(fp, file)
	if err != nil {
		if err == files.ErrNotFound {
			response.RespondWithMessage(rw, http.StatusNotFound, "File does not exist")
			return
		}
		response.RespondWithMessage(rw, http.StatusInternalServerError, "Failed to update the file")
		return
	}

	response.RespondWithMessage(rw, http.StatusOK, "Image updated sucessfully")
}

// swagger:route DELETE /images/{category}/{id} images deleteImage
//
// Remove image from the image set
//
// consumes:
//  - application/json
//
// produces:
//	- application/json
//
// Responses:
// 	204: empty
//	404: message
//	500: message
func (h *ImagesHandler) DeleteImage(rw http.ResponseWriter, r *http.Request) {
	h.logger.Info("Processing DELETE Image request")

	id := r.PathValue("id")
	c := r.PathValue("category")
	if c == "" || id == "" {
		response.RespondWithMessage(rw, http.StatusBadRequest, "Category and ID are required")
		return
	}

	i := &data.Image{}
	err := utils.FromJSON(i, r.Body)
	if err != nil {
		response.RespondWithMessage(rw, http.StatusBadRequest, "Invalid data format")
		return
	}

	fn := i.ConstructImageName()
	fp := filepath.Join(c, id, fn)

	err = h.store.DeleteFile(fp)
	if err != nil {
		if err == files.ErrNotFound {
			response.RespondWithMessage(rw, http.StatusNotFound, "Image was not found")
			return
		}
		response.RespondWithMessage(rw, http.StatusBadRequest, "Failed to delete the image")
		return
	}

	rw.WriteHeader(http.StatusNoContent)
}
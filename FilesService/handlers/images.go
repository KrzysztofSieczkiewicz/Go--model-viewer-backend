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

/*
Example curls:
GET IMAGE URL:
curl -v -X GET http://localhost:9090/images/url -H "Content-Type: application/json" -d "{\"category\":\"random/test\",\"id\":\"1\",\"type\":\"albedo\",\"resolution\":\"2048x2048\",\"extension\":\"png\"}"

POST IMAGE:
curl -v -i -X POST http://localhost:9090/images -H "Content-Type: multipart/form-data" -F "metadata={\"category\":\"random/test\",\"id\":\"1\",\"type\":\"albedo\",\"resolution\":\"2048x2048\",\"extension\":\"png\"};type=application/json" -F "file=@FilesService/thumbnail.png;type=image/png"

PUT IMAGE:
curl -v -i -X PUT http://localhost:9090/images -H "Content-Type: multipart/form-data" -F "metadata={\"category\":\"random/test\",\"id\":\"1\",\"type\":\"albedo\",\"resolution\":\"2048x2048\",\"extension\":\"png\"};type=application/json" -F "file=@FilesService/thumbnail.png;type=image/png"

DELETE IMAGE:
curl -v -i -X DELETE http://localhost:9090/images -H "Content-Type: application/json" -d "{\"category\":\"random/test\",\"id\":\"1\",\"type\":\"albedo\",\"resolution\":\"2048x2048\",\"extension\":\"png\"}"
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

// swagger:route GET /images images getImageUrl
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
//  400: message
//	404: message
//	500: message
func (h *ImagesHandler) GetImageUrl(rw http.ResponseWriter, r *http.Request) {
	h.logger.Info("Processing GET Image URL request")

	image := &models.Image{}
	err := utils.FromJSON(image, r.Body)
	if err != nil {
		response.RespondWithMessage(rw, http.StatusBadRequest, response.MessageInvalidJsonFormat)
		return
	}

	err = image.Validate()
	if err != nil {
		response.RespondWithMessage(rw, http.StatusBadRequest, response.MessageInvalidData)
		return
	}

	err = h.store.CheckAsset(image)
	if err != nil {
		if err == files.ErrNotFound {
			response.RespondWithMessage(rw, http.StatusNotFound, response.MessageNotFound)
			return
		}
		response.RespondWithMessage(rw, http.StatusInternalServerError, response.MessageFailedRead)
		return
	}

	tmpId := caches.GenerateUUID()
	h.cache.Set(tmpId, image)
	url := h.signedUrl.GenerateSignedUrl(tmpId)

    urlResponse := models.FileUrlResponse{
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
			response.RespondWithMessage(rw, http.StatusForbidden, response.MessageExpiredUrl)
			return
		}
		if err == signedurl.ErrInvalidSignature {
			response.RespondWithMessage(rw, http.StatusBadRequest, response.MessageInvalidSignature)
			return
		}
		if err == signedurl.ErrInvalidTimestamp {
			response.RespondWithMessage(rw, http.StatusBadRequest, response.MessageInvalidTimestamp)
			return
		}
		response.RespondWithMessage(rw, http.StatusBadRequest, response.MessageInvalidUrl)
		return
	}

	fp, err := h.cache.Get(id)
	if err != nil {
		response.RespondWithMessage(rw, http.StatusInternalServerError, response.MessageFailedCacheGet)
		return
	}

	rw.Header().Set("Content-Type", "application/octet-stream")

	err = h.store.GetAsset(fp, rw)
	if err != nil {
		response.RespondWithMessage(rw, http.StatusInternalServerError, response.MessageFailedRead)
		return
	}

	rw.WriteHeader(http.StatusOK)
}

// swagger:route POST /images images postImage
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
//	404: message
// 	500: message
func (h *ImagesHandler) PostImage(rw http.ResponseWriter, r *http.Request) {
	h.logger.Info("Processing POST Image request")

	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		response.RespondWithMessage(rw, http.StatusBadRequest, response.MessageFailedDataParsing)
		return
	}

	image := &models.Image{}
	json := r.FormValue("metadata")
	if json == "" {
		response.RespondWithMessage(rw, http.StatusBadRequest, response.MessageInvalidMultipartJson)
		return
	}

	err = utils.FromJSONString(image, json)
	if err != nil {
		response.RespondWithMessage(rw, http.StatusBadRequest, response.MessageInvalidJsonFormat)
		return
	}

	err = image.Validate()
	if err != nil {
		response.RespondWithMessage(rw, http.StatusBadRequest, response.MessageInvalidData)
		return
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		response.RespondWithMessage(rw, http.StatusBadRequest, response.MessageInvalidMultipartFile)
		return
	}
	defer file.Close()

	err = h.store.CreateAsset(image, file)
	if err != nil {
		if err == files.ErrAlreadyExists {
			response.RespondWithMessage(rw, http.StatusBadRequest, response.MessageAlreadyExists)
			return
		}
		if err == files.ErrNotFound {
			response.RespondWithMessage(rw, http.StatusNotFound, response.MessageNotFound)
			return
		}
		response.RespondWithMessage(rw, http.StatusInternalServerError, response.MessageFailedCreate)
		return
	}

	response.RespondWithMessage(rw, http.StatusCreated, response.MessageUploadSuccessful)
}

// swagger:route PUT /images/update images putImage
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
//  400: message
// 	404: message
// 	500: message
func (h *ImagesHandler) PutImage(rw http.ResponseWriter, r *http.Request) {
	h.logger.Info("Processing PUT Image request")

	image := &models.Image{}
	json := r.FormValue("metadata")
	if json == "" {
		response.RespondWithMessage(rw, http.StatusBadRequest, response.MessageInvalidMultipartJson)
		return
	}

	err := utils.FromJSONString(image, json)
	if err != nil {
		response.RespondWithMessage(rw, http.StatusBadRequest, response.MessageInvalidJsonFormat)
		return
	}
	
	err = image.Validate()
	if err != nil {
		response.RespondWithMessage(rw, http.StatusBadRequest, response.MessageInvalidData)
		return
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		response.RespondWithMessage(rw, http.StatusBadRequest, response.MessageInvalidMultipartFile)
		return
	}
	defer file.Close()

	err = h.store.OverwriteAsset(image, file)
	if err != nil {
		if err == files.ErrNotFound {
			response.RespondWithMessage(rw, http.StatusNotFound, response.MessageNotFound)
			return
		}
		response.RespondWithMessage(rw, http.StatusInternalServerError, response.MessageFailedUpdate)
		return
	}

	response.RespondWithMessage(rw, http.StatusOK, response.MessageUpdateSuccessful)
}

// swagger:route PUT /images/update images putImageData
//
// Updates the image data without changing the file contents
//
// consumes:
//  - application/json
//
// produces:
//	- application/json
//
// Responses:
// 	200: message
//  400: message
// 	404: message
// 	500: message
func (h *ImagesHandler) PutImageData(rw http.ResponseWriter, r *http.Request) {
	h.logger.Info("Processing PUT Image request")

	request := &models.PutRequest[models.Image]{}
	err := utils.FromJSON(request, r.Body)
	if err != nil {
		response.RespondWithMessage(rw, http.StatusBadRequest, response.MessageInvalidJsonFormat)
		return
	}

	err = request.Existing.Validate()
	if err != nil {
		response.RespondWithMessage(rw, http.StatusBadRequest, response.MessageInvalidData)
		return
	}

	err = request.New.Validate()
	if err != nil {
		response.RespondWithMessage(rw, http.StatusBadRequest, response.MessageInvalidData)
		return
	}

	err = h.store.UpdateAsset(&request.Existing, &request.New)
	if err != nil {
		if err == files.ErrNotFound {
			response.RespondWithMessage(rw, http.StatusNotFound, response.MessageNotFound)
			return
		}
		response.RespondWithMessage(rw, http.StatusInternalServerError, response.MessageFailedUpdate)
		return
	}

	response.RespondWithMessage(rw, http.StatusOK, response.MessageUpdateSuccessful)
}

// swagger:route DELETE /images images deleteImage
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
//  400: message
//	404: message
//	500: message
func (h *ImagesHandler) DeleteImage(rw http.ResponseWriter, r *http.Request) {
	h.logger.Info("Processing DELETE Image request")

	image := &models.Image{}
	err := utils.FromJSON(image, r.Body)
	if err != nil {
		response.RespondWithMessage(rw, http.StatusBadRequest, response.MessageInvalidJsonFormat)
		return
	}

	err = image.Validate()
	if err != nil {
		response.RespondWithMessage(rw, http.StatusBadRequest, response.MessageInvalidData)
		return
	}

	err = h.store.DeleteAsset(image)
	if err != nil {
		if err == files.ErrNotFound {
			response.RespondWithMessage(rw, http.StatusNotFound, response.MessageNotFound)
			return
		}
		response.RespondWithMessage(rw, http.StatusBadRequest, response.MessageFailedDelete)
		return
	}

	rw.WriteHeader(http.StatusNoContent)
}
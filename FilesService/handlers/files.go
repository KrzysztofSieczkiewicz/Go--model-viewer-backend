package handlers

import (
	"log"
	"net/http"
	"path/filepath"
	"time"

	"github.com/KrzysztofSieczkiewicz/go--model-viewer-backend/FilesService/caches"
	"github.com/KrzysztofSieczkiewicz/go--model-viewer-backend/FilesService/files"
	"github.com/KrzysztofSieczkiewicz/go--model-viewer-backend/FilesService/signedurl"
)

// Example curls:
// Get file: curl -v localhost:9090/files/random/1/thumbnail.png
// Post file: curl -v -X POST -H "Content-Type: image/png" --data-binary @FilesService/thumbnail.png localhost:9090/files/random/1/thumbnail.png
// Get url: curl -v localhost:9090/url/random/1/thumbnail.png

// Handler for reading and writing files to provided storage
type Files struct {
	baseUrl		string
	logger		*log.Logger
	store		files.Storage
	cache		caches.Cache
	signedUrl	signedurl.SignedUrl
}

func NewFiles(baseUrl string, s files.Storage, l *log.Logger, c caches.Cache) *Files {
	return &Files{
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

// swagger:route POST /{category}/{id}/{filename} files postFile
//
// Adds a file to the filesystem. Creates necessary folders. 
// Returns an error when a file already exists
//
// Responses:
// 	201: empty
// 	403: error
// 	500: error
func (f *Files) PostFile(rw http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	c := r.PathValue("category")
	fn := r.PathValue("filename")
	fp := filepath.Join(c, id, fn)

	err := f.store.Write(fp, r.Body)
	if err != nil {
		if err == files.ErrFileAlreadyExists {
			http.Error(rw, err.Error(), http.StatusForbidden)
			return
		}
		http.Error(rw, "Failed to create the file: \n" + err.Error(), http.StatusInternalServerError)
		return
	}

	rw.WriteHeader(http.StatusCreated)
}

// swagger:route PUT /{category}/{id}/{filename} files putFile
//
// Updates a file in the filesystem. Returns an error on file not found
//
// Responses:
// 	200: empty
// 	404: error
// 	500: error
func (f *Files) PutFile(rw http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	c := r.PathValue("category")
	fn := r.PathValue("filename")

	fp := filepath.Join(c, id, fn)

	err := f.store.Overwrite(fp, r.Body)
	if err != nil {
		if err == files.ErrFileNotFound {
			http.Error(rw, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(rw, "Failed to update the file: \n" + err.Error(), http.StatusInternalServerError)
		return
	}

	rw.WriteHeader(http.StatusOK)
}

// swagger:route GET /{category}/{id}/{filename} files getFileUrl
//
// Returns a signed url to requested resource. Url is timed
//
// Responses:
// 	200: empty
//	404: error
//	500: error
func (f *Files) GetFileUrl(rw http.ResponseWriter, r *http.Request) {
	c := r.PathValue("category")
	id := r.PathValue("id")
	fn := r.PathValue("filename")

	fp := filepath.Join(c, id, fn)

	// verify if file exists
	err := f.store.CheckFile(fp)
	if err != nil {
		if err == files.ErrFileNotFound {
			http.Error(rw, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(rw, err.Error(), http.StatusInternalServerError)
	}

	// generate uid and cache it with corresponding filepath
	tmpId := caches.GenerateUUID()
	f.cache.Set(tmpId, fp)

	// generate signedurl query based on uid
	url := f.signedUrl.GenerateSignedUrl(tmpId)

	rw.Write([]byte(url))

	rw.WriteHeader(http.StatusOK)
}


// swagger:route GET /{id}&{expires}&{signature} files getFile
//
// Returns a file. Handles signed URLs created with GetFileUrl function
//
// Responses:
// 	200: empty
//	400: error
//	404: error
//	500: error
func (f *Files) GetFile(rw http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	exp := r.URL.Query().Get("expires")
	sign := r.URL.Query().Get("signature")

	err := f.signedUrl.ValidateSignedUrl(id, exp, sign)
	if err != nil {
		http.Error(rw, "Invalid request: "+err.Error(), http.StatusBadRequest)
		return
	}

	fp, err := f.cache.Get(id)
	if err != nil {
		http.Error(rw, "Failed to read filepath from cache", http.StatusInternalServerError)
		return
	}

	err = f.store.Read(fp, rw)
	if err != nil {
		http.Error(rw, "Failed to read the file: \n" + err.Error(), http.StatusInternalServerError)
		return
	}

	rw.Header().Set("Content-Type", "application/octet-stream")
	rw.WriteHeader(http.StatusOK)
}

// swagger:route DELETE /{category}/{id}/{filename} files deleteFile
//
// Removes requested file from the filesystem. Doesn't remove any directories. Returns an error when file is not found
//
// Responses:
// 	204: empty
//	404: error
//	500: error
func (f *Files) DeleteFile(rw http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	c := r.PathValue("category")
	fn := r.PathValue("filename")

	fp := filepath.Join(c, id, fn)

	err := f.store.Delete(fp)
	if err != nil {
		if err == files.ErrFileNotFound {
			http.Error(rw, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(rw, "Failed to delete the file: \n" + err.Error(), http.StatusInternalServerError)
		return
	}

	rw.WriteHeader(http.StatusNoContent)
}
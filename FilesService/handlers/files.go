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

// TODO: Add paginated requests handling

// Handles post file request. Doesn't allow for overwriting
func (f *Files) PostFile(rw http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	c := r.PathValue("category")
	fn := r.PathValue("filename")

	fp := filepath.Join(c, id, fn)

	err := f.store.Write(fp, r.Body)
	if err != nil {
		http.Error(rw, "Failed to create the file: \n" + err.Error(), http.StatusBadRequest)
		return
	}
}

// Handles put file request. Doesn't allow for file creation
func (f *Files) PutFile(rw http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	c := r.PathValue("category")
	fn := r.PathValue("filename")

	fp := filepath.Join(c, id, fn)

	err := f.store.Overwrite(fp, r.Body)
	if err != nil {
		http.Error(rw, "Failed to update the file: \n" + err.Error(), http.StatusBadRequest)
		return
	}
}

func (f *Files) GetFileUrl(rw http.ResponseWriter, r *http.Request) {
	c := r.PathValue("category")
	id := r.PathValue("id")
	fn := r.PathValue("filename")

	fp := filepath.Join(c, id, fn)

	// generate uid and cache it with corresponding filepath
	tmpId := caches.GenerateUUID()
	f.cache.Set(tmpId, fp)

	// generate signedurl query based on uid
	url := f.signedUrl.GenerateSignedUrl(tmpId)

	rw.Write([]byte(url))
}

// Handles get file request from signed url
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
		http.Error(rw, "Failed to read cache", http.StatusInternalServerError)
	}

	err = f.store.Read(fp, rw)
	if err != nil {
		http.Error(rw, "Failed to read the file: \n" + err.Error(), http.StatusNotFound)
		return
	}

	rw.Header().Set("Content-Type", "application/octet-stream")
}

// Handles delete file request
func (f *Files) DeleteFile(rw http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	c := r.PathValue("category")
	fn := r.PathValue("filename")

	fp := filepath.Join(c, id, fn)

	err := f.store.Delete(fp)
	if err != nil {
		http.Error(rw, "Failed to delete the file: \n" + err.Error(), http.StatusBadRequest)
		return
	}
}
package handlers

import (
	"fmt"
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
	logger	*log.Logger
	store	files.Storage
	cache	caches.Cache
	signedUrl	signedurl.SignedUrl
}

func NewFiles(s files.Storage, l *log.Logger, c caches.Cache) *Files {
	return &Files{
		store: s, 
		logger: l,
		cache: c,
		signedUrl: *signedurl.NewSignedUrl(
			"Secret key my boy",
			"localhost:9090/url",
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
	ss := f.signedUrl.GenerateSignedUrl(tmpId)

	// return signedurl

	// TODO: Modify signedquery to generate full signedUrl
	// Add a basePath in constructor parameter so each endpoint can handle it's own files
	// Instead of handling it by a separate endpoint

	// TODO: CONTINUE FROM HERE
	rw.Write([]byte(ss))
}

// Handles get file request
func (f *Files) GetFile(rw http.ResponseWriter, r *http.Request) {
	c := r.PathValue("category")
	id := r.PathValue("id")
	fn := r.PathValue("filename")

	fp := filepath.Join(c, id, fn)

	err := f.store.Read(fp, rw)
	if err != nil {
		http.Error(rw, "Failed to read the file: \n" + err.Error(), http.StatusBadRequest)
		return
	}

	rw.Header().Set("Content-Type", "application/octet-stream")
    rw.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fn))
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
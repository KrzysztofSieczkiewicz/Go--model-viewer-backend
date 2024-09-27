package handlers

import (
	"fmt"
	"log"
	"net/http"
	"path/filepath"

	"github.com/KrzysztofSieczkiewicz/go--model-viewer-backend/FilesService/caches"
	"github.com/KrzysztofSieczkiewicz/go--model-viewer-backend/FilesService/files"
)

// Handler for reading and writing files to provided storage
type Files struct {
	logger	*log.Logger
	store	files.Storage
	cache	caches.Cache
}

func NewFiles(s files.Storage, l *log.Logger, c caches.Cache) *Files {
	return &Files{store: s, logger: l}
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

// Handles get file request
func (f *Files) GetFile(rw http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	c := r.PathValue("category")
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
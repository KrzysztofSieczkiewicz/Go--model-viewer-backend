package handlers

import (
	"FilesService/files"
	"log"
	"net/http"
	"path/filepath"
)

// Handler for reading and writing files to provided storage
type Files struct {
	logger	*log.Logger
	store	files.Storage
}

func NewFiles(s files.Storage, l *log.Logger) *Files {
	return &Files{store: s, logger: l}
}

func (f *Files)PostFile(rw http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	fn := r.PathValue("filname")

	// TODO: CREATE A data/file.go to handle file collections (and to allow querying for types, names, etc)
	// ALSO -> handle paginated requests

	f.saveFile(id, fn, rw, r)
}



func (f *Files) saveFile(id string, path string, rw http.ResponseWriter, r *http.Request) {
	fp := filepath.Join(id, path)
	err := f.store.Write(fp, r.Body)
	if err != nil {
		http.Error(rw, "Unable to save file", http.StatusInternalServerError)
	}
}
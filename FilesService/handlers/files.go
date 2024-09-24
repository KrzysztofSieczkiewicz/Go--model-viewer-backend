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

// TODO: Add paginated requests handling

// TODO: think about a way of providing filepath and other data (including the file) to the request so a structured 
// file system can be created
// preferably, filepath should start with file TYPE (this will be solved separate endpoints for each file type) so each filetype will also have separate storage
// next - there is no need for nested folders structure as each file has unique id
// the only question is - do You need a filename to be added to an ID 
// potentially it increases readability for human readers and further reduces collison chance, but that's not that impactful
// on the other hand it requires passing additional data with requests (like filename).
// CURRENTLY - it seems that better solution is to rely purely on ID

// Handles post file request. Doesn't allow for overwriting
func (f *Files) PostFile(rw http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	fn := r.PathValue("filename")

	fp := filepath.Join(id, fn)

	f.store.Write(fp, r.Body)
}

// Handles put file request. Doesn't allow for file creation
func (f *Files) PutFile(rw http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	fn := r.PathValue("filename")

	fp := filepath.Join(id, fn)

	f.store.Overwrite(fp, r.Body)
}
// TODO: should be moved to the models package
package response

import "github.com/KrzysztofSieczkiewicz/go--model-viewer-backend/FilesService/models"

// swagger:model empty
type EmptyResponse struct {}

// swagger:model message
type MessageResponse struct {
	Message string `json:"message"`
}

// swagger:model directoryContents
type DirectoryContentsResponse struct {
    // List of directories stored inside given category
    Contents    []models.CollectionContent  `json:"contents"`
}

// swagger:model fileByteStream
type FileResponse struct {
    // File bytestream
    File []byte
}

// swagger:model fileUrl
type FileUrlResponse struct {
    // filename
    Filename   string  `json:"filename"`
    // url pointing to the resource
    URL        string  `json:"url"`
}
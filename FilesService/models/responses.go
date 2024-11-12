package models

// swagger:model empty
type EmptyResponse struct {}

// swagger:model message
type MessageResponse struct {
	Message string `json:"message"`
}

// swagger:model directoryContents
type DirectoryContentsResponse struct {
    // List of directories stored inside given category
    Contents    []DirContent  `json:"contents"`
}

// swagger:model fileByteStream
type FileResponse struct {
    // File bytestream
    File []byte
}

// swagger:model fileUrl
type FileUrlResponse struct {
    // url pointing to the resource
    URL        string  `json:"url"`
}
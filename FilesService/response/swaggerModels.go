package response

// swagger:model empty
type EmptyResponse struct {}

// swagger:model message
type MessageResponse struct {
	Message string `json:"message"`
}

// swagger:model categoryContents
type CategoryResponse struct {
    // List of directories stored inside given category
    Directories    []string  `json:"directories"`
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
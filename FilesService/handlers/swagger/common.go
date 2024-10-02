package swagger

// swagger:response empty
type emptyResponse struct {
}

// swagger:response error
type errorResponse struct {
	// Error description
	// in: body
    // type: string
	Message string `json:"message"`
}

// swagger:response fileResponse
type fileResponse struct {
    // The file being returned
    // in: body
    // type: file
    File []byte
}

// swagger:response urlResponse
type urlResponse struct {
    // The URL pointing to the file
    // in: body
    // type: string
    FileUrl string
}

// Dummy function to avoid "unused" errors
func init() {
    _ = emptyResponse{}
    _ = errorResponse{}
    _ = fileResponse{}
    _ = urlResponse{}
}
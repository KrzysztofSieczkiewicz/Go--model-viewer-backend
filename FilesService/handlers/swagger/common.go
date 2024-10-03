package swagger

// swagger:response message
type message struct {
    // Returned message
	Message string `json:"message"`
}

// swagger:response imageUrlJson
type imageUrlJson struct {
    // ID of imageset
    ID          string  `json:"id"`
    // filename of returned image
    Filename    string  `json:"filename"`
    // url pointing to the resource
    URL         string  `json:"url"`
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
    _ = message{}
    _ = imageUrlJson{}
    
    _ = errorResponse{}
    _ = fileResponse{}
    _ = urlResponse{}
}
package response

// swagger:response imageUrlJson
type ImageUrlResponse struct {
	// ID of imageset
    ID          string  `json:"id"`
    // filename of returned image
    Filename    string  `json:"filename"`
    // url pointing to the resource
    URL         string  `json:"url"`
}
package response

import "github.com/KrzysztofSieczkiewicz/go--model-viewer-backend/FilesService/data"

// swagger:response imageUrlJson
type ImageUrlResponse struct {
    // filename of returned image
    Filename   string  `json:"filename"`
    // url pointing to the resource
    URL        string  `json:"url"`
}

// swagger:response imageSetJson
type ImageSetResponse struct {
    // ID of imageset
    ID         string        `json:"id"`
    // Category
    Category   string        `json:"category"`
    // Available images
    Images     []*data.Image  `json:"images"`
}

// swagger:response categoryContentsJson
type CategoryResponse struct {
    // List of directories stored inside given category
    ImageSets    []string  `json:"imageSets"`
}
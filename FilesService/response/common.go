package response

// swagger:response empty
type EmptyResponse struct {}

// swagger:response messageJson
type MessageResponse struct {
    // Returned message
	Message string `json:"message"`
}
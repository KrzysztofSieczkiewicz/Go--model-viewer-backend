package swagger

// Generic empty response with no content
// swagger:response empty
type EmptyResponse struct {
}

// Generic error message returned with a string message
// swagger:response error
type ErrorResponse struct {
	// Error description
	// in:body
	Message string
}
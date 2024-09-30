package swagger

// Empty response with no content
// swagger:response empty
type EmptyResponse struct {
}

// Error returned with a string message
// swagger:response error
type ErrorResponse struct {
	// Error description
	// in:body
	Message string
}
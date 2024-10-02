package swagger

// Generic empty response with no content
// swagger:response empty
type EmptyResponse struct {
}

// Generic eror message returned as a string
// swagger:response error
type ErrorResponse struct {
	// Error description
	// in:body
	Message string
}

// Validation errors defined as an array of strings
// swagger:response errorValidation
type ValidationErrorResponse struct {
	// Collection of encountered errors
	// in:body
	Body GenericErrors
}


// GenericError is a generic error message returned by a server
type GenericError struct {
	Message string `json:"message"`
}

// GenericErrors is a collection of validation error messages
type GenericErrors struct {
	Messages []string `json:"messages"`
}
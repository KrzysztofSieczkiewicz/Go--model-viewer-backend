package swagger

// swagger:response okResponse
type OkResponse struct {
}

// swagger:response createdResponse
type CreatedResponse struct {
}

// swagger:response noContentResponse
type NoContentResponse struct {
	// explanation message
	// in:body
	Message string
}

// swagger:response notFoundResponse
type NotFoundResponse struct {
	// explanation message
	// in:body
	Message string
}

// swagger:response internalServerErrorResponse
type InternalServerErrorReponse struct {
	// explanation message
	// in:body
	Message string
}
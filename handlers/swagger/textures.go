package swagger

import "github.com/KrzysztofSieczkiewicz/ModelViewerBackend/data"

// swagger:parameters getTexture putTexture deleteTexture
type TextureIdParameter struct {
	// in:path
	// name: id
	// description: The id of the texture in the database
	// required: true
	ID string `json:"id"`
}



// swagger:response getTextureResponse
type GetTextureResponse struct {
	// Parameters of the texture returned based on the id
	// in: body
	Body data.Texture
}

// swagger:response getTexturesResponse
type GetTexturesResponse struct {
	// Data of all textures in the database
	// in: body
	Body []data.Texture
}

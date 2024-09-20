package swagger

import "github.com/KrzysztofSieczkiewicz/go--model-viewer-backend/SceneManagerApi/data"

// swagger:parameters getTexture putTexture deleteTexture
type TextureIdParams struct {
	// in:path
	// name: id
	// description: The id of the texture in the database
	// required: true
	ID string `json:"id"`
}

// swagger:parameters postTexture putTexture
type TextureParams struct {
	// Texture data structure to Create or Update
	// Note: id field will be ignored by both Create and Update operations
	// in:body
	Body data.Texture
}



// swagger:response getTexture
type GetTextureResponse struct {
	// Parameters of the texture returned based on the id
	// in: body
	Body data.Texture
}

// swagger:response getTextures
type GetTexturesResponse struct {
	// Data of all textures in the database
	// in: body
	Body []data.Texture
}

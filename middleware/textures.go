package middleware

import (
	"context"
	"fmt"
	"net/http"

	"github.com/KrzysztofSieczkiewicz/ModelViewerBackend/data"
	"github.com/KrzysztofSieczkiewicz/ModelViewerBackend/internal/utils"
)

type KeyTexture struct{}

func TextureJsonValidation(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		texture := &data.Texture{}

		err := utils.FromJSON(texture, r.Body)
		if err != nil {
			http.Error(rw, "Unable to unmarshal Texture object from JSON:\n"+err.Error(), http.StatusBadRequest)
			return
		}

		err = texture.Validate()
		if err != nil {
			http.Error(
				rw, 
				fmt.Sprintf("Error validating texture:\n %s", err), 
				http.StatusBadRequest,
			)
			return
		}

		ctx := context.WithValue(r.Context(), KeyTexture{}, texture)
		req := r.WithContext(ctx)

		next.ServeHTTP(rw, req)
	})
}
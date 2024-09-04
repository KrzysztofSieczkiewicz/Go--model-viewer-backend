package middleware

import (
	"context"
	"net/http"

	"github.com/KrzysztofSieczkiewicz/ModelViewerBackend/internal/data"
)

type KeyTexture struct{}

func TextureJsonValidation(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		texture := &data.Texture{}

		err := texture.FromJSON(r.Body)
		if err != nil {
			http.Error(rw, "Unable to unmarshal Texture object from JSON:\n"+err.Error(), http.StatusBadRequest)
			return
		}

		ctx := context.WithValue(r.Context(), KeyTexture{}, texture)
		req := r.WithContext(ctx)

		next.ServeHTTP(rw, req)
	})
}
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	extMidddleware "github.com/go-openapi/runtime/middleware"

	"github.com/KrzysztofSieczkiewicz/ModelViewerBackend/handlers"
	"github.com/KrzysztofSieczkiewicz/ModelViewerBackend/middleware"
)

// TODO: Handle proper response headers writing (for no reason You're always responding with plain text instead of json)

func main() {
	l := log.New(os.Stdout, "texture-api", log.LstdFlags)

	// Create the handlers
	texturesHandler := handlers.NewHandler(l);

	// Initialize the ServeMux and register the handlers
	router := http.NewServeMux()

	router.HandleFunc("GET /textures", texturesHandler.GetTextures)
	router.HandleFunc("POST /textures", withMiddleware(texturesHandler.PostTexture, middleware.TextureJsonValidation))
	router.HandleFunc("PUT /textures/{id}", withMiddleware(texturesHandler.PutTexture, middleware.TextureJsonValidation))
	router.HandleFunc("GET /textures/{id}", texturesHandler.GetTexture)
	router.HandleFunc("DELETE /textures/{id}", texturesHandler.DeleteTexture)
	//router.HandleFunc("GET /textures/{id}/thumbnail", texturesHandler.GetThumbnail) // BETTER HANDLED BY GET TEXTURE
	//router.HandleFunc("GET /textures/{id}/image/{type}/{size}", texturesHandler.GetImage)

	// Handle OpenAPI doc request
	opts := extMidddleware.RedocOpts{SpecURL: "/swagger.yaml"}
	sh := extMidddleware.Redoc(opts, nil)
	router.Handle("/docs", sh)
	router.Handle("/swagger.yaml", http.FileServer(http.Dir("./")))

	stack := middleware.CreateStack(
		middleware.Cors,
		middleware.Logging,
	)
	// Initialize the new server
	s := &http.Server{
		Addr: ":9090",
		Handler: stack(router),
		IdleTimeout: 120*time.Second,
		ReadTimeout: 1*time.Second,
		WriteTimeout: 1*time.Second,
	}

	// Start the server
	go func() {
		err := s.ListenAndServe()
		if err != nil {
			l.Fatal(err)
		}
	}()

	// Register signals for graceful service termination
	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, os.Interrupt)
	signal.Notify(signalChannel, syscall.SIGTERM)

	sig := <- signalChannel
	l.Println("Received terminate. Gracefully shutting down...", sig)

	tc, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	s.Shutdown(tc)
}


// Wraps function in the provided middleware.
// Returns wrapper as a HandlerFunc
// Provides a way for single middleware injections
func withMiddleware(handlerFunction func(http.ResponseWriter, *http.Request), mw func(http.Handler) http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mw(http.HandlerFunc(handlerFunction)).ServeHTTP(w, r)
	})
}
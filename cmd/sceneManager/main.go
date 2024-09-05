package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/KrzysztofSieczkiewicz/ModelViewerBackend/internal/handlers"
	"github.com/KrzysztofSieczkiewicz/ModelViewerBackend/internal/middleware"
)

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

	stack := middleware.CreateStack(
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

	tc, err := context.WithTimeout(context.Background(), 30*time.Second)
	if err != nil {
		l.Fatal("Failed to set context with timeout. Shutting down abruptly... \n", err)
	}
	s.Shutdown(tc)
}


// Wraps function in the provided middleware.
// Returns HandlerFunc to be provided to the router.HandleFunc()
// Provides a way for single middleware injections for particular routes
func withMiddleware(handlerFunction func(http.ResponseWriter, *http.Request), mw func(http.Handler) http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mw(http.HandlerFunc(handlerFunction)).ServeHTTP(w, r)
	})
}
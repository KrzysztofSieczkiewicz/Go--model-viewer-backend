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
	router.HandleFunc("POST /textures", texturesHandler.PostTexture)
	router.HandleFunc("PUT /textures/{id}", texturesHandler.PutTexture)
	router.HandleFunc("GET /textures/{id}", texturesHandler.GetTexture)

	// Initialize the new server
	s := &http.Server{
		Addr: ":9090",
		Handler: middleware.Logging(router),
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

	tc, _ := context.WithTimeout(context.Background(), 30*time.Second)
	s.Shutdown(tc)
}
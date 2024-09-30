// Package classification FilesService.
//
// Documentation of FilesApi
//
//	Schemes: http, https
//	Host: localhost:9090
//	BasePath: /files/
//	Version: 1.0.0
//
// swagger:meta
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/KrzysztofSieczkiewicz/go--model-viewer-backend/FilesService/caches"
	"github.com/KrzysztofSieczkiewicz/go--model-viewer-backend/FilesService/files"
	"github.com/KrzysztofSieczkiewicz/go--model-viewer-backend/FilesService/handlers"
	"github.com/KrzysztofSieczkiewicz/go--model-viewer-backend/FilesService/middleware"
	extMidddleware "github.com/go-openapi/runtime/middleware"

	"github.com/joho/godotenv"
)

// DONE: Add swagger documentation
// TODO: Add proper success headers and responses for handler. Update swagger desc.
// Give response body to 200 responses (?)
// Register content type for GetUrl
// TODO: Create separate handlers for different file types
// TODO: Update gitignore
// TODO: Improve logging

func main() {
	l := log.New(os.Stdout, "FilesService", log.LstdFlags)

	// Load .env file and get env variables
	err := godotenv.Load()
    if err != nil {
        log.Fatalf("Error loading .env file")
    }
	hostUrl := os.Getenv("HOST")
	bindAddress := os.Getenv("BIND_ADDRESS")
	baseFilePath := os.Getenv("BASE_FILE_PATH")

	baseUrl := hostUrl + bindAddress

	// Initialize the local files storage with Max file size: 5MB
	fs, err := files.NewLocal(baseFilePath, 5)
	if err != nil {
		l.Fatal("Unable to initialize local storage")
	}

	// Initialize a cache
	fc := caches.NewFreeCache(50, 2)

	// Initialize the ServeMux
	router := http.NewServeMux();

	// Initialize and register the handlers
	fh := handlers.NewFiles(baseUrl, fs, l, fc)
	router.HandleFunc("GET /files/", fh.GetFile)
	router.HandleFunc("POST /files/{category}/{id}/{filename}", fh.PostFile)
	router.HandleFunc("PUT /files/{category}/{id}/{filename}", fh.PutFile)
	router.HandleFunc("DELETE /files/{category}/{id}/{filename}", fh.DeleteFile)

	router.HandleFunc("GET /url/{category}/{id}/{filename}", fh.GetFileUrl)


	// Handle OpenAPI doc request
	opts := extMidddleware.RedocOpts{SpecURL: "/swagger.yaml"}
	sh := extMidddleware.Redoc(opts, nil)
	router.Handle("/docs", sh)
	router.Handle("/swagger.yaml", http.FileServer(http.Dir("./")))

	// Create middleware stack
	stack := middleware.CreateStack(
	)

	// Initialize the new server
	s := &http.Server{
		Addr: bindAddress,
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
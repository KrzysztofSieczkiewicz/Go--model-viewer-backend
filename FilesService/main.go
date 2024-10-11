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
// DONE: Add proper success headers and responses for handler. Update swagger desc.
// Give response body to 200 responses (?)
// Register content type for GetUrl
// DONE: Create separate handlers for different file types
// NO NEED: Update gitignore
// TODO: Improve logging
// TODO: Implement file type validation (based on filename decide if file is correct)

func main() {
	// Initialize logger
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
	router.HandleFunc("GET /files/{category}/{id}/{filename}", fh.GetFileUrl)
	router.HandleFunc("POST /files/{category}/{id}/{filename}", fh.PostFile)
	router.HandleFunc("PUT /files/{category}/{id}/{filename}", fh.PutFile)
	router.HandleFunc("DELETE /files/{category}/{id}/{filename}", fh.DeleteFile)

	// IMAGES
	ih := handlers.NewImages(baseUrl, fs, l, fc)
	router.HandleFunc("GET /images/{category}/{id}", ih.GetUrl)
	router.HandleFunc("GET /images/", ih.GetImage) // TODO: HANDLE THIS PROPERLY
	router.HandleFunc("POST /images/{category}/{id}", ih.PostImage)
	router.HandleFunc("PUT /images/{category}/{id}", ih.PutImage)
	router.HandleFunc("DELETE /images/{category}/{id}", ih.DeleteImage)

	// IMAGE SETS & CATEGORIES
	ish := handlers.NewImageSets(baseUrl, fs, l, fc)
	router.HandleFunc("GET /imageSets/{category}/{id}", ish.GetImageSet)
	router.HandleFunc("POST /imageSets/{category}/{id}", ish.PostImageSet)
	router.HandleFunc("PUT /imageSets/{category}/{id}", ish.PutImageSet)
	router.HandleFunc("DELETE /imageSets/{category}/{id}", ish.DeleteImageSet)

	router.HandleFunc("GET /imageCategories/{category}", ish.GetCategory)
	router.HandleFunc("POST /imageCategories/{category}", ish.PostCategory)
	router.HandleFunc("PUT /imageCategories/{category}", ish.PutCategory)
	router.HandleFunc("DELETE /imageCategories/{category}", ish.DeleteCategory)

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
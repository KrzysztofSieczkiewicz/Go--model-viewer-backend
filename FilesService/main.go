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
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/KrzysztofSieczkiewicz/go--model-viewer-backend/FilesService/caches"
	"github.com/KrzysztofSieczkiewicz/go--model-viewer-backend/FilesService/files"
	"github.com/KrzysztofSieczkiewicz/go--model-viewer-backend/FilesService/handlers"
	"github.com/KrzysztofSieczkiewicz/go--model-viewer-backend/FilesService/middleware"
	swaggerMiddleware "github.com/go-openapi/runtime/middleware"

	"github.com/joho/godotenv"
)

// DONE: Add swagger documentation
// DONE: Add proper success headers and responses for handler. Update swagger desc.
// DONE: Give response body to 200 responses (?)
// DONE: Register content type for GetUrl
// DONE: Create separate handlers for different file types
// NO NEED: Update gitignore
// DONE: Improve logging
// DONE: Improve swagger annotations (add model annotations, clean up the response annotations)
// DONE: Clean up models, responses etc
// DONE: Improve local.go with proper code sharing and new common funcs - too much repetiton + occasional verbose/non-functioning checks
// continue clearing the code, remember about unused errors.go in the files directory
// DONE: Implement file type validation (based on filename decide if file is correct) - check Validator implementation from sceneManager
// DONE: Clean up the handlers and methods - consider what data should be moved to jsons - preferably remove most data from url into json body

// TODO: Test all endpoints + fix file write err (access is denied)

// TODO: Revise data validators
// TODO: Modify current endpoint structure -> instead of categories -> just manage directories for separate storages with reused handler
// TODO: Write unit tests for storage and data packages

// TODO: TESTING
// - PutImageSet -> test if it changes only imageset name/location without renaming category
// -

func main() {
	// Initialize logger
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))

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
	fs, err := files.NewLocal(baseFilePath, 5, logger)
	if err != nil {
		logger.Error("Unable to initialize local storage")
	}

	// Initialize a cache
	fc := caches.NewFreeCache(50, 2)

	// Initialize the ServeMux
	router := http.NewServeMux();

	// Initialize and register the handlers
	fh := handlers.NewFiles(baseUrl, fs, logger, fc)
	router.HandleFunc("GET /files/", fh.GetFile)
	router.HandleFunc("GET /files/{category}/{id}/{filename}", fh.GetFileUrl)
	router.HandleFunc("POST /files/{category}/{id}/{filename}", fh.PostFile)
	router.HandleFunc("PUT /files/{category}/{id}/{filename}", fh.PutFile)
	router.HandleFunc("DELETE /files/{category}/{id}/{filename}", fh.DeleteFile)

	// IMAGES
	ih := handlers.NewImages(baseUrl, fs, logger, fc)
	router.HandleFunc("GET /images/{category}/{id}", ih.GetUrl)
	router.HandleFunc("GET /images/", ih.GetImage) // TODO: HANDLE THIS PROPERLY
	router.HandleFunc("POST /images/{category}/{id}", ih.PostImage)
	router.HandleFunc("PUT /images/{category}/{id}", ih.PutImage)
	router.HandleFunc("DELETE /images/{category}/{id}", ih.DeleteImage)

	// IMAGE SETS & CATEGORIES
	ish := handlers.NewImageSets(baseUrl, fs, logger, fc)
	router.HandleFunc("GET /imageSets", ish.GetImageSet)
	router.HandleFunc("POST /imageSets", ish.PostImageSet)
	router.HandleFunc("PUT /imageSets", ish.PutImageSet)
	router.HandleFunc("DELETE /imageSets", ish.DeleteImageSet)

	router.HandleFunc("GET /imageCategories", ish.GetCategory)
	router.HandleFunc("POST /imageCategories", ish.PostCategory)
	router.HandleFunc("PUT /imageCategories", ish.PutCategory)
	router.HandleFunc("DELETE /imageCategories", ish.DeleteCategory)

	// Handle OpenAPI doc request
	opts := swaggerMiddleware.RedocOpts{SpecURL: "/swagger.yaml"}
	sh := swaggerMiddleware.Redoc(opts, nil)
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
			logger.Warn(err.Error())
		}
	}()

	// Register signals for graceful service termination
	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, os.Interrupt)
	signal.Notify(signalChannel, syscall.SIGTERM)

	sig := <- signalChannel
	logger.Info(fmt.Sprintf("Received terminate. Gracefully shutting down... %s", sig.String()))

	tc, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	s.Shutdown(tc)
}
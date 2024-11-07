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
// DONE: Update gitignore
// DONE: Improve logging
// DONE: Improve swagger annotations (add model annotations, clean up the response annotations)
// DONE: Clean up models, responses etc
// DONE: Improve local.go with proper code sharing and new common funcs - too much repetiton + occasional verbose/non-functioning checks
// 		 continue clearing the code, remember about unused errors.go in the files directory
// DONE: Implement file type validation (based on filename decide if file is correct) - check Validator implementation from sceneManager
// DONE: Clean up the handlers and methods - consider what data should be moved to jsons - preferably remove most data from url into json body
// DONE: Test all endpoints + fix file write err (access is denied)
// DONE: Revise data validators
// DONE: Update Images models for requests (include category and id in the metadata)
// DONE: Enforce that category name cannot have ID-like structure and enforce specific ID formatting
// DONE: Add 3D assets handling
// DONE: Clean and fix validators
// DONE: Write unit tests for storage and data packages
// DONE: Register all endpoints and funcs

// TODO: Retest all endpoints with test data
// TODO: Resolve singular TODOs
// TODO: Last iteration through swagger annotations
// TODO: Pop a champagne

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
	modStorage, err := files.NewLocal(baseFilePath + "models", 5, logger)
	if err != nil {
		logger.Error("Failed to initialize models storage")
	}
	imgStorage, err := files.NewLocal(baseFilePath + "images", 5, logger)
	if err != nil {
		logger.Error("Failed to initialize images storage")
	}

	// Initialize a cache
	imgCache := caches.NewFreeCache(64, 2)
	modCache := caches.NewFreeCache(128, 5)

	// Initialize the ServeMux
	router := http.NewServeMux()

	// MODELS
	modh := handlers.NewModels(baseUrl, modStorage, logger, modCache)
	router.HandleFunc("GET /models/url", modh.GetModelUrl)
	router.HandleFunc("GET /models/{id}/{expires}/{signature}", modh.GetModel)
	router.HandleFunc("POST /models", modh.PostModel)
	router.HandleFunc("PUT /models/overwrite", modh.PutModel)
	router.HandleFunc("PUT /models/update", modh.PutModelData)
	router.HandleFunc("DELETE /models", modh.DeleteModel)

	modColh := handlers.NewCollections(baseUrl, modStorage, logger)
	router.HandleFunc("GET /models/collections", modColh.GetCollection)
	router.HandleFunc("POST /models/collections", modColh.PostCollection)
	router.HandleFunc("PUT /models/collections", modColh.PutCollection)
	router.HandleFunc("DELETE /models/collections", modColh.DeleteCollection)

	modCath := handlers.NewCategories(baseUrl, modStorage, logger)
	router.HandleFunc("GET /models/categories", modCath.GetCategory)
	router.HandleFunc("POST /models/categories", modCath.PostCategory)
	router.HandleFunc("PUT /models/categories", modCath.PutCategory)
	router.HandleFunc("DELETE /models/categories", modCath.DeleteCategory)

	// IMAGES
	imgh := handlers.NewImages(baseUrl, imgStorage, logger, imgCache)
	router.HandleFunc("GET /images/url", imgh.GetImageUrl)
	router.HandleFunc("GET /images/{id}/{expires}/{signature}", imgh.GetImage)
	router.HandleFunc("POST /images", imgh.PostImage)
	router.HandleFunc("PUT /images/overwrite", imgh.PutImage)
	router.HandleFunc("PUT /images/update", imgh.PutImageData)
	router.HandleFunc("DELETE /images", imgh.DeleteImage)

	imgColh := handlers.NewCollections(baseUrl, imgStorage, logger)
	router.HandleFunc("GET /images/collections", imgColh.GetCollection)
	router.HandleFunc("POST /images/collections", imgColh.PostCollection)
	router.HandleFunc("PUT /images/collections", imgColh.PutCollection)
	router.HandleFunc("DELETE /images/collections", imgColh.DeleteCollection)

	imgCath := handlers.NewCategories(baseUrl, imgStorage, logger)
	router.HandleFunc("GET /images/categories", imgCath.GetCategory)
	router.HandleFunc("POST /images/categories", imgCath.PostCategory)
	router.HandleFunc("PUT /images/categories", imgCath.PutCategory)
	router.HandleFunc("DELETE /images/categories", imgCath.DeleteCategory)


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
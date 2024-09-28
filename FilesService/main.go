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
)

var basePath = "./fileStore"

func main() {
	l := log.New(os.Stdout, "FilesService", log.LstdFlags)

	// Create the local files storage
	// Max file size: 5MB
	fs, err := files.NewLocal(basePath, 5*1024*1000)
	if err != nil {
		l.Fatal("Unable to initialize local storage")
	}

	// Create a cache
	fc := caches.NewFreeCache(50, 2)

	// Create the handlers
	fh := handlers.NewFiles(fs, l, fc)

	// Initialize the ServeMux and register handler functions
	router := http.NewServeMux();

	router.HandleFunc("GET /files/{category}/{id}/{filename}", fh.GetFile)
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
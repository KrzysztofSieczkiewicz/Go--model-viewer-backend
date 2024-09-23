package main

import (
	"FilesService/files"
	"FilesService/middleware"
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	extMidddleware "github.com/go-openapi/runtime/middleware"
)

var basePath = "./store"

func main() {
	l := log.New(os.Stdout, "FilesService", log.LstdFlags)

	// Create the storage class with local storage
	// Max file size: 5MB - TODO: Move to a variable

	// TODO: create a handler for this new storage	
	_, err := files.NewLocal(basePath, 5*1024*1000)
	if err != nil {
		l.Fatal("Unable to initialize local storage")
	}

	// Create the handlers

	// Initialize the ServeMux and register handler functions
	router := http.NewServeMux();


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
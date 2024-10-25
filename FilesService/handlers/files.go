package handlers

import (
	"log/slog"
	"time"

	"github.com/KrzysztofSieczkiewicz/go--model-viewer-backend/FilesService/caches"
	"github.com/KrzysztofSieczkiewicz/go--model-viewer-backend/FilesService/files"
	"github.com/KrzysztofSieczkiewicz/go--model-viewer-backend/FilesService/signedurl"
)

// Handler managing files in given storage
type FilesHandler struct {
	baseUrl		string
	logger		*slog.Logger
	store		files.Storage
	cache		caches.Cache
	signedUrl	signedurl.SignedUrl
}

func NewFiles(baseUrl string, s files.Storage, l *slog.Logger, c caches.Cache) *FilesHandler {
	logger := l.With(slog.String("handler", "collections")) // TODO: do this when initializing logger in the main (you can pass the same logger to the store then)

	return &FilesHandler{
		baseUrl: baseUrl,
		store:   s,
		logger:  logger,
		cache:   c,
		signedUrl: *signedurl.NewSignedUrl(
			"Secret key my boy",
			baseUrl+"/files", // TODO: accept as parameter from main.go
			time.Duration(5*int(time.Minute)),
		),
	}
}

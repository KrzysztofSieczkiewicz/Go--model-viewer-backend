package handlers

import (
	"FilesService/files"
	"log"
)

// Handler for reading and writing files to provided storage
type Files struct {
	logger	log.Logger
	store	files.Storage
}

func NewFiles(s files.Storage, l log.Logger) *Files {
	return &Files{store: s, logger: l}
}


package files

import (
	"io"
	"os"
)

// Defines behavior for file operations
// Implementations might allow for local/cloud storage
type Storage interface {
	Save(path string, file io.Reader) error
	Read(path string) (*os.File, error)
}
package files

import (
	"io"
	"os"
)

// Defines behavior for file operations.
// Different implementations might allow for local/cloud storage
type Storage interface {
	Read(path string) (*os.File, error)
	Write(path string, file io.Reader) error
	Overwrite(path string, file io.Reader) error
}
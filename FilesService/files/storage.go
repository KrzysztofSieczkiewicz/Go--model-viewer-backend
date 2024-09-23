package files

import "io"

// Defines behavior for file operations
// Implementations might allow for local/cloud storage
type Storage interface {
	Save(path string, file io.Reader) error
}
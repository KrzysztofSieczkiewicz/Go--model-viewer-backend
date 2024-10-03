package files

import (
	"io"
)

// Defines behavior for file operations.
// Different implementations might allow for local/cloud storage
type Storage interface {
	Read(path string, writer io.Writer) error
	Write(path string, file io.Reader) error
	Overwrite(path string, file io.Reader) error
	Delete(path string) error
	CheckFile(path string) error

	MakeDirectory(path string) error
}
package files

import (
	"io"
)

// Defines behavior for file operations.
// Different implementations might allow for local/cloud storage
type Storage interface {
	// Reads the file at the provided path and returns a reader
	Read(path string, writer io.Writer) error

	// Create and write a file under provided path. Returns an error if file already exists
	Write(path string, file io.Reader) error

	// Overwrites provided file using temp file. Fails if requested file doesn't exist
	Overwrite(path string, file io.Reader) error

	// Deletes file under provided path. Returns error if file doesn't exist
	Delete(path string) error

	// Checks if file is stored in the filesystem
	CheckFile(path string) error


	// Creates requested directory or dir structure, returns an error if path already exists
	MakeDirectory(path string) error

	// Changes dir name and path. If old and new paths are in different directories functions as move
	RenameDirectory(oldPath string, newPath string) error
	
	// Deletes directory. Fails if it contains other directories
	DeleteDirectory(path string) error
}
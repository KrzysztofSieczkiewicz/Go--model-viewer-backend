package files

import (
	"io"
)

// Defines behavior for file operations.
// Different implementations might allow for local/cloud storage
type Storage interface {
	// Checks if filepath can be found in the filesystem
	IfExists(path string) error


	// Reads the file at the provided path and returns a reader
	ReadFile(path string, writer io.Writer) error

	// Create and write a file under provided path. Returns an error if file already exists
	WriteFile(path string, file io.Reader) error

	// Overwrites provided file using temp file. Fails if requested file doesn't exist
	OverwriteFile(path string, file io.Reader) error

	// Deletes file under provided path. Returns error if file doesn't exist
	DeleteFile(path string) error


	// Creates requested directory or dir structure, returns an error if path already exists
	CreateDirectory(path string) error

	// Changes dir name and path. If old and new paths are in different directories functions as move. Doesn't create new directories
	RenameDirectory(oldPath string, newPath string) error

	// Change directory name and path, Creates needed directories
	MoveDirectory(oldPath string, newPath string) error

	// Remove files stored in provided filepath. Omits subdirectories
	DeleteFiles(path string) error

	// Remove subdirectories in provided filepath. Omits files
	DeleteSubdirectories(path string) error
	
	// Deletes directory. Fails if directory is not empty
	DeleteDirectory(path string) error

	// Lists files in the directory
	ListFiles(path string) ([]string, error)

	// List subdirectories in the directory
	ListDirectories(path string) ([]string, error)
}
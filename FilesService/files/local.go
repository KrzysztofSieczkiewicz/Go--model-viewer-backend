package files

import (
	"io"
	"os"
	"path/filepath"

	"golang.org/x/xerrors"
)

// Implementation of the Storage interface that works for local disk
type Local struct {
	maxFileSize int    // Max file size in bytes
	basePath    string // Base path to the storage root
}

// Creates new Local filesystem with given basePath and max file size
func NewLocal(basePath string, maxSize int) (*Local, error) {
	p, err := filepath.Abs(basePath)
	if err != nil {
		return nil, err
	}

	return &Local{basePath: p}, nil
}

// Save the contents of provided wroter to the given relative path
// Removes the old file and writes a new one
func (l *Local) Save(path string, contents io.Reader) error {
	// get full filepath
	fp := l.fullPath(path)

	// get the file directory and ensure it exists
	d := filepath.Dir(fp)
	err := os.MkdirAll(d, os.ModePerm)
	if err != nil {
		return xerrors.Errorf("Unable to create directory: %w", err)
	}

	// if the file exists - delete it
	_, err = os.Stat(fp)
	if !os.IsNotExist(err) {
		return xerrors.Errorf("Unable to get file info: %w", err)
	}
	if err != nil {
		return xerrors.Errorf("Unable to delete file: %w", err)
	}

	// create a new file at the path
	f, err := os.Create(fp)
	if err != nil {
		return xerrors.Errorf("Unable to create file: %w", err)
	}
	defer f.Close()

	// write the contents to the new file
	// TODO: add a check to make sure max filesize is not exceeded
	_, err = io.Copy(f, contents)
	if err != nil {
		return xerrors.Errorf("Unable to write to file: %w", err)
	}

	return nil
}

// Gets the file at the provided path and returns a reader
func (l *Local) Read(path string) (*os.File, error) {
	// get the full filepath
	fp := l.fullPath(path)

	// open the file
	r, err := os.Open(fp)
	if err != nil {
		return nil, xerrors.Errorf("Unable to open file: %w", err)
	}

	return r, nil
}


// Returns the absolute path from provided relative one
func (l *Local) fullPath(path string) string {
	return filepath.Join(l.basePath, path)
}
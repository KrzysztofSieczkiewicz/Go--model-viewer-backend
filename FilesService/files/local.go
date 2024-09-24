package files

import (
	"io"
	"os"
	"path/filepath"

	"golang.org/x/xerrors"
)

// Implementation of the Storage interface that works for local disk
type Local struct {
	maxFileSize int64    // Max file size in bytes
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

// Reads the file at the provided path and returns a reader
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

// Create and write a file under provided path. Does not overwrite existing files 
// and will return an error if there already is an identical file
func (l *Local) Write(path string, contents io.Reader) error {
	// get full filepath
	fp := l.fullPath(path)

	// create a new file at the path
	f, err := os.OpenFile(fp, os.O_CREATE|os.O_EXCL|os.O_RDWR, 0666)
	if os.IsExist(err) {
		return xerrors.Errorf("file already exists: %s", fp)
	}
	if err != nil {
		return xerrors.Errorf("Unable to create file: %w", err)
	}

	// close the file when work is done, delete if any error occured
	defer func() {
        if err != nil {
            os.Remove(fp)
        }
        f.Close()
    }()

	// create a LimitedReader for max file size handling
    limitedReader := &io.LimitedReader{
        R: contents,
        N: l.maxFileSize + 1,
    }

	// write the contents to the new file
	_, err = io.Copy(f, limitedReader)
	if err != nil {
		os.Remove(fp)
		return xerrors.Errorf("Unable to write to file: %w", err)
	}

	// check if filesize limit was reached
	if limitedReader.N == 0 {
		os.Remove(fp)
		return xerrors.Errorf("file size exceeds the maximum limit of %d bytes", l.maxFileSize)
	}

	return nil
}

// Overwrites provided file using temp file. Fails if requested file doesn't exist
func (l *Local) Overwrite(path string, contents io.Reader) error {
	// get the firectory and filename
	fp := l.fullPath(path)
	tfp := fp + ".tmp"

	// check if requested file exists
	_, err := os.Stat(fp)
	if !os.IsNotExist(err) {
		return xerrors.Errorf("File does not exist: %w", err)
	}
	if err != nil {
		return xerrors.Errorf("Error during checking target file: %w", err)
	}

	// create and write to the temp file
	err = l.Write(tfp, contents)
	if err != nil {
		return xerrors.Errorf("Unable to create and write to the tmp file: %w", err)
	}

	// replace the original file with the temporary file
	err = os.Rename(tfp, fp)
    if err != nil {
		os.Remove(tfp)
        return xerrors.Errorf("Unable to replace target file: %w", err)
    }

	return nil
}

// Deletes file under provided path. Returns error if file doesn't exist
func (l *Local) Delete(path string) error {
	fp := l.fullPath(path)
	err := os.Remove(fp)
	if os.IsNotExist(err) {
		return xerrors.Errorf("Requested file doesn't exist: %w", err)
	}
	if err != nil {
		return xerrors.Errorf("Unable to remove target file: %w", err)
	}

	return nil
}


// Returns the absolute path from provided relative path
func (l *Local) fullPath(path string) string {
	return filepath.Join(l.basePath, path)
}
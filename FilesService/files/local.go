package files

import (
	"io"
	"os"
	"path/filepath"
)

// Implementation of the Storage interface that works for local disk
type Local struct {
	maxFileSize int64	// Max file size in bytes
	basePath    string	// Base path to the storage root
}

// Creates new Local filesystem with given basePath and max file size
func NewLocal(basePath string, maxSizeMB int) (*Local, error) {
	p, err := filepath.Abs(basePath)
	if err != nil {
		return nil, err
	}

	return &Local{
		basePath: p, 
		maxFileSize: int64(maxSizeMB*1024*1000),
	}, nil
}

// Reads the file at the provided path and returns a reader
func (l *Local) Read(path string, w io.Writer) error {
	fp := l.fullPath(path)

	// check if requested file exists
	_, err := os.Stat(fp)
	if err != nil {
		if os.IsNotExist(err) {
			return ErrFileNotFound
		}
		return ErrFileStat
	}

	// open the file
    f, err := os.Open(fp)
    if err != nil {
        return ErrFileRead
    }
    defer f.Close()

	// write the file contents into the writer
	_, err = io.Copy(w, f)
    if err != nil {
        return ErrFileWrite
    }

    return nil
}

// Create and write a file under provided path. Does not overwrite existing files 
// and will return an error if there already is an identical file
func (l *Local) Write(path string, contents io.Reader) error {
	fp := l.fullPath(path)

	// create directory stucture if it doesn't exist
	err := os.MkdirAll(filepath.Dir(fp), 0755)
	if err != nil {
		return ErrDirectoryCreate
	}

	// create a new file at the path
	f, err := os.OpenFile(fp, os.O_CREATE|os.O_EXCL|os.O_RDWR, 0666)
	if err != nil {
		if os.IsExist(err) {
			return ErrFileAlreadyExists
		}
		return ErrFileCreate
	}

	// close the file when done, delete the file if error occured
	defer func() {
        if err != nil {
            os.Remove(fp)
        }
        f.Close()
    }()

	// create a LimitedReader to limit file size
    limitedReader := &io.LimitedReader{
        R: contents,
        N: l.maxFileSize + 1,
    }

	// write the contents to the new file
	_, err = io.Copy(f, limitedReader)
	if err != nil {
		return ErrFileWrite
	}

	// check if filesize limit was reached
	if limitedReader.N == 0 {
		os.Remove(fp)
		return ErrFileSizeExceeded
	}

	return nil
}

// Overwrites provided file using temp file. Fails if requested file doesn't exist
func (l *Local) Overwrite(path string, contents io.Reader) error {
	fp := l.fullPath(path)
	tfp := fp + ".tmp"

	// check if requested file exists
	_, err := os.Stat(fp)
	if err != nil {
		if os.IsNotExist(err) {
			return ErrFileNotFound
		}
		return ErrFileStat
	}

	// create and write to the temp file
	err = l.Write(tfp, contents)
	if err != nil {
		return ErrFileWrite
	}

	// replace the original file with the temporary file
	err = os.Rename(tfp, fp)
    if err != nil {
		os.Remove(tfp)
        return ErrFileReplace
    }

	return nil
}

// Deletes file under provided path. Returns error if file doesn't exist
func (l *Local) Delete(path string) error {
	fp := l.fullPath(path)
	err := os.Remove(fp)
	if err != nil {
		if os.IsNotExist(err) {
			return ErrFileNotFound
		}
		return ErrFileDelete
	}

	return nil
}


// Returns the absolute path from provided relative path
func (l *Local) fullPath(path string) string {
	return filepath.Join(l.basePath, path)
}
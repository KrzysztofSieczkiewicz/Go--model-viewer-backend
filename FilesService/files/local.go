package files

import (
	"errors"
	"fmt"
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


func (l *Local) Read(path string, w io.Writer) error {
	fp := l.fullPath(path)

	// check if requested file exists
	l.CheckFile(path)

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


func (l *Local) Write(path string, contents io.Reader) error {
	fp := l.fullPath(path)

	// check if the directory exists
	dir := filepath.Dir(fp)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return ErrDirectoryNotFound
	}

	// create a new file at the path
	f, err := os.OpenFile(fp, os.O_CREATE|os.O_EXCL|os.O_RDWR, 0666)
	if err != nil {
		if os.IsExist(err) {
			return ErrFileAlreadyExists
		}
		fmt.Println(err)
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

// [IMMEDIATE] - add logging - find the issue with http 500 on each request
func (l *Local) Overwrite(path string, contents io.Reader) error {
	// temp filename
	tp := path + "_tmp"

	// check if requested file exists
	err := l.CheckFile(path)
	if err != nil {
		fmt.Println(err)
		return err
	}

	// create and write to the temp file
	err = l.Write(tp, contents)
	if err != nil {
		fmt.Println(err)
		return ErrFileWrite
	}

	// combine into full full paths
	fp := l.fullPath(path)
	tfp := l.fullPath(tp)

	// replace the original file with the temporary file
	err = os.Rename(tfp, fp)
    if err != nil {
		fmt.Println(err)
		os.Remove(tfp)
        return ErrFileReplace
    }

	return nil
}


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


func (l *Local) CheckFile(path string) error {
	fp := l.fullPath(path)

	_, err := os.Stat(fp)
	if err != nil {
		if os.IsNotExist(err) {
			return ErrFileNotFound
		}
		fmt.Println(err)
		return ErrFileStat
	}

	return nil
}


func (l *Local) MakeDirectory(path string) error {
	fp := l.fullPath(path)

	// check if the directory exists
	_, err := os.Stat(fp)
	if !os.IsNotExist(err) {
		return ErrDirectoryAlreadyExists
	}

	// create the directory
	err = os.MkdirAll(fp, 0755)
	if err != nil {
		return ErrDirectoryCreate
	}

	return nil
}


func (l *Local) RenameDirectory(oldPath string, newPath string) error {
	fop := l.fullPath(oldPath)
	fnp := l.fullPath(newPath)

	// check if the requested directory exists
	_, err := os.Stat(fop)
	if errors.Is(err, os.ErrNotExist) {
		return ErrDirectoryNotFound
	}

	// rename the directory
	err = os.Rename(fop, fnp)
	if err != nil {
		return ErrDirectoryRename
	}

	return nil
}


func (l *Local) MoveDirectory(oldPath string, newPath string) error {
	fop := l.fullPath(oldPath)
	fnp := l.fullPath(newPath)

	// check if the requested directory exists
	_, err := os.Stat(fop)
	if os.IsNotExist(err) {
		return ErrDirectoryNotFound
	}

	// check if the desired directory doesn't already exist
	_, err = os.Stat(fnp)
	if os.IsExist(err) {
		return ErrDirectoryAlreadyExists
	}

	// move the directory
	// TODO

	return nil
}


func (l *Local) DeleteDirectory(path string) error {
	fp := l.fullPath(path)

	// check if directory exists
	_, err := os.Stat(fp)
	if os.IsNotExist(err) {
		fmt.Printf("stat: %s", err)
		return ErrDirectoryNotFound
	}

	// open the dir
	dir, err := os.Open(fp)
	if err != nil {
		fmt.Printf("open: %s", err)
		return ErrDirectoryRead
	}
	defer dir.Close()

	// Read dir contents
	entries, err := dir.Readdir(-1)
	if err != nil {
		return ErrDirectoryRead
	}

	// Check if any entry is a directory
	for _, entry := range entries {
		if entry.IsDir() {
			fmt.Printf("check subdir: %s", err)
			return ErrDirectorySubdirectoryFound
		}
	}

	// Remove directory contents
	for _, entry := range entries {
		err = os.Remove(fp + "/" + entry.Name())
        if err != nil {
            return ErrFileDelete
        }
	}

	// close and remove the dir
	dir.Close()
	err = os.Remove(fp)
	if err != nil {
		fmt.Printf("delete: %s", err)
		return ErrDirectoryDelete
	}

	return nil
}


func (l *Local) ListFiles(path string) ([]string, error) {
	fp := l.fullPath(path)

	// check if the directory exists
	info, err := os.Stat(fp)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrDirectoryNotFound
		}
		return nil, ErrDirectoryStat
	}
	if !info.IsDir() {
		return nil, ErrNotDirectory
	}

	// open the dir
	dir, err := os.Open(fp)
	if err != nil {
		return nil, ErrDirectoryRead
	}
	defer dir.Close()

	// Read directory contents
	entries, err := dir.Readdir(-1)
	if err != nil {
		return nil, ErrDirectoryRead
	}

	// save filenames
	filenames := make([]string, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() { // Check if the entry is a file
			filenames = append(filenames, entry.Name()) // Add the filename to the slice
		}
	}

	return filenames, nil
}


func (l *Local) ListDirectories(path string) ([]string, error) {
	fp := l.fullPath(path)

	// check if the directory exists
	info, err := os.Stat(fp)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrDirectoryNotFound
		}
		return nil, ErrDirectoryStat
	}
	if !info.IsDir() {
		return nil, ErrNotDirectory
	}

	// open the dir
	dir, err := os.Open(fp)
	if err != nil {
		return nil, ErrDirectoryRead
	}
	defer dir.Close()

	// Read directory contents
	entries, err := dir.Readdir(-1)
	if err != nil {
		return nil, ErrDirectoryRead
	}

	// save subdirectories
	dirs := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() { // Check if the entry is a file
			dirs = append(dirs, entry.Name()) // Add the filename to the slice
		}
	}

	return dirs, nil
}

// Returns the absolute path from provided relative path
func (l *Local) fullPath(path string) string {
	return filepath.Join(l.basePath, path)
}
package files

import (
	"errors"
	"io"
	"log/slog"
	"os"
	"path/filepath"
)

// Implementation of the Storage interface that works for local disk
type Local struct {
	maxFileSize int64			// Max file size in bytes
	basePath    string			// Base path to the storage root
	logger		*slog.Logger	// Logger
}

// Creates new Local filesystem with given basePath and max file size
func NewLocal(basePath string, maxSizeMB int, l *slog.Logger) (*Local, error) {
	// Add logger detail
	logger := l.With(slog.String("store", basePath))

	// convert path to absolute path
	p, err := filepath.Abs(basePath)
	if err != nil {
		return nil, err
	}

	return &Local{
		basePath: p, 
		maxFileSize: int64(maxSizeMB*1024*1000),
		logger: logger,
	}, nil
}


func (l *Local) Read(path string, w io.Writer) error {
	l.logger.Info("Reading the file: " + path)

	fp := l.fullPath(path)

	// check if requested file exists
	l.CheckFile(path)

	// open the file
    f, err := os.Open(fp)
    if err != nil {
		l.logger.Error(err.Error())
        return ErrFileRead
    }
    defer f.Close()

	// write the file contents into the writer
	_, err = io.Copy(w, f)
    if err != nil {
		l.logger.Error(err.Error())
        return ErrFileWrite
    }

	l.logger.Info("Done reading the file: " + path)
    return nil
}


func (l *Local) Write(path string, contents io.Reader) error {
	l.logger.Info("Writing the file: " + path)

	fp := l.fullPath(path)

	// check if the directory exists
	dir := filepath.Dir(fp)
	_, err := os.Stat(dir)
	if os.IsNotExist(err) {
		l.logger.Warn(err.Error())
		return ErrDirectoryNotFound
	}

	// create a new file at the path
	f, err := os.OpenFile(fp, os.O_CREATE|os.O_EXCL|os.O_RDWR, 0666)
	if err != nil {
		if os.IsExist(err) {
			l.logger.Warn(err.Error())
			return ErrFileAlreadyExists
		}
		l.logger.Error(err.Error())
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
		l.logger.Error(err.Error())
		return ErrFileWrite
	}

	// check if filesize limit was reached
	if limitedReader.N == 0 {
		os.Remove(fp)
		l.logger.Error(err.Error())
		return ErrFileSizeExceeded
	}

	l.logger.Info("Done writing the file: " + path)
	return nil
}

// [IMMEDIATE] - add logging - find the issue with http 500 on each request
func (l *Local) Overwrite(path string, contents io.Reader) error {
	l.logger.Info("Overwriting the file: " + path)

	// temp filename
	tp := path + "_tmp"

	// check if requested file exists
	err := l.CheckFile(path)
	if err != nil {
		l.logger.Warn(err.Error())
		return err
	}

	// create and write to the temp file
	err = l.Write(tp, contents)
	if err != nil {
		l.logger.Error(err.Error())
		return ErrFileWrite
	}

	// combine into full full paths
	fp := l.fullPath(path)
	tfp := l.fullPath(tp)

	// replace the original file with the temporary file
	err = os.Rename(tfp, fp)
    if err != nil {
		os.Remove(tfp)
		l.logger.Error(err.Error())
        return ErrFileReplace
    }

	l.logger.Info("File overwritten: " + path)
	return nil
}


func (l *Local) Delete(path string) error {
	l.logger.Info("Deleting the file: " + path)

	fp := l.fullPath(path)
	err := os.Remove(fp)
	if err != nil {
		if os.IsNotExist(err) {
			l.logger.Warn(err.Error())
			return ErrFileNotFound
		}
		l.logger.Error(err.Error())
		return ErrFileDelete
	}

	l.logger.Info("Deleted the file: " + path)
	return nil
}


func (l *Local) CheckFile(path string) error {
	l.logger.Info("Checking the file: " + path)

	fp := l.fullPath(path)

	_, err := os.Stat(fp)
	if err != nil {
		if os.IsNotExist(err) {
			l.logger.Warn(err.Error())
			return ErrFileNotFound
		}
		l.logger.Error(err.Error())
		return ErrFileStat
	}

	l.logger.Info("File checked: " + path)
	return nil
}


func (l *Local) MakeDirectory(path string) error {
	l.logger.Info("Creating directory: " + path)
	fp := l.fullPath(path)

	// check if the directory exists
	_, err := os.Stat(fp)
	if !os.IsNotExist(err) {
		l.logger.Warn(err.Error())
		return ErrDirectoryAlreadyExists
	}

	// create the directory
	err = os.MkdirAll(fp, 0755)
	if err != nil {
		l.logger.Error(err.Error())
		return ErrDirectoryCreate
	}

	l.logger.Info("Created directory: " + path)
	return nil
}


func (l *Local) RenameDirectory(oldPath string, newPath string) error {
	l.logger.Info("Renaming directory: " + oldPath + " to: " + newPath)

	fop := l.fullPath(oldPath)
	fnp := l.fullPath(newPath)

	// check if the requested directory exists
	_, err := os.Stat(fop)
	if errors.Is(err, os.ErrNotExist) {
		l.logger.Warn(err.Error())
		return ErrDirectoryNotFound
	}

	// rename the directory
	err = os.Rename(fop, fnp)
	if err != nil {
		l.logger.Error(err.Error())
		return ErrDirectoryRename
	}

	l.logger.Info("Renamed directory: " + oldPath + " to: " + newPath)
	return nil
}


func (l *Local) MoveDirectory(oldPath string, newPath string) error {
	l.logger.Info("Moving directory from: " + oldPath + " to: " + newPath)

	fop := l.fullPath(oldPath)
	fnp := l.fullPath(newPath)

	// check if requested directory exists
	_, err := os.Stat(fop)
	if os.IsNotExist(err) {
		l.logger.Warn(err.Error())
		return ErrDirectoryNotFound
	}

	// check if the desired directory doesn't already exist
	_, err = os.Stat(fnp)
	if os.IsExist(err) {
		l.logger.Warn(err.Error())
		return ErrDirectoryAlreadyExists
	}

	// check if requested directory contains non-directories
	dir, err := os.Open(fop)
	if err != nil {
		l.logger.Error(err.Error())
		return ErrDirectoryRead
	}
	defer dir.Close()

	entries, err := dir.Readdir(-1)
	if err != nil {
		l.logger.Error(err.Error())
		return ErrDirectoryRead
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			l.logger.Warn(ErrDirectoryNonDirectoryFound.Error())
			return ErrDirectoryNonDirectoryFound
		}
	}
	dir.Close()

	// move requested directory
	err = os.Rename(fop, fnp)
	if err != nil {
		l.logger.Error(err.Error())
		return ErrDirectoryMove
	}

	l.logger.Info("Moved directory from: " + oldPath + " to: " + newPath)
	return nil
}

func (l *Local) DeleteFiles(path string) error {
	l.logger.Info("Removing files from directory: " + path)

	fp := l.fullPath(path)

	// check if directory exists
	_, err := os.Stat(fp)
	if os.IsNotExist(err) {
		l.logger.Warn(err.Error())
		return ErrDirectoryNotFound
	}

	// open the dir
	dir, err := os.Open(fp)
	if err != nil {
		l.logger.Error(err.Error())
		return ErrDirectoryRead
	}
	defer dir.Close()

	// Read dir contents
	entries, err := dir.Readdir(-1)
	if err != nil {
		l.logger.Error(err.Error())
		return ErrDirectoryRead
	}

	// Remove directory contents
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		err = os.Remove(fp + "/" + entry.Name())
        if err != nil {
			l.logger.Error(err.Error())
            return ErrFileDelete
        }
	}

	l.logger.Info("Removed files from directory: " + path)
	return nil
}

func (l *Local) DeleteSubdirectories(path string) error {
	l.logger.Info("Removing subdirectories from directory: " + path)

	fp := l.fullPath(path)

	// check if directory exists
	_, err := os.Stat(fp)
	if os.IsNotExist(err) {
		l.logger.Warn(err.Error())
		return ErrDirectoryNotFound
	}

	// open the dir
	dir, err := os.Open(fp)
	if err != nil {
		l.logger.Error(err.Error())
		return ErrDirectoryRead
	}
	defer dir.Close()

	// Read dir contents
	entries, err := dir.Readdir(-1)
	if err != nil {
		l.logger.Error(err.Error())
		return ErrDirectoryRead
	}

	// Remove directory contents
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		err = os.Remove(fp + "/" + entry.Name())
        if err != nil {
			l.logger.Error(err.Error())
            return ErrFileDelete
        }
	}
	
	l.logger.Info("Removed subdirectories from directory: " + path)
	return nil
}


func (l *Local) DeleteDirectory(path string) error {
	l.logger.Info("Removing directory: " + path)

	fp := l.fullPath(path)

	// check if directory exists
	_, err := os.Stat(fp)
	if os.IsNotExist(err) {
		l.logger.Warn(err.Error())
		return ErrDirectoryNotFound
	}

	// open the dir
	dir, err := os.Open(fp)
	if err != nil {
		l.logger.Error(err.Error())
		return ErrDirectoryRead
	}
	defer dir.Close()

	// Read dir contents
	entries, err := dir.Readdir(-1)
	if err != nil {
		l.logger.Error(err.Error())
		return ErrDirectoryRead
	}
	// Check if empty
	if len(entries) > 0 {
		l.logger.Warn(ErrDirectoryNotEmpty.Error())
		return ErrDirectoryNotEmpty
	}
	dir.Close()

	// remove the dir
	err = os.Remove(fp)
	if err != nil {
		l.logger.Error(err.Error())
		return ErrDirectoryDelete
	}

	l.logger.Info("Removed directory: " + path)
	return nil
}


func (l *Local) ListFiles(path string) ([]string, error) {
	l.logger.Info("Reading files from directory: " + path)
	fp := l.fullPath(path)

	// check if the directory exists
	info, err := os.Stat(fp)
	if err != nil {
		if os.IsNotExist(err) {
			l.logger.Warn(err.Error())
			return nil, ErrDirectoryNotFound
		}
		l.logger.Error(err.Error())
		return nil, ErrDirectoryStat
	}
	if !info.IsDir() {
		l.logger.Warn(ErrNotDirectory.Error())
		return nil, ErrNotDirectory
	}

	// open the dir
	dir, err := os.Open(fp)
	if err != nil {
		l.logger.Error(err.Error())
		return nil, ErrDirectoryRead
	}
	defer dir.Close()

	// Read directory contents
	entries, err := dir.Readdir(-1)
	if err != nil {
		l.logger.Error(err.Error())
		return nil, ErrDirectoryRead
	}

	// save filenames
	filenames := make([]string, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() { // Check if the entry is a file
			filenames = append(filenames, entry.Name()) // Add the filename to the slice
		}
	}

	l.logger.Info("Done reading files from directory: " + path)
	return filenames, nil
}


func (l *Local) ListDirectories(path string) ([]string, error) {
	l.logger.Info("Listing subdirectories in directory: " + path)
	fp := l.fullPath(path)

	// check if the directory exists
	info, err := os.Stat(fp)
	if err != nil {
		if os.IsNotExist(err) {
			l.logger.Warn(err.Error())
			return nil, ErrDirectoryNotFound
		}
		l.logger.Error(err.Error())
		return nil, ErrDirectoryStat
	}
	if !info.IsDir() {
		l.logger.Warn(ErrNotDirectory.Error())
		return nil, ErrNotDirectory
	}

	// open the dir
	dir, err := os.Open(fp)
	if err != nil {
		l.logger.Error(err.Error())
		return nil, ErrDirectoryRead
	}
	defer dir.Close()

	// Read directory contents
	entries, err := dir.Readdir(-1)
	if err != nil {
		l.logger.Error(err.Error())
		return nil, ErrDirectoryRead
	}

	// save subdirectories
	dirs := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() { // Check if the entry is a file
			dirs = append(dirs, entry.Name()) // Add the filename to the slice
		}
	}

	l.logger.Info("Listed subdirectories in directory: " + path)
	return dirs, nil
}

// Returns the absolute path from provided relative path
func (l *Local) fullPath(path string) string {
	return filepath.Join(l.basePath, path)
}
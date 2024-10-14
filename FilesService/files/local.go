package files

import (
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


func (l *Local) IfExists(path string) error {
	l.logger.Info("Looking for file: " + path)

	fp := l.fullPath(path)

	exists := l.verifyIfExists(fp)
	if !exists {
		l.logger.Info("File not found: " + path)
		return ErrNotFound
	}

	l.logger.Info("File found: " + path)
	return nil
}


func (l *Local) ReadFile(path string, w io.Writer) error {
	l.logger.Info("Reading the file: " + path)

	fp := l.fullPath(path)

	// check if requested file exists
	exists := l.verifyIfExists(fp)
	if !exists {
		l.logger.Warn("Requested file not found: " + path)
		return ErrNotFound
	}

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


func (l *Local) WriteFile(path string, contents io.Reader) error {
	l.logger.Info("Writing the file: " + path)

	fp := l.fullPath(path)

	// check if the directory exists
	dir := filepath.Dir(fp)
	exists := l.verifyIfExists(dir)
	if !exists {
		l.logger.Warn("Requested directory not found: " + path)
		return ErrNotFound
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

func (l *Local) OverwriteFile(path string, contents io.Reader) error {
	l.logger.Info("Overwriting the file: " + path)

	// combine into full paths
	fp := l.fullPath(path)
	tfp := l.fullPath(path + "_tmp")

	// check if file exists
	exists := l.verifyIfExists(fp)
	if !exists {
		l.logger.Warn("Requested file not found: " + path)
		return ErrNotFound
	}

	// create and write to the temp file
	err := l.WriteFile(tfp, contents)
	if err != nil {
		l.logger.Error(err.Error())
		return ErrFileWrite
	}

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


func (l *Local) DeleteFile(path string) error {
	l.logger.Info("Deleting the file: " + path)

	fp := l.fullPath(path)

	// check if file exists
	exists := l.verifyIfExists(fp)
	if !exists {
		l.logger.Warn("Requested file not found: " + path)
		return ErrNotFound
	}

	// remove the file
	err := os.Remove(fp)
	if err != nil {
		l.logger.Error(err.Error())
		return ErrFileDelete
	}

	l.logger.Info("Deleted the file: " + path)
	return nil
}

func (l *Local) MakeDirectory(path string) error {
	l.logger.Info("Creating directory: " + path)
	fp := l.fullPath(path)

	// check if the directory exists
	exists := l.verifyIfExists(fp)
	if exists {
		l.logger.Warn("Requested directory not found: " + path)
		return ErrFileAlreadyExists
	}

	// create the directory
	err := os.MkdirAll(fp, 0755)
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
	exists := l.verifyIfExists(fop)
	if !exists {
		l.logger.Warn("Requested directory not found: " + oldPath)
		return ErrNotFound
	}

	// rename the directory
	err := os.Rename(fop, fnp)
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
	exists := l.verifyIfExists(fop)
	if !exists {
		l.logger.Warn("Requested directory not found: " + oldPath)
		return ErrNotFound
	}

	// check if the desired directory doesn't already exist
	exists = l.verifyIfExists(fnp)
	if exists {
		l.logger.Warn("Requested directory already exists: " + newPath)
		return ErrNotFound
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
	exists := l.verifyIfExists(fp)
	if !exists {
		l.logger.Warn("Requested directory not found: " + path)
		return ErrNotFound
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
	exists := l.verifyIfExists(fp)
	if !exists {
		l.logger.Warn("Requested directory not found: " + path)
		return ErrNotFound
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
	exists := l.verifyIfExists(fp)
	if !exists {
		l.logger.Warn("Requested directory not found: " + path)
		return ErrNotFound
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
		l.logger.Warn(ErrDirNotEmpty.Error())
		return ErrDirNotEmpty
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
	exists := l.verifyIfExists(fp)
	if !exists {
		l.logger.Warn("Requested directory not found: " + path)
		return nil, ErrNotFound
	}

	// check if filepath is a directory
	info, _ := os.Stat(fp)
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
	exists := l.verifyIfExists(fp)
	if !exists {
		l.logger.Warn("Requested directory not found: " + path)
		return nil, ErrNotFound
	}

	// check if filepath is a directory
	info, _ := os.Stat(fp)
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

// Verifies if filepath exists in the filesystem
func (l *Local) verifyIfExists(fullPath string) (bool) {
	l.logger.Info("Looking for file: " + fullPath)

	_, err := os.Stat(fullPath)
	if err != nil {
		l.logger.Info("Filepath not found: " + fullPath)
		return false
	}

	l.logger.Info("Filepath found: " + fullPath)
	return true
}
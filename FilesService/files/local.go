package files

import (
	"fmt"
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

	exists, err := l.verifyExists(fp)
	if err != nil {
		return err
	}
	if !exists {
		l.logger.Warn(ErrNotFound.Error() + fp)
		return ErrNotFound
	}

	l.logger.Info("File found: " + path)
	return nil
}


func (l *Local) ReadFile(path string, w io.Writer) error {
	l.logger.Info("Reading the file: " + path)

	fp := l.fullPath(path)

	// check if requested file exists
	exists, err := l.verifyExists(fp)
	if err != nil {
		return err
	}
	if !exists {
		l.logger.Warn(ErrNotFound.Error() + fp)
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
	l.logger.Info("Saving the file: " + path)

	fp := l.fullPath(path)

	// check if the directory exists
	dir := filepath.Dir(fp)
	exists, err := l.verifyExists(dir)
	if err != nil {
		return err
	}
	if !exists {
		l.logger.Warn(ErrNotFound.Error() + fp)
		return ErrNotFound
	}

	// check if the file doesn't already exist
	exists, err = l.verifyExists(fp)
	if err != nil {
		return err
	}
	if exists {
		l.logger.Warn(ErrAlreadyExists.Error() + fp)
		return ErrAlreadyExists
	}

	// create the file
	writer, err := l.createFile(fp)
	if err != nil {
		return err
	}
	defer func() {
		writer.Close()
	}()

	// write the contents into the file
	err = l.writeFile(fp, writer, contents)
	if err != nil {
		return err
	}

	l.logger.Info("Saved the file: " + path)
	return nil
}

func (l *Local) OverwriteFile(path string, contents io.Reader) error {
	l.logger.Info("Updating the file: " + path)

	fp := l.fullPath(path)
	tfp := l.fullPath(path + "_tmp")

	// check if file exists
	exists, err := l.verifyExists(fp)
	if err != nil {
		return err
	}
	if !exists {
		l.logger.Warn(ErrNotFound.Error() + fp)
		return ErrNotFound
	}

	// create and write to the temp file
	writer, err := l.createFile(tfp)
	if err != nil {
		return err
	}
	defer func() {
		writer.Close()
		os.Remove(tfp)
	}()
	err = l.writeFile(tfp, writer, contents)
	if err != nil {
		return err
	}

	// replace the original file with the temporary file
	err = l.changeFilepath(tfp, fp)
    if err != nil {
        return err
    }

	l.logger.Info("Updated the file: " + path)
	return nil
}

func (l *Local) DeleteFile(path string) error {
	l.logger.Info("Deleting the file: " + path)

	fp := l.fullPath(path)

	// check if file exists
	exists, err := l.verifyExists(fp)
	if err != nil {
		return err
	}
	if !exists {
		l.logger.Warn(ErrNotFound.Error() + fp)
		return ErrNotFound
	}

	// check if filepath is file
	isFile, err := l.verifyIsFile(fp)
	if err != nil {
		return err
	}
	if !isFile {
		l.logger.Warn(ErrNotFile.Error() + fp)
		return ErrNotFile
	}

	// remove the file
	err = l.remove(fp)
	if err != nil {
		return err
	}

	l.logger.Info("Deleted the file: " + path)
	return nil
}

func (l *Local) MakeDirectory(path string) error {
	l.logger.Info("Creating directory: " + path)
	fp := l.fullPath(path)

	// check if the directory already exists
	exists, err := l.verifyExists(fp)
	if err != nil {
		return err
	}
	if exists {
		l.logger.Warn(ErrAlreadyExists.Error() + fp)
		return ErrAlreadyExists
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
	exists, err := l.verifyExists(fop)
	if err != nil {
		return err
	}
	if !exists {
		l.logger.Warn(ErrNotFound.Error() + fop)
		return ErrNotFound
	}

	// check if the directory doesn't exist
	exists, err = l.verifyExists(fnp)
	if err != nil {
		return err
	}
	if exists {
		l.logger.Warn(ErrAlreadyExists.Error() + fop)
		return ErrAlreadyExists
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
	exists, err := l.verifyExists(fop)
	if err != nil {
		return err
	}
	if !exists {
		l.logger.Warn(ErrNotFound.Error() + fop)
		return ErrNotFound
	}

	// check if the desired directory doesn't already exist
	exists, err = l.verifyExists(fnp)
	if err != nil {
		return err
	}
	if exists {
		l.logger.Warn(ErrAlreadyExists.Error() + fnp)
		return ErrAlreadyExists
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
	exists, err := l.verifyExists(fp)
	if err != nil {
		return err
	}
	if !exists {
		l.logger.Warn(ErrNotFound.Error() + fp)
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
	exists, err := l.verifyExists(fp)
	if err != nil {
		return err
	}
	if !exists {
		l.logger.Warn(ErrNotFound.Error() + fp)
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
	exists, err := l.verifyExists(fp)
	if err != nil {
		return err
	}
	if !exists {
		l.logger.Warn(ErrNotFound.Error() + fp)
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
	exists, err := l.verifyExists(fp)
	if err != nil {
		return nil, err
	}
	if !exists {
		l.logger.Warn(ErrNotFound.Error() + fp)
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
	exists, err := l.verifyExists(fp)
	if err != nil {
		return nil, err
	}
	if !exists {
		l.logger.Warn(ErrNotFound.Error() + fp)
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
func (l *Local) verifyExists(fullpath string) (bool, error) {
	l.logger.Info("Looking for file: " + fullpath)

	_, err := os.Stat(fullpath)
	if err != nil {
		if os.IsNotExist(err) {
			l.logger.Info("Filepath not found: " + fullpath)
			return false, nil
		}
		l.logger.Error(err.Error())
		return false, err
	}

	l.logger.Info("Filepath found: " + fullpath)
	return true, nil
}

func (l *Local) verifyIsFile(fullpath string) (bool, error) {
	l.logger.Info("Verifying the file: " + fullpath)

	file, err := os.Stat(fullpath)
	if err != nil {
		l.logger.Error(err.Error())
		return false, ErrStat
	}
	if file.IsDir() {
		l.logger.Warn(fmt.Sprintf("Filepath '%s' doesn't point to the file", fullpath))
		return false, ErrNotFile
	}

	l.logger.Info("Verified the file: " + fullpath)
	return true, nil
}

// Create the file under specified filepath
func (l *Local) createFile(fullpath string) (io.WriteCloser, error) {
	l.logger.Info("Creating the file: " + fullpath)

	f, err := os.Create(fullpath)
	if err != nil {
		l.logger.Error(err.Error())
		return nil, ErrFileCreate
	}

	l.logger.Info("Created the file: " + fullpath)
	return f, nil
}

// Writes reader contents into provided file writer
func (l *Local) writeFile(fullpath string, writer io.WriteCloser, contents io.Reader) error {
	l.logger.Info("Writing into the file: " + fullpath)

	// create a LimitedReader to limit file size
    limitedReader := &io.LimitedReader{
        R: contents,
        N: l.maxFileSize + 1,
    }

	// write the contents to the new file
	_, err := io.Copy(writer, limitedReader)
	if err != nil {
		l.logger.Error(err.Error())
		return ErrFileWrite
	}

	// check if filesize limit was reached
	if limitedReader.N == 0 {
		l.logger.Error("File size was exceeded")
		return ErrFileSizeExceeded
	}

	l.logger.Info("Finished writing into the file: " + fullpath)
	return nil
}

// remove requested filepath
func (l *Local) remove(fullPath string) error {
	l.logger.Info("Removing the filepath: " + fullPath)

	err := os.Remove(fullPath)
	if err != nil {
		l.logger.Error(err.Error())
		return ErrFileDelete
	}

	l.logger.Info("Removed the file: " + fullPath)
	return nil
}

// change filepath to the new provided string. Doesn't create directories.
func (l *Local) changeFilepath(old string, new string) error {
	l.logger.Info(fmt.Sprintf("Modifying filepath from: %s\nto: %s", old, new))

	err := os.Rename(old, new)
    if err != nil {
		l.logger.Error(err.Error())
        return ErrRename
    }

	l.logger.Info(fmt.Sprintf("Filepath changed from: %s\nto: %s", old, new))
	return nil
}
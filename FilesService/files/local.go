package files

import (
	"fmt"
	"io"
	"io/fs"
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

/*
	GENERAL
*/
func (l *Local) IfExists(path string) error {
	l.logger.Info("Looking for the filepath")

	fp := l.fullPath(path)

	exists, err := l.exists(fp)
	if err != nil {
		return err
	}
	if !exists {
		l.logger.Warn(ErrNotFound.Error() + fp)
		return ErrNotFound
	}

	l.logger.Info("Filepath found")
	return nil
}

func (l *Local) ReadFile(path string, w io.Writer) error {
	l.logger.Info("Reading the file: " + path)

	fp := l.fullPath(path)

	// check if requested file exists
	exists, err := l.exists(fp)
	if err != nil {
		return err
	}
	if !exists {
		l.logger.Warn(ErrNotFound.Error() + fp)
		return ErrNotFound
	}

	// read the file contents into the writer
	err = l.readFile(fp, w)
	if err != nil {
		return err
	}

	l.logger.Info("Done reading the file: " + path)
    return nil
}

func (l *Local) WriteFile(path string, contents io.Reader) error {
	l.logger.Info("Saving the file: " + path)

	fp := l.fullPath(path)

	// check if the directory exists
	dir := filepath.Dir(fp)
	exists, err := l.exists(dir)
	if err != nil {
		return err
	}
	if !exists {
		l.logger.Warn(ErrNotFound.Error() + fp)
		return ErrNotFound
	}

	// check if the file doesn't already exist
	exists, err = l.exists(fp)
	if err != nil {
		return err
	}
	if exists {
		l.logger.Warn(ErrAlreadyExists.Error() + fp)
		return ErrAlreadyExists
	}

	// create the file
	_, err = l.createFile(fp)
	if err != nil {
		return err
	}

	// write the contents into the file
	err = l.writeFile(fp, contents)
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
	exists, err := l.exists(fp)
	if err != nil {
		return err
	}
	if !exists {
		l.logger.Warn(ErrNotFound.Error() + fp)
		return ErrNotFound
	}

	// create and write to the temp file
	_, err = l.createFile(tfp)
	if err != nil {
		return err
	}
	err = l.writeFile(tfp, contents)
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
	exists, err := l.exists(fp)
	if err != nil {
		return err
	}
	if !exists {
		l.logger.Warn(ErrNotFound.Error() + fp)
		return ErrNotFound
	}

	// check if filepath is file
	isFile, err := l.isFile(fp)
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

func (l *Local) CreateDirectory(path string) error {
	l.logger.Info("Creating the directory: " + path)
	fp := l.fullPath(path)

	// check if the directory already exists
	exists, err := l.exists(fp)
	if err != nil {
		return err
	}
	if exists {
		l.logger.Warn(ErrAlreadyExists.Error() + fp)
		return ErrAlreadyExists
	}

	// create the directory
	err = l.createFilepath(fp)
	if err != nil {
		return err
	}

	l.logger.Info("Created the directory: " + path)
	return nil
}


func (l *Local) ChangeDirectory(oldPath string, newPath string) error {
	l.logger.Info("Moving the directory from: " + oldPath + " to: " + newPath)

	fop := l.fullPath(oldPath)
	fnp := l.fullPath(newPath)

	// check if requested directory exists
	exists, err := l.exists(fop)
	if err != nil {
		return err
	}
	if !exists {
		l.logger.Warn(ErrNotFound.Error() + fop)
		return ErrNotFound
	}

	// check if the target directory doesn't already exist
	exists, err = l.exists(fnp)
	if err != nil {
		return err
	}
	if exists {
		l.logger.Warn(ErrAlreadyExists.Error() + fnp)
		return ErrAlreadyExists
	}

	// move requested directory
	err = l.changeFilepath(fop, fnp)
	if err != nil {
		return err
	}

	l.logger.Info("Moved the directory from: " + oldPath + " to: " + newPath)
	return nil
}

func (l *Local) DeleteFiles(path string) error {
	l.logger.Info("Removing files from the directory: " + path)

	fp := l.fullPath(path)

	// check if directory exists
	exists, err := l.exists(fp)
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

	// read dir contents
	entries, err := dir.Readdir(-1)
	if err != nil {
		l.logger.Error(err.Error())
		return ErrDirectoryRead
	}

	// remove directory contents
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		err = os.Remove(fp + "/" + entry.Name())
        if err != nil {
			l.logger.Error(err.Error())
            return ErrDelete
        }
	}

	l.logger.Info("Removed files from the directory: " + path)
	return nil
}

func (l *Local) DeleteSubdirectories(path string) error {
	l.logger.Info("Removing subdirectories from the directory: " + path)

	fp := l.fullPath(path)

	// check if directory exists
	exists, err := l.exists(fp)
	if err != nil {
		return err
	}
	if !exists {
		l.logger.Warn(ErrNotFound.Error() + fp)
		return ErrNotFound
	}

	// read dir contents
	entries, err := l.readDirectory(fp)
	if err != nil {
		return err
	}

	// remove subdirectories
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		err = os.Remove(fp + "/" + entry.Name())
        if err != nil {
			l.logger.Error(err.Error())
            return ErrDelete
        }
	}
	
	l.logger.Info("Removed subdirectories from the directory: " + path)
	return nil
}


func (l *Local) DeleteDirectory(path string) error {
	l.logger.Info("Removing directory: " + path)

	fp := l.fullPath(path)

	// check if directory exists
	exists, err := l.exists(fp)
	if err != nil {
		return err
	}
	if !exists {
		l.logger.Warn(ErrNotFound.Error() + fp)
		return ErrNotFound
	}

	// read dir contents
	entries, err := l.readDirectory(fp)
	if err != nil {
		return err
	}
	
	// check if empty
	if len(entries) > 0 {
		l.logger.Warn(ErrDirNotEmpty.Error())
		return ErrDirNotEmpty
	}

	// remove the dir
	err = os.Remove(fp)
	if err != nil {
		l.logger.Error(err.Error())
		return ErrDelete
	}

	l.logger.Info("Removed directory: " + path)
	return nil
}


func (l *Local) ListFiles(path string) ([]string, error) {
	l.logger.Info("Reading files from directory: " + path)
	fp := l.fullPath(path)

	// check if the directory exists
	exists, err := l.exists(fp)
	if err != nil {
		return nil, err
	}
	if !exists {
		l.logger.Warn(ErrNotFound.Error() + fp)
		return nil, ErrNotFound
	}

	// check if filepath is a directory
	isFile, err := l.isFile(fp)
	if err != nil {
		return nil, err
	}
	if isFile {
		l.logger.Warn(ErrNotDirectory.Error())
		return nil, ErrNotDirectory
	}

	// read directory contents
	entries, err := l.readDirectory(fp)
	if err != nil {
		return nil, err
	}

	// save filenames
	filenames := make([]string, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() { // Check if the entry is a file
			filenames = append(filenames, entry.Name()) // Add the filename to the slice
		}
	}

	l.logger.Info("Finished reading files from directory: " + path)
	return filenames, nil
}


func (l *Local) ListDirectories(path string) ([]string, error) {
	l.logger.Info("Listing subdirectories in the directory: " + path)
	fp := l.fullPath(path)

	// check if the directory exists
	exists, err := l.exists(fp)
	if err != nil {
		return nil, err
	}
	if !exists {
		l.logger.Warn(ErrNotFound.Error() + fp)
		return nil, ErrNotFound
	}

	// check if filepath is a directory
	isFile, err := l.isFile(fp)
	if err != nil {
		return nil, err
	}
	if isFile {
		l.logger.Warn(ErrNotDirectory.Error())
		return nil, ErrNotDirectory
	}

	// read directory contents
	entries, err := l.readDirectory(fp)
	if err != nil {
		return nil, err
	}

	// save subdirectories list
	dirs := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() { // Check if the entry is a file
			dirs = append(dirs, entry.Name()) // Add the filename to the slice
		}
	}

	l.logger.Info("Listed subdirectories in the directory: " + path)
	return dirs, nil
}


/*
	FILEPATH
*/

// Changes filepath to the new provided string. Doesn't create directories.
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

// Removes requested filepath
func (l *Local) remove(fullPath string) error {
	l.logger.Info("Removing the filepath: " + fullPath)

	err := os.Remove(fullPath)
	if err != nil {
		l.logger.Error(err.Error())
		return ErrDelete
	}

	l.logger.Info("Removed the filepath: " + fullPath)
	return nil
}

/*
	FILE
*/

// Creates the file under specified filepath
func (l *Local) createFile(fullpath string) (io.WriteCloser, error) {
	l.logger.Info("Creating the file: " + fullpath)

	f, err := os.Create(fullpath)
	if err != nil {
		l.logger.Error(err.Error())
		return nil, ErrFileCreate
	}
	defer f.Close()

	l.logger.Info("Created the file: " + fullpath)
	return f, nil
}

// Reads file contents into provided reader
func (l *Local) readFile(fullpath string, writer io.Writer) error {
	l.logger.Info("Reading the file: " + fullpath)

	// open the file
    f, err := os.Open(fullpath)
    if err != nil {
		l.logger.Error(err.Error())
        return ErrFileRead
    }
    defer f.Close()

	// write the file contents into the writer
	_, err = io.Copy(writer, f)
    if err != nil {
		l.logger.Error(err.Error())
        return ErrFileRead
    }

	l.logger.Info("Finished reading the file: " + fullpath)
	return nil
}

// Writes reader contents into provided file writer
func (l *Local) writeFile(fullpath string, contents io.Reader) error {
	l.logger.Info("Writing into the file: " + fullpath)

	// create a LimitedReader to limit file size
    limitedReader := &io.LimitedReader{
        R: contents,
        N: l.maxFileSize + 1,
    }

	// open the file
	f, err := os.OpenFile(fullpath, os.O_WRONLY, 0)
    if err != nil {
		l.logger.Error(err.Error())
        return ErrFileRead
    }
    defer f.Close()

	// write the contents to the new file
	_, err = io.Copy(f, limitedReader)
	if err != nil {
		l.logger.Error(err.Error())
		return ErrFileWrite
	}

	// check if filesize limit was reached
	if limitedReader.N == 0 {
		l.logger.Error(ErrWriteSizeExceeded.Error())
		return ErrWriteSizeExceeded
	}

	l.logger.Info("Finished writing into the file: " + fullpath)
	return nil
}

// Verifies if provided filepath leads to a file
func (l *Local) isFile(fullpath string) (bool, error) {
	l.logger.Info("Verifying the file: " + fullpath)

	file, err := os.Stat(fullpath)
	if err != nil {
		l.logger.Error(err.Error())
		return false, ErrStat
	}
	if file.IsDir() {
		return false, nil
	}

	l.logger.Info("Verified the file: " + fullpath)
	return true, nil
}


/*
	DIRECTORY
*/

// Reads directory contents
func (l *Local) readDirectory(fullpath string) ([]fs.FileInfo, error) {
	l.logger.Info("Reading the directory: " + fullpath)

	// open the dir
	dir, err := os.Open(fullpath)
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

	l.logger.Info("Finished reading the directory: " + fullpath)

	return entries, nil
}
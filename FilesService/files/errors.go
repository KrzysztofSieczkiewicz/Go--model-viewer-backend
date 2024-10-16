package files

import (
	"errors"
)

// General
var ErrStat = errors.New("filepath couldn't be verified")

var ErrNotFound = errors.New("filepath was not found")

var ErrAlreadyExists = errors.New("filepath already exists")

var ErrRename = errors.New("cannot alter the filepath")

var ErrDelete = errors.New("filepath couldn't be deleted")

var ErrWriteSizeExceeded = errors.New("maximum file size was exceeded")


// Files
var ErrFileCreate = errors.New("unable to create file")

var ErrFileRead = errors.New("file couldn't be read")

var ErrFileWrite = errors.New("file couldn't be written into")

var ErrNotFile = errors.New("filepath is not a file")


// Directories
var ErrNotDirectory = errors.New("filepath doesn't end with directory")

var ErrDirectoryCreate = errors.New("unable to create directory")

var ErrDirectoryRead = errors.New("unable to open directory")

var ErrDirNotEmpty = errors.New("directory is not empty")

var ErrDirContainsFiles = errors.New("directory contains files")
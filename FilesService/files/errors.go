package files

import (
	"errors"
)

// Files
var ErrFileNotFound = errors.New("file was not found")

var ErrFileAlreadyExists = errors.New("file already exists")

var ErrFileCreate = errors.New("unable to create file")

var ErrFileRead = errors.New("file couldn't be read")

var ErrFileWrite = errors.New("file couldn't be written into")

var ErrFileDelete = errors.New("fiel couldn't be deleted")

var ErrFileStat = errors.New("file couldn't be verified")

var ErrFileReplace = errors.New("file couldn't be replaced")

var ErrFileSizeExceeded = errors.New("maximum file size was exceeded")


// Directories
var ErrDirectoryNotFound = errors.New("directory doesn't exist")

var ErrDirectoryCreate = errors.New("unable to create directory")


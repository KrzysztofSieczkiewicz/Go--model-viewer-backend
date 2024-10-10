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

var ErrFileDelete = errors.New("file couldn't be deleted")

var ErrFileStat = errors.New("file couldn't be verified")

var ErrFileReplace = errors.New("file couldn't be replaced")

var ErrFileSizeExceeded = errors.New("maximum file size was exceeded")


// Directories
var ErrDirectoryNotFound = errors.New("directory doesn't exist")

var ErrNotDirectory = errors.New("filepath doesn't end with directory")

var ErrDirectoryAlreadyExists = errors.New("directory already exists")

var ErrDirectoryCreate = errors.New("unable to create directory")

var ErrDirectoryRead = errors.New("unable to open directory")

var ErrDirectoryDelete = errors.New("directory couldn't be deleted")

var ErrDirectoryStat = errors.New("directory couldn't be verified")

var ErrDirectoryRename = errors.New("unable to rename directory")

var ErrDirectoryMove = errors.New("unable to move directory")

var ErrDirectoryNotEmpty = errors.New("directory is not empty")

var ErrDirectoryNonDirectoryFound = errors.New("directory contains files that are not files")


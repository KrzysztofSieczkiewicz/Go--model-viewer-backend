package files

import (
	"io"

	"github.com/KrzysztofSieczkiewicz/go--model-viewer-backend/FilesService/models"
)

// Defines behavior for file operations.
// Different implementations might allow for local/cloud storage
type Storage interface {
/*
	GENERAL
*/
	// Checks if filepath can be found in the filesystem
	IfExists(path string) error

	// Checks if filepath is a category
//	IsCategory(path string) error

/*
	FILES
*/

	// Create and write a file under provided path. Returns an error if file already exists
	WriteFile(path string, file io.Reader) error

	// Overwrites provided file using temp file. Fails if requested file doesn't exist
	OverwriteFile(path string, file io.Reader) error

	// Deletes file under provided path. Returns error if file doesn't exist
	DeleteFile(path string) error


	// Creates requested directory or dir structure, returns an error if path already exists
	CreateDirectory(path string) error

	// Changes dir name and path. If old and new paths are in different directories functions as move. Doesn't create new directories
	ChangeDirectory(oldPath string, newPath string) error

	// Remove files stored in provided filepath. Omits subdirectories
	DeleteFiles(path string) error

	// Remove subdirectories in provided filepath. Omits files
	DeleteSubdirectories(path string) error
	
	// Deletes directory. Fails if directory is not empty
	DeleteDirectory(path string) error

	// Lists files in the directory
	ListFiles(path string) ([]string, error)

	// List subdirectories in the directory
	ListDirectories(path string) ([]string, error)


	/*
		NEW
	*/
	// Assets
	CheckAsset(asset models.Asset) error

	GetAsset(filepath string, w io.Writer) error

	AddAsset(asset models.Asset, r io.Reader) error

	OverwriteAsset(asset models.Asset, r io.Reader) error

	UpdateAsset(asset models.Asset, newAsset models.Asset) error

	DeleteAsset(asset models.Asset) error

	// Collections
	ListCollectionContents(category string, id string) ([]models.CollectionContent, error) 

	CreateCollection(category string, id string) error

	UpdateCollection(category string, id string, newCategory string, newId string) error

	DeleteCollection(category string, id string) error

	// Categories
	ListCategoryContents(path string) ([]models.CollectionContent, error)

	CreateCategory(path string) error

	UpdateCategory(path string, name string) error

	DeleteCategory(path string) error
}
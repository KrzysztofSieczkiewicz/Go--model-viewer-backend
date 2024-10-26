package files

import (
	"io"

	"github.com/KrzysztofSieczkiewicz/go--model-viewer-backend/FilesService/models"
)

// Defines behavior for file operations.
// Different implementations might allow for local/cloud storage
type Storage interface {

	// Assets
	CheckAsset(asset models.Asset) error

	GetAsset(filepath string, w io.Writer) error

	AddAsset(asset models.Asset, r io.Reader) error

	OverwriteAsset(asset models.Asset, r io.Reader) error

	UpdateAsset(asset models.Asset, newAsset models.Asset) error

	DeleteAsset(asset models.Asset) error

	// Collections
	ListCollectionContents(collection *models.AssetsCollection) ([]models.CollectionContent, error)

	CreateCollection(collection *models.AssetsCollection) error

	UpdateCollection(collection *models.AssetsCollection, newCollection *models.AssetsCollection) error

	DeleteCollection(collection *models.AssetsCollection) error

	// Categories
	ListCategoryContents(category *models.Category) ([]models.CollectionContent, error)

	CreateCategory(category *models.Category) error

	UpdateCategory(category *models.Category, newCategory *models.Category) error

	DeleteCategory(category *models.Category) error
}
package files_test

import (
	"io"
	"log/slog"
	"testing"

	"github.com/KrzysztofSieczkiewicz/go--model-viewer-backend/FilesService/files"
	"github.com/KrzysztofSieczkiewicz/go--model-viewer-backend/FilesService/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper function to create Local instance for testing
func setupLocal(t *testing.T, maxFileSizeMB int) (*files.Local) {
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))

	tempDir := t.TempDir()

	local, err := files.NewLocal(tempDir, maxFileSizeMB, logger)
	assert.NoError(t, err)
	return local
}

func setupCategory(t *testing.T, storage *files.Local) *models.Category{
	category := &models.Category{
		Path: "/test/category",
	}
	err := storage.CreateCategory(category)
    require.NoError(t, err)

	return category
}

func TestNewLocal(t *testing.T) {
	localStorage := setupLocal(t, 1)

	assert.NotNil(t, localStorage)
	assert.Equal(t, int64(1*1000*1024), localStorage.MaxFileSize())
}

func TestAddCategory(t *testing.T) {
	localStorage := setupLocal(t, 1)

	category := &models.Category{
		Path: "/test/path",
	}

	err := localStorage.CreateCategory(category)
	assert.NoError(t, err)

	_, err = localStorage.ListCategoryContents(category)
	assert.NoError(t, err)
}

func TestAddCategory_IllegalPath(t *testing.T) {
	localStorage := setupLocal(t, 1)

	category := &models.Category{
		Path: "/",
	}
	err := localStorage.CreateCategory(category)
	assert.Error(t, err)

	category = &models.Category{
		Path: "",
	}
	err = localStorage.CreateCategory(category)
	assert.Error(t, err)

	category = &models.Category{
		Path: "//",
	}
	err = localStorage.CreateCategory(category)
	assert.Error(t, err)

	category = &models.Category{
		Path: "/path/..",
	}
	err = localStorage.CreateCategory(category)
	assert.Error(t, err)
}

func TestUpdateCategory(t *testing.T) {
	localStorage := setupLocal(t, 1)

	category := &models.Category{
		Path: "/test/path",
	}
	err := localStorage.CreateCategory(category)
    require.NoError(t, err)

	newCategory := &models.Category{
		Path: "/test/path2",
	}
	err = localStorage.UpdateCategory(category, newCategory)
    assert.NoError(t, err)

	_, err = localStorage.ListCategoryContents(newCategory)
	assert.NoError(t, err)

	_, err = localStorage.ListCategoryContents(category)
	assert.Error(t, err)
}

func TestUpdateCategory_InvalidPath(t *testing.T) {
	localStorage := setupLocal(t, 1)

	category := &models.Category{
		Path: "/_test/_path",
	}
	err := localStorage.CreateCategory(category)
    require.NoError(t, err)

	newCategory := &models.Category{
		Path: "/_test/\\path2/",
	}
	err = localStorage.UpdateCategory(category, newCategory)
    assert.Error(t, err)
	_, err = localStorage.ListCategoryContents(category)
	assert.NoError(t, err)
	_, err = localStorage.ListCategoryContents(newCategory)
	assert.Error(t, err)

	newCategory = &models.Category{
		Path: "/test/path2/",
	}
	err = localStorage.UpdateCategory(category, newCategory)
    assert.Error(t, err)
	_, err = localStorage.ListCategoryContents(category)
	assert.NoError(t, err)
	_, err = localStorage.ListCategoryContents(newCategory)
	assert.Error(t, err)

	newCategory = &models.Category{
		Path: "/_test/path2/",
	}
	err = localStorage.UpdateCategory(category, newCategory)
    assert.Error(t, err)
	_, err = localStorage.ListCategoryContents(category)
	assert.NoError(t, err)
	_, err = localStorage.ListCategoryContents(newCategory)
	assert.Error(t, err)

	newCategory = &models.Category{
		Path: "\"",
	}
	err = localStorage.UpdateCategory(newCategory, newCategory)
    assert.Error(t, err)
	_, err = localStorage.ListCategoryContents(category)
	assert.NoError(t, err)

	newCategory3 := &models.Category{
		Path: "",
	}
	err = localStorage.UpdateCategory(newCategory, newCategory3)
    assert.Error(t, err)
	_, err = localStorage.ListCategoryContents(category)
	assert.NoError(t, err)
}

func TestDeleteCategory(t *testing.T) {
	localStorage := setupLocal(t, 1)

	category := &models.Category{
		Path: "/test/path",
	}
	err := localStorage.CreateCategory(category)
    require.NoError(t, err)

	err = localStorage.DeleteCategory(category)
	assert.NoError(t, err)

	_, err = localStorage.ListCategoryContents(category)
	assert.Error(t, err)
}

func TestDeleteCategory_IllegalPath(t *testing.T) {
	localStorage := setupLocal(t, 1)

	category := &models.Category{
		Path: ".",
	}
	err := localStorage.DeleteCategory(category)
	assert.Error(t, err)
	
	category = &models.Category{
		Path: "..",
	}
	err = localStorage.DeleteCategory(category)
	assert.Error(t, err)

	category = &models.Category{
		Path: "./..",
	}
	err = localStorage.DeleteCategory(category)
	assert.Error(t, err)

	category = &models.Category{
		Path: "",
	}
	err = localStorage.DeleteCategory(category)
	assert.Error(t, err)
}

func TestDeleteCategory_NotEmpty(t *testing.T) {
	localStorage := setupLocal(t, 1)

	category := &models.Category{
		Path: "/test/path",
	}
	err := localStorage.CreateCategory(category)
    require.NoError(t, err)

	collection := &models.Collection{
		Category: *category,
		ID: "testCategory",
	}
	err = localStorage.CreateCollection(collection)
	require.NoError(t, err)

	contents, err := localStorage.ListCategoryContents(category)
	require.NoError(t, err)
	assert.NotEmpty(t, contents)

	err = localStorage.DeleteCategory(category)
	assert.Error(t, err)

	contents, err = localStorage.ListCategoryContents(category)
	assert.NoError(t, err)
	assert.NotEmpty(t, contents)
}

func TestCreateCollection(t *testing.T) {
	localStorage := setupLocal(t, 1)
	category := setupCategory(t, localStorage)

	collection := &models.Collection{
		Category: *category,
		ID: "testCollection",
	}

	err := localStorage.CreateCollection(collection)
	assert.NoError(t, err)

	contents, err := localStorage.ListCollectionContents(collection)
	assert.NoError(t, err)
	assert.NotEmpty(t, contents)
}

func TestCreateCollection_IllegalID(t *testing.T) {
	localStorage := setupLocal(t, 1)
	category := setupCategory(t, localStorage)

	collection := &models.Collection{
		Category: *category,
		ID: ".",
	}
	err := localStorage.CreateCollection(collection)
	assert.Error(t, err)

	collection = &models.Collection{
		Category: *category,
		ID: "..",
	}
	err = localStorage.CreateCollection(collection)
	assert.Error(t, err)

	collection = &models.Collection{
		Category: *category,
		ID: "./..",
	}
	err = localStorage.CreateCollection(collection)
	assert.Error(t, err)

	collection = &models.Collection{
		Category: *category,
		ID: "",
	}
	err = localStorage.CreateCollection(collection)
	assert.Error(t, err)

	collection = &models.Collection{
		Category: *category,
		ID: "_testID",
	}
	err = localStorage.CreateCollection(collection)
	assert.Error(t, err)
}

func TestCreateCollection_NoSuchCategory(t *testing.T) {
	localStorage := setupLocal(t, 1)

	category := &models.Category{
		Path: "/new/category/path",
	}

	collection := &models.Collection{
		Category: *category,
		ID: "testCollection",
	}

	err := localStorage.CreateCollection(collection)
	assert.Error(t, err)

	_, err = localStorage.ListCollectionContents(collection)
	assert.Error(t, err)
}


func TestUpdateCollection(t *testing.T) {
	localStorage := setupLocal(t, 1)
	category := setupCategory(t, localStorage)

	collection := &models.Collection{
		Category: *category,
		ID: "testCollection",
	}
	err := localStorage.CreateCollection(collection)
	require.NoError(t, err)

	newCollection := &models.Collection{
		Category: *category,
		ID: "testCollection2",
	}
	err = localStorage.UpdateCollection(collection, newCollection)
	assert.NoError(t, err)

	_, err = localStorage.ListCollectionContents(collection)
	assert.Error(t, err)
	_, err = localStorage.ListCollectionContents(newCollection)
	assert.NoError(t, err)
}

func TestUpdateCollection_changeCategory(t *testing.T) {
	localStorage := setupLocal(t, 1)
	category := setupCategory(t, localStorage)

	collection := &models.Collection{
		Category: *category,
		ID: "testCollection",
	}
	err := localStorage.CreateCollection(collection)
	require.NoError(t, err)

	newCategory := &models.Category{
		Path: "/new/category/path",
	}
	err = localStorage.CreateCategory(newCategory)
	require.NoError(t, err)

	newCollection := &models.Collection{
		Category: *newCategory,
		ID: "testCollection2",
	}
	err = localStorage.UpdateCollection(collection, newCollection)
	assert.NoError(t, err)

	_, err = localStorage.ListCollectionContents(collection)
	assert.Error(t, err)
	_, err = localStorage.ListCollectionContents(newCollection)
	assert.NoError(t, err)
}

func TestUpdateCollection_IllegalID(t *testing.T) {
	localStorage := setupLocal(t, 1)
	category := setupCategory(t, localStorage)

	collection := &models.Collection{
		Category: *category,
		ID: "testCollection",
	}
	err := localStorage.CreateCollection(collection)
	require.NoError(t, err)

	newCollection := &models.Collection{
		Category: *category,
		ID: ".",
	}
	err = localStorage.UpdateCollection(collection, newCollection)
	assert.Error(t, err)
	_, err = localStorage.ListCollectionContents(collection)
	assert.NoError(t, err)
	_, err = localStorage.ListCollectionContents(newCollection)
	assert.Error(t, err)

	newCollection = &models.Collection{
		Category: *category,
		ID: "..",
	}
	err = localStorage.UpdateCollection(collection, newCollection)
	assert.Error(t, err)
	_, err = localStorage.ListCollectionContents(collection)
	assert.NoError(t, err)
	_, err = localStorage.ListCollectionContents(newCollection)
	assert.Error(t, err)

	newCollection = &models.Collection{
		Category: *category,
		ID: "./..",
	}
	err = localStorage.UpdateCollection(collection, newCollection)
	assert.Error(t, err)
	_, err = localStorage.ListCollectionContents(collection)
	assert.NoError(t, err)
	_, err = localStorage.ListCollectionContents(newCollection)
	assert.Error(t, err)

	newCollection = &models.Collection{
		Category: *category,
		ID: "",
	}
	err = localStorage.UpdateCollection(collection, newCollection)
	assert.Error(t, err)
	_, err = localStorage.ListCollectionContents(collection)
	assert.NoError(t, err)
	_, err = localStorage.ListCollectionContents(newCollection)
	assert.Error(t, err)

	newCollection = &models.Collection{
		Category: *category,
		ID: "test/.",
	}
	err = localStorage.UpdateCollection(collection, newCollection)
	assert.Error(t, err)
	_, err = localStorage.ListCollectionContents(collection)
	assert.NoError(t, err)
	_, err = localStorage.ListCollectionContents(newCollection)
	assert.Error(t, err)

	newCollection = &models.Collection{
		Category: *category,
		ID: "test/",
	}
	err = localStorage.UpdateCollection(collection, newCollection)
	assert.Error(t, err)
	_, err = localStorage.ListCollectionContents(collection)
	assert.NoError(t, err)
	_, err = localStorage.ListCollectionContents(newCollection)
	assert.Error(t, err)

	newCollection = &models.Collection{
		Category: *category,
		ID: "_testID",
	}
	err = localStorage.UpdateCollection(collection, newCollection)
	assert.Error(t, err)
	_, err = localStorage.ListCollectionContents(collection)
	assert.NoError(t, err)
	_, err = localStorage.ListCollectionContents(newCollection)
	assert.Error(t, err)
}

func TestUpdateCollection_noSuchCategory(t *testing.T) {
	localStorage := setupLocal(t, 1)
	category := setupCategory(t, localStorage)

	collection := &models.Collection{
		Category: *category,
		ID: "testCollection",
	}
	err := localStorage.CreateCollection(collection)
	require.NoError(t, err)

	newCategory := &models.Category{
		Path: "/new/category/path",
	}
	newCollection := &models.Collection{
		Category: *newCategory,
		ID: "testCollection2",
	}
	err = localStorage.UpdateCollection(collection, newCollection)
	assert.Error(t, err)

	_, err = localStorage.ListCollectionContents(collection)
	assert.NoError(t, err)
	_, err = localStorage.ListCollectionContents(newCollection)
	assert.Error(t, err)
}

func TestDeleteCollection(t *testing.T) {
	localStorage := setupLocal(t, 1)
	category := setupCategory(t, localStorage)

	collection := &models.Collection{
		Category: *category,
		ID: "testCollection",
	}
	err := localStorage.CreateCollection(collection)
	require.NoError(t, err)

	err = localStorage.DeleteCollection(collection)
	assert.NoError(t, err)

	_, err = localStorage.ListCollectionContents(collection)
	assert.Error(t, err)
}

func TestDeleteCollection_IllegalCollectionID(t *testing.T) {
	localStorage := setupLocal(t, 1)
	category := setupCategory(t, localStorage)

	collection := &models.Collection{
		Category: *category,
		ID: ".",
	}
	err := localStorage.DeleteCollection(collection)
	assert.Error(t, err)

	collection = &models.Collection{
		Category: *category,
		ID: "..",
	}
	err = localStorage.DeleteCollection(collection)
	assert.Error(t, err)

	collection = &models.Collection{
		Category: *category,
		ID: "./..",
	}
	err = localStorage.DeleteCollection(collection)
	assert.Error(t, err)

	collection = &models.Collection{
		Category: *category,
		ID: "",
	}
	err = localStorage.DeleteCollection(collection)
	assert.Error(t, err)
}

func TestDeleteCollection_NoCollectionID(t *testing.T) {
	localStorage := setupLocal(t, 1)
	category := setupCategory(t, localStorage)

	collection := &models.Collection{
		Category: *category,
		ID: "",
	}

	err := localStorage.DeleteCollection(collection)
	assert.Error(t, err)
}
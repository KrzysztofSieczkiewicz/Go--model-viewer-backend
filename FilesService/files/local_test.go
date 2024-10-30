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

	newCategory2 := &models.Category{
		Path: "\"",
	}
	err = localStorage.UpdateCategory(newCategory, newCategory2)
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
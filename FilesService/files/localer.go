package files

import (
	"io"
	"os"
	"path/filepath"

	"github.com/KrzysztofSieczkiewicz/go--model-viewer-backend/FilesService/models"
)

/*
	ASSET
*/
func (l *Local) CheckAsset(asset models.Asset) error {
	l.logger.Info("Checking the asset")

	p := asset.ConstructFilepath()
	fp := l.fullPath(p)

	// check if requested file exists
	exists, err := l.exists(fp)
	if err != nil {
		return err
	}
	if !exists {
		l.logger.Warn(ErrNotFound.Error())
		return ErrNotFound
	}

	return nil
}

func (l *Local) GetAsset(filepath string, w io.Writer) error {
	l.logger.Info("Reading the asset")

	fp := l.fullPath(filepath)

	// check if requested file exists
	exists, err := l.exists(fp)
	if err != nil {
		return err
	}
	if !exists {
		l.logger.Warn(ErrNotFound.Error())
		return ErrNotFound
	}

	// read the file contents into the writer
	err = l.readFile(fp, w)
	if err != nil {
		return err
	}

	l.logger.Info("Finished reading the asset")
    return nil
}

func (l *Local) AddAsset(asset models.Asset, r io.Reader) error {
	l.logger.Info("Writing the asset")

	p := asset.ConstructFilepath()
	fp := l.fullPath(p)

	// check if the directory exists
	dir := filepath.Dir(fp)
	exists, err := l.exists(dir)
	if err != nil {
		return err
	}
	if !exists {
		l.logger.Warn(ErrNotFound.Error())
		return ErrNotFound
	}

	// check if requested file doesn't already exist
	exists, err = l.exists(fp)
	if err != nil {
		return err
	}
	if exists {
		l.logger.Warn(ErrAlreadyExists.Error())
		return ErrAlreadyExists
	}

	// create and write to the file
	_, err = l.createFile(fp)
	if err != nil {
		return err
	}
	err = l.writeFile(fp, r)
	if err != nil {
		return err
	}

	l.logger.Info("Finished writing the asset")
	return nil
}

func (l *Local) OverwriteAsset(asset models.Asset, r io.Reader) error {
	l.logger.Info("Updating the asset")

	p := asset.ConstructFilepath()
	fp := l.fullPath(p)
	tfp := fp + "_tmp"

	// check if file exists
	exists, err := l.exists(fp)
	if err != nil {
		return err
	}
	if !exists {
		l.logger.Warn(ErrNotFound.Error())
		return ErrNotFound
	}

	// create and write to the temp file
	_, err = l.createFile(tfp)
	if err != nil {
		return err
	}
	err = l.writeFile(tfp, r)
	if err != nil {
		return err
	}

	// replace the original file with the temporary file
	err = l.changeFilepath(tfp, fp)
    if err != nil {
        return err
    }

	l.logger.Info("Updated the asset")
	return nil
}

func (l *Local) UpdateAsset(asset models.Asset, newAsset models.Asset) error {
	l.logger.Info("Rename the asset")

	p := asset.ConstructFilepath()
	fp := l.fullPath(p)
	np := newAsset.ConstructFilepath()
	nfp := l.fullPath(np)

	// check if file exists
	exists, err := l.exists(fp)
	if err != nil {
		return err
	}
	if !exists {
		l.logger.Warn(ErrNotFound.Error())
		return ErrNotFound
	}

	// check if target file doesn't already exist
	exists, err = l.exists(fp)
	if err != nil {
		return err
	}
	if exists {
		l.logger.Warn(ErrAlreadyExists.Error())
		return ErrAlreadyExists
	}

	// rename the file
	err = l.changeFilepath(fp, nfp)
    if err != nil {
        return err
    }

	l.logger.Info("Renamed the asset")
	return nil
}

func (l *Local) DeleteAsset(asset models.Asset) error {
	l.logger.Info("Removing the asset")

	p := asset.ConstructFilepath()
	fp := l.fullPath(p)

	// check if file exists
	exists, err := l.exists(fp)
	if err != nil {
		return err
	}
	if !exists {
		l.logger.Warn(ErrNotFound.Error())
		return ErrNotFound
	}

	// check if filepath is file
	isFile, err := l.isFile(fp)
	if err != nil {
		return err
	}
	if !isFile {
		l.logger.Warn(ErrNotFile.Error())
		return ErrNotFile
	}

	// remove the file
	err = l.remove(fp)
	if err != nil {
		return err
	}

	l.logger.Info("Removed the file")
	return nil
}

/*
	COLLECTION
*/
func (l *Local) ListCollectionContents(collection *models.AssetsCollection) ([]models.CollectionContent, error) {
	l.logger.Info("Listing the collection contents")

	cp := collection.ConstructCollectionPath()
	fp := l.fullPath(cp)

	// check if collection exists
	exists, err := l.exists(fp)
	if err != nil {
		return nil, err
	}
	if !exists {
		l.logger.Warn(ErrNotFound.Error())
		return nil, ErrNotFound
	}

	contents, err := l.listContents(fp)
	if err != nil {
		return nil, err
	}

	l.logger.Info("Listed the collection contents")
	return contents, nil
}

func (l *Local) CreateCollection(collection *models.AssetsCollection) error {
	l.logger.Info("Creating the collection")

	// full path
	p := collection.ConstructCollectionPath()
	fp := l.fullPath(p)

	// check if category exists
	exists, err := l.exists(filepath.Dir(fp))
	if err != nil {
		return err
	}
	if !exists {
		l.logger.Warn(ErrNotFound.Error())
		return ErrNotFound
	}

	// check if target already exists
	exists, err = l.exists(fp)
	if err != nil {
		return err
	}
	if exists {
		l.logger.Warn(ErrAlreadyExists.Error())
		return ErrAlreadyExists
	}

	// create directory
	err = l.createFilepath(fp)
	if err != nil {
		return err
	}

	l.logger.Info("Created the collection")
	return nil
}

func (l *Local) UpdateCollection(collection *models.AssetsCollection, newCollection *models.AssetsCollection) error {
	l.logger.Info("Renaming the collection")

	// current path
	cp := collection.ConstructCollectionPath()
	fp := l.fullPath(cp)

	// desired path
	np := newCollection.ConstructCollectionPath()
	nfp := l.fullPath(np)

	// check if collection exists
	exists, err := l.exists(fp)
	if err != nil {
		return err
	}
	if !exists {
		l.logger.Warn(ErrNotFound.Error())
		return ErrNotFound
	}

	// check if target collection doesn't exist
	exists, err = l.exists(fp)
	if err != nil {
		return err
	}
	if exists {
		l.logger.Warn(ErrAlreadyExists.Error())
		return ErrAlreadyExists
	}

	// update the collection
	err = l.changeFilepath(fp, nfp)
	if err != nil {
		return err
	}

	l.logger.Info("Renamed the collection")
	return nil
}

func (l *Local) DeleteCollection(collection *models.AssetsCollection) error {
	l.logger.Info("Removing the collection")

	cp := collection.ConstructCollectionPath()
	fp := l.fullPath(cp)

	// check if collection exists
	exists, err := l.exists(fp)
	if err != nil {
		return err
	}
	if !exists {
		l.logger.Warn(ErrNotFound.Error())
		return ErrNotFound
	}

	// remove the collection
	err = l.remove(fp)
	if err != nil {
		return err
	}

	l.logger.Info("Removed the collection")
	return nil
}


/*
	CATEGORY
*/
func (l *Local) ListCategoryContents(category *models.Category) ([]models.CollectionContent, error) {
	l.logger.Info("Listing the category contents")

	cp := category.ConstructCategoryPath()
	fp := l.fullPath(cp)

	// check if collection exists
	exists, err := l.exists(fp)
	if err != nil {
		return nil, err
	}
	if !exists {
		l.logger.Warn(ErrNotFound.Error())
		return nil, ErrNotFound
	}

	// list and return contents
	contents, err := l.listContents(fp)
	if err != nil {
		return nil, err
	}

	l.logger.Info("Listed the category contents")
	return contents, nil
}

func (l *Local) CreateCategory(category *models.Category) error {
	l.logger.Info("Creating the category")

	cp := category.ConstructCategoryPath()
	fp := l.fullPath(cp)

	// check if the directory already exists
	exists, err := l.exists(fp)
	if err != nil {
		return err
	}
	if exists {
		l.logger.Warn(ErrAlreadyExists.Error())
		return ErrAlreadyExists
	}

	// create new filepath
	err = l.createFilepath(fp)
	if err != nil {
		return err
	}

	l.logger.Info("Created the category")
	return nil
}

func (l *Local) UpdateCategory(category *models.Category, newCategory *models.Category) error {
	l.logger.Info("Updating the category")

	// construct filepath for the current path
	ocp := category.ConstructCategoryPath()
	ofp := l.fullPath(ocp)

	// construct filepath for the new path
	ncp := newCategory.ConstructCategoryPath()
	nfp := l.fullPath(ncp)

	// check if requested category exists
	exists, err := l.exists(ofp)
	if err != nil {
		return err
	}
	if !exists {
		l.logger.Warn(ErrNotFound.Error())
		return ErrNotFound
	}

	// check if target category already doesn't already exist
	exists, err = l.exists(nfp)
	if err != nil {
		return err
	}
	if exists {
		l.logger.Warn(ErrAlreadyExists.Error())
		return ErrAlreadyExists
	}

	// rename requested directory
	err = l.changeFilepath(ofp, nfp)
	if err != nil {
		return err
	}

	l.logger.Info("Updated the category")
	return nil
}

func (l *Local) DeleteCategory(category *models.Category) error {
	l.logger.Info("Removing the category")

	cp := category.ConstructCategoryPath()
	fp := l.fullPath(cp)

	// check if category exists
	exists, err := l.exists(fp)
	if err != nil {
		return err
	}
	if !exists {
		l.logger.Warn(ErrNotFound.Error())
		return ErrNotFound
	}

	// read category contents and check if it is empty
	entries, err := l.readDirectory(fp)
	if err != nil {
		return err
	}
	if len(entries) > 0 {
		l.logger.Warn(ErrDirNotEmpty.Error())
		return ErrDirNotEmpty
	}

	// remove the category
	err = l.remove(fp)
	if err != nil {
		return err
	}

	l.logger.Info("Removed the category")
	return nil
}


/*
	INTERNAL HELPERS
*/

// Returns the absolute path from the relative path
func (l *Local) fullPath(path string) string {
	return filepath.Join(l.basePath, path)
}

// Creates directories structure matching requested filepath
func (l *Local) createFilepath(fullpath string) error {
	l.logger.Info("Creating filepath: " + fullpath)
	
	err := os.MkdirAll(fullpath, 0755)
	if err != nil {
		l.logger.Error(err.Error())
		return ErrDirectoryCreate
	}

	return nil
}

// Verifies if filepath exists in the filesystem
func (l *Local) exists(fullpath string) (bool, error) {
	l.logger.Info("Checking filepath: " + fullpath)

	_, err := os.Stat(fullpath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		l.logger.Error(err.Error())
		return false, err
	}

	return true, nil
}

// Removes requested filepath
func (l *Local) remove(fullPath string) error {
	l.logger.Info("Removing the filepath: " + fullPath)

	err := os.Remove(fullPath)
	if err != nil {
		l.logger.Error(err.Error())
		return ErrDelete
	}

	return nil
}

// Lists directory contents
func (l *Local) listContents(fullpath string)  ([]models.CollectionContent, error) {
	l.logger.Info("Listing the contents: " + fullpath) 

	// read the directory
	entries, err := l.readDirectory(fullpath)
	if err != nil {
		return nil, err
	}

	// check all entries and assign types
	contents := make([]models.CollectionContent, 0, len(entries))
	var fileType models.FileType
	for _, entry := range entries {
		if entry.IsDir() {
			fileType = models.FileTypeDirectory
		} else {
			fileType = models.FileTypeFile
		}

		contents = append(
			contents, 
			models.CollectionContent{
				Filename: entry.Name(),
				FileType: fileType,
			},
		)
	}

	return contents, nil
}
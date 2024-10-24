package files

import (
	"io"
	"os"
	"path/filepath"
	"strings"
)

/*
	FILE
*/
func (l *Local) ReadFile_(category string, collection string, filename string, w io.Writer) error {
	l.logger.Info("Reading the file")

	cp := l.constructCategoryPath(category)
	p := filepath.Join(cp, collection, filename)
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

	// read the file contents into the writer
	err = l.readFile(fp, w)
	if err != nil {
		return err
	}

	l.logger.Info("Finished reading the file")
    return nil
}

func (l *Local) WriteFile_(category string, collection string, filename string, r io.Reader) error {
	l.logger.Info("Writing the file")

	cp := l.constructCategoryPath(category)
	p := filepath.Join(cp, collection, filename)
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

	l.logger.Info("Finished writing the file")
	return nil
}

func (l *Local) OverwriteFile_(category string, collection string, filename string, r io.Reader) error {
	l.logger.Info("Updating the file")

	cp := l.constructCategoryPath(category)
	p := filepath.Join(cp, collection, filename)
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

	l.logger.Info("Updated the file")
	return nil
}

func (l *Local) DeleteFile_(category string, collection string, filename string) error {
	l.logger.Info("Deleting the file")

	cp := l.constructCategoryPath(category)
	p := filepath.Join(cp, collection, filename)
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

	l.logger.Info("Deleted the file")
	return nil
}

/*
	COLLECTION
*/
func (l *Local) VerifyCollection(category string, id string) error {
	l.logger.Info("Verifying the collection")

	cp := l.constructCategoryPath(category)
	p := filepath.Join(cp, id)
	fp := l.fullPath(p)

	// check if target exists
	exists, err := l.exists(fp)
	if err != nil {
		return err
	}
	if !exists {
		l.logger.Warn(ErrNotFound.Error())
		return ErrNotFound
	}

	// verify the path is correct
	is := l.verifyCollectionPath(fp)
	if !is {
		return ErrNotCollection
	}

	l.logger.Info("Verified the collection")
	return nil
}

func (l *Local) CreateCollection(category string, id string) error {
	l.logger.Info("Creating the collection")

	cp := l.constructCategoryPath(category)
	cfp := l.fullPath(cp)
	fp := filepath.Join(cfp, id)

	// check if category exists
	exists, err := l.exists(cfp)
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

func (l *Local) UpdateCollection(category string, id string, newCategory string, newId string) error {
	l.logger.Info("Renaming the collection")

	// current path
	cp := l.constructCategoryPath(category)
	p := filepath.Join(cp, id)
	fp := l.fullPath(p)

	// desired path
	ncp := l.constructCategoryPath(newCategory)
	np := filepath.Join(ncp, newId)
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

func (l *Local) DeleteCollection(category string, id string) error {
	l.logger.Info("Removing the collection")

	cp := l.constructCategoryPath(category)
	p := filepath.Join(cp, id)
	fp := l.fullPath(p)

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
func (l *Local) VerifyCategoryPath(path string) error {
	l.logger.Info("Verifying the category")

	is := l.verifyCategoryPath(path)
	if !is {
		return ErrNotCategory
	}

	l.logger.Info("Verified the category")
	return nil
}

func (l *Local) CreateCategory(path string) error {
	l.logger.Info("Creating the category")

	cp := l.constructCategoryPath(path)
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

func (l *Local) DeleteCategory(path string) error {
	l.logger.Info("Removing the category")

	cp := l.constructCategoryPath(path)
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

	// read category contents and check if empty
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

func (l *Local) RenameCategory(path string, name string) error {
	l.logger.Info("Renaming the category")

	// construct filepath for the current path
	ocp := l.constructCategoryPath(path)
	ofp := l.fullPath(ocp)

	// construct filepath for the new path
	fn := l.constructCategoryName(name)
	fp := filepath.Dir(ocp)
	ncp := filepath.Join(fn, fp)
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

	// check if target category already exists
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

	l.logger.Info("Renamed the category")
	return nil
}


/*
	COLLECTION
*/

//  Verifies if provided filepath leads to the collection
func (l *Local) verifyCollectionPath(filename string) bool {
	base := filepath.Base(filename)

	return !strings.HasPrefix(base, "_")
}


/*
	CATEGORY
*/

// Converts the filename to the filesystem compliant category name
func (l *Local) constructCategoryName(name string) string {
	return "_" + name
}

// Converts the provided path to the filesystem compliant category path
func (l *Local) constructCategoryPath(path string) string {
	dirs := strings.Split(path, string(filepath.Separator))

	for i, dir := range dirs {
		dirs[i] = "_" + dir
	}

	newPath := strings.Join(dirs, string(filepath.Separator))

	return newPath
}

// Converts the provided filesystem compliant category path to the neutral path
func (l *Local) deconstructCategoryPath(path string) (string, error) {
	dirs := strings.Split(path, string(filepath.Separator))

	for i, dir := range dirs {
		if strings.HasPrefix(dir, "_") {
			dirs[i] = dir[1:]
		} else {
			l.logger.Error(ErrNotCategoryPath.Error())
			return "", ErrNotCategoryPath
		}
	}

	newPath := strings.Join(dirs, string(filepath.Separator))

	return newPath, nil
}

// Verifies if provided filename is a category
func (l *Local) verifyCategory(filename string) bool {
	base := filepath.Base(filename)

	return strings.HasPrefix(base, "_")
}

// Verifies if provided filepath contains only categories
func (l *Local) verifyCategoryPath(path string) bool {
	dirs := strings.Split(path, string(filepath.Separator))

	for _, dir := range dirs {
		if !strings.HasPrefix(dir, "_") {
			l.logger.Warn("Path contains non-categories")
			return false
		}
	}

	return true
}


/*
	FILEPATH
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
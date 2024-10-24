package files

import (
	"os"
	"path/filepath"
	"strings"
)

/*
	COLLECTION
*/
func (l *Local) VerifyCollection(path string, id string) error {
	l.logger.Info("Verifying the collection")

	cp := l.constructCategoryPath(path)
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

func (l *Local) CreateCollection(path string, id string) error {
	l.logger.Info("Creating the collection")

	cp := l.constructCategoryPath(path)
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

func (l *Local) UpdateCollection(path string, id string, newPath string, newId string) error {
	l.logger.Info("Renaming the collection")

	// Current path
	cp := l.constructCategoryPath(path)
	p := filepath.Join(cp, id)
	fp := l.fullPath(p)

	// Desired path
	ncp := l.constructCategoryPath(newPath)
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

// Converts the filesystem compliant category name to the neutral name
func (l *Local) deconstructCategoryName(name string) (string, error) {
	if strings.HasPrefix(name, "_") {
		return name[1:], nil
	}
	l.logger.Error(ErrNotCategory.Error())
	return "", ErrNotCategory
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
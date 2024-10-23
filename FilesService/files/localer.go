package files

import (
	"os"
	"path/filepath"
	"strings"
)

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

func (l *Local) MakeCategory(path string) error {
	l.logger.Info("Creating the category")

	cp := l.constructCategoryPath(path)
	fp := l.fullPath(cp)

	// check if the directory already exists
	exists, err := l.exists(fp)
	if err != nil {
		return err
	}
	if exists {
		l.logger.Warn(ErrAlreadyExists.Error() + fp)
		return ErrAlreadyExists
	}

	err = l.createFilepath(fp)
	if err != nil {
		return err
	}

	l.logger.Info("Created the category")
	return nil
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
	cleanPath := filepath.Clean(filename)
	base := filepath.Base(cleanPath)

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
	COMMON
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
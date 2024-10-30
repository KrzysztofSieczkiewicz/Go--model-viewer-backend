package files

import (
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/KrzysztofSieczkiewicz/go--model-viewer-backend/FilesService/models"
)

// Implementation of the Storage interface that works for local disk
type Local struct {
	maxFileSize int64			// Max file size in bytes
	basePath    string			// Base path to the storage root
	logger		*slog.Logger	// Logger
}

// Creates new Local filesystem with given basePath and max file size
func NewLocal(basePath string, maxSizeMB int, l *slog.Logger) (*Local, error) {
	// Add logger detail
	logger := l.With(slog.String("store", basePath))

	// convert path to absolute path
	p, err := filepath.Abs(basePath)
	if err != nil {
		return nil, err
	}

	return &Local{
		basePath: p, 
		maxFileSize: int64(maxSizeMB*1024*1000),
		logger: logger,
	}, nil
}

func (l *Local) MaxFileSize() int64 {
    return l.maxFileSize
}

func (l *Local) BasePath() string {
    return l.basePath
}

func (l *Local) Logger() *slog.Logger {
    return l.logger
}


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

func (l *Local) CreateAsset(asset models.Asset, r io.Reader) error {
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
func (l *Local) ListCollectionContents(collection *models.Collection) ([]models.DirContent, error) {
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

func (l *Local) CreateCollection(collection *models.Collection) error {
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

func (l *Local) UpdateCollection(collection *models.Collection, newCollection *models.Collection) error {
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

func (l *Local) DeleteCollection(collection *models.Collection) error {
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
func (l *Local) ListCategoryContents(category *models.Category) ([]models.DirContent, error) {
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


// Returns the absolute path from the relative path
func (l *Local) fullPath(path string) string {
	return filepath.Join(l.basePath, path)
}

// Creates the file under specified filepath
func (l *Local) createFile(fullpath string) (io.WriteCloser, error) {
	l.logger.Info("Creating the file: " + fullpath)

	f, err := os.Create(fullpath)
	if err != nil {
		l.logger.Error(err.Error())
		return nil, ErrFileCreate
	}
	defer f.Close()

	return f, nil
}

// Reads file contents into provided reader
func (l *Local) readFile(fullpath string, writer io.Writer) error {
	l.logger.Info("Reading the file: " + fullpath)

	// open the file
    f, err := os.Open(fullpath)
    if err != nil {
		l.logger.Error(err.Error())
        return ErrFileRead
    }
    defer f.Close()

	// write the file contents into the writer
	_, err = io.Copy(writer, f)
    if err != nil {
		l.logger.Error(err.Error())
        return ErrFileRead
    }

	return nil
}

// Writes reader contents into provided file writer
func (l *Local) writeFile(fullpath string, contents io.Reader) error {
	l.logger.Info("Writing into the file: " + fullpath)

	// create a LimitedReader to limit file size
    limitedReader := &io.LimitedReader{
        R: contents,
        N: l.maxFileSize + 1,
    }

	// open the file
	f, err := os.OpenFile(fullpath, os.O_WRONLY, 0)
    if err != nil {
		l.logger.Error(err.Error())
        return ErrFileRead
    }
    defer f.Close()

	// write the contents to the new file
	_, err = io.Copy(f, limitedReader)
	if err != nil {
		l.logger.Error(err.Error())
		return ErrFileWrite
	}

	// check if filesize limit was reached
	if limitedReader.N == 0 {
		l.logger.Error(ErrWriteSizeExceeded.Error())
		return ErrWriteSizeExceeded
	}

	return nil
}

// Verifies if provided filepath leads to a file
func (l *Local) isFile(fullpath string) (bool, error) {
	l.logger.Info("Verifying the file: " + fullpath)

	file, err := os.Stat(fullpath)
	if err != nil {
		l.logger.Error(err.Error())
		return false, ErrStat
	}
	if file.IsDir() {
		return false, nil
	}

	return true, nil
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

// Changes filepath to the new provided string. Doesn't create directories.
func (l *Local) changeFilepath(old string, new string) error {
	l.logger.Info(fmt.Sprintf("Modifying filepath from: %s\nto: %s", old, new))

	err := os.Rename(old, new)
    if err != nil {
		l.logger.Error(err.Error())
        return ErrRename
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

// Reads directory contents
func (l *Local) readDirectory(fullpath string) ([]fs.FileInfo, error) {
	l.logger.Info("Reading the directory: " + fullpath)

	// open the dir
	dir, err := os.Open(fullpath)
	if err != nil {
		l.logger.Error(err.Error())
		return nil, err
	}
	defer dir.Close()

	// Read directory contents
	entries, err := dir.Readdir(-1)
	if err != nil {
		l.logger.Error(err.Error())
		return nil, err
	}

	l.logger.Info("Finished reading the directory: " + fullpath)

	return entries, nil
}

// Lists directory contents
func (l *Local) listContents(fullpath string)  ([]models.DirContent, error) {
	l.logger.Info("Listing the contents: " + fullpath) 

	// read the directory
	entries, err := l.readDirectory(fullpath)
	if err != nil {
		return nil, err
	}

	// check all entries and assign types
	contents := make([]models.DirContent, 0, len(entries))
	var fileType models.FileType
	for _, entry := range entries {
		if entry.IsDir() {
			fileType = models.FileTypeDirectory
		} else {
			fileType = models.FileTypeFile
		}

		contents = append(
			contents, 
			models.DirContent{
				Filename: entry.Name(),
				FileType: fileType,
			},
		)
	}

	return contents, nil
}
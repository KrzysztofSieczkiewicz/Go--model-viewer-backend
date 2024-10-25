package models

type FileType string

const (
	FileTypeDirectory FileType = "directory"
	FileTypeFile      FileType = "file"
)

type CollectionContent struct {
	Filename string   `json:"filename" validate:"required"`
	FileType FileType `json:"type" validate:"required"`
}
package models

type FileType string

const (
	FileTypeDirectory FileType = "directory"
	FileTypeFile      FileType = "file"
)

type DirContent struct {
	Filename string   `json:"filename" validate:"required"`
	FileType FileType `json:"type" validate:"required"`
}
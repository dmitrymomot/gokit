package storage

import "errors"

var (
	ErrFailedToUploadFile      = errors.New("failed to upload file")
	ErrFailedToListFiles       = errors.New("failed to list files")
	ErrFailedToDeleteFile      = errors.New("failed to delete file")
	ErrFailedToDeleteDirectory = errors.New("failed to delete directory")
)

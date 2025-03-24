package storage

import (
	"errors"
)

// Package-level error declarations.
var (
	// ErrFailedToUploadFile indicates a failure during file upload.
	ErrFailedToUploadFile = errors.New("failed to upload file")

	// ErrFileTooLarge indicates that the file size exceeds the allowed maximum.
	ErrFileTooLarge = errors.New("file size exceeds the maximum allowed limit")

	// ErrFailedToDeleteFile indicates a failure during file deletion.
	ErrFailedToDeleteFile = errors.New("failed to delete file")

	// ErrFailedToDeleteDirectory indicates a failure during directory deletion.
	ErrFailedToDeleteDirectory = errors.New("failed to delete directory")

	// ErrFailedToListFiles indicates a failure during file listing.
	ErrFailedToListFiles = errors.New("failed to list files")

	// ErrInvalidRequest indicates an invalid request or missing required file data.
	ErrInvalidRequest = errors.New("invalid request or missing file data")

	// ErrMissingConfig indicates that the required configuration is missing.
	// This error is returned when the minimum required configuration is not provided.
	ErrMissingConfig = errors.New("missing required configuration")

	// ErrInvalidEndpoint indicates that the provided endpoint is invalid.
	// This error is returned when the endpoint URL is not a valid URL.
	ErrInvalidEndpoint = errors.New("invalid endpoint URL")

	// ErrFailedToLoadConfig indicates a failure to load the AWS configuration.
	ErrFailedToLoadConfig = errors.New("failed to load AWS configuration")
)

package template

import "errors"

var (
	// ErrTemplateNotFound is returned when a template with the given name is not found.
	ErrTemplateNotFound = errors.New("template not found")
	
	// ErrInvalidLayout is returned when the specified layout template is invalid.
	ErrInvalidLayout = errors.New("invalid layout template")
	
	// ErrRenderFailed is returned when template rendering fails.
	ErrRenderFailed = errors.New("failed to render template")
	
	// ErrInvalidDirectory is returned when the template directory is invalid.
	ErrInvalidDirectory = errors.New("invalid template directory")
	
	// ErrInvalidFunction is returned when a template function is invalid.
	ErrInvalidFunction = errors.New("invalid template function")
)
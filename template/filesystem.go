package template

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

// TemplateFS defines the interface for template file systems.
type TemplateFS interface {
	// ParseTemplate parses a template with the given name.
	ParseTemplate(name, extension string) (string, error)
	// Exists checks if a template with the given name exists.
	Exists(name, extension string) bool
}

// LocalFS implements TemplateFS using the local file system.
type LocalFS struct {
	root string
}

// NewLocalFS creates a new LocalFS.
func NewLocalFS(root string) (*LocalFS, error) {
	// Check if directory exists
	info, err := os.Stat(root)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidDirectory, err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("%w: path is not a directory", ErrInvalidDirectory)
	}
	return &LocalFS{root: root}, nil
}

// ParseTemplate implements TemplateFS.ParseTemplate for LocalFS.
func (fs *LocalFS) ParseTemplate(name, extension string) (string, error) {
	path := fs.getPath(name, extension)
	content, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("%w: %s - %v", ErrTemplateNotFound, name, err)
	}
	return string(content), nil
}

// Exists implements TemplateFS.Exists for LocalFS.
func (fs *LocalFS) Exists(name, extension string) bool {
	path := fs.getPath(name, extension)
	_, err := os.Stat(path)
	return err == nil
}

// getPath returns the full path for a template name.
func (fs *LocalFS) getPath(name, extension string) string {
	// If it already has the extension, don't add it
	if filepath.Ext(name) == "" {
		name = name + extension
	}
	return filepath.Join(fs.root, name)
}

// EmbeddedFS implements TemplateFS using an embedded file system.
type EmbeddedFS struct {
	fs fs.FS
}

// NewEmbeddedFS creates a new EmbeddedFS.
func NewEmbeddedFS(filesystem fs.FS) *EmbeddedFS {
	return &EmbeddedFS{fs: filesystem}
}

// ParseTemplate implements TemplateFS.ParseTemplate for EmbeddedFS.
func (efs *EmbeddedFS) ParseTemplate(name, extension string) (string, error) {
	path := efs.getPath(name, extension)
	content, err := fs.ReadFile(efs.fs, path)
	if err != nil {
		return "", fmt.Errorf("%w: %s - %v", ErrTemplateNotFound, name, err)
	}
	return string(content), nil
}

// Exists implements TemplateFS.Exists for EmbeddedFS.
func (efs *EmbeddedFS) Exists(name, extension string) bool {
	path := efs.getPath(name, extension)
	_, err := fs.Stat(efs.fs, path)
	return err == nil
}

// getPath returns the full path for a template name in the embedded file system.
func (efs *EmbeddedFS) getPath(name, extension string) string {
	// If it already has the extension, don't add it
	if filepath.Ext(name) == "" {
		name = name + extension
	}
	return name
}

// WithFS sets the template file system.
func WithFS(filesystem TemplateFS) Option {
	return func(e *Engine) error {
		e.fs = filesystem
		return nil
	}
}
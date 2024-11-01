package dotprompt

import (
	"os"
	"path/filepath"
	"strings"
)

const (
	defaultPath         string = "prompts"
	promptFileExtension string = ".prompt"
)

// FileStoreError represents an error encountered in file store operations.
// It contains a message describing the error and an optional underlying error.
type FileStoreError struct {
	Message string
	Err     error
}

// Error returns the error message contained in the FileStoreError.
func (e FileStoreError) Error() string {
	return e.Message
}

// FileStore represents a file-based storage system for handling prompt files.
type FileStore struct {
	path string
}

// Load retrieves all prompt files from the specified file path and returns a slice of PromptFile objects or an error.
func (f *FileStore) Load() ([]PromptFile, error) {
	promptFiles := make([]PromptFile, 0)

	err := filepath.Walk(f.path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		fileExtension := filepath.Ext(path)
		if strings.ToLower(fileExtension) == promptFileExtension {
			promptFile, promptFileErr := NewPromptFileFromFile(path)
			if promptFileErr != nil {
				return promptFileErr
			}
			promptFiles = append(promptFiles, *promptFile)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return promptFiles, nil
}

// NewFileStore creates a new FileStore instance using the default file path ("prompts").
func NewFileStore() (*FileStore, error) {
	return NewFileStoreFromPath(defaultPath)
}

// NewFileStoreFromPath creates a new FileStore instance from the specified directory path.
func NewFileStoreFromPath(path string) (*FileStore, error) {
	trimmedPath := strings.TrimSpace(path)
	if trimmedPath == "" {
		return nil, &FileStoreError{
			Message: "The specified path is empty",
		}
	}

	info, err := os.Stat(trimmedPath)
	if os.IsNotExist(err) {
		return nil, &FileStoreError{
			Message: "The specified path does not exist",
			Err:     err,
		}
	} else if err != nil {
		return nil, err
	}

	if !info.IsDir() {
		return nil, &FileStoreError{
			Message: "The specified path is not a directory",
		}
	}

	return &FileStore{path: trimmedPath}, nil
}

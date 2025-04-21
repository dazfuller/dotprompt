package dotprompt

import (
	"io/fs"
	"path"
	"path/filepath"
	"strings"
)

// FSStore represents a file-system-based storage system for handling prompt files.
type FSStore struct {
	dirFs fs.ReadDirFS
}

// NewFSStore creates a new FSStore instance using the provided fs.ReadDirFS for reading and managing prompt files.
func NewFSStore(dirFs fs.ReadDirFS) *FSStore {
	return &FSStore{
		dirFs: dirFs,
	}
}

// Load retrieves all prompt files from the root directory and its subdirectories in the file system storage.
// Returns a slice of PromptFile and an error if any issue occurs during the loading process.
func (f *FSStore) Load() ([]PromptFile, error) {
	return f.loadFromDir(".")
}

// loadFromDir recursively loads prompt files from the specified directory path and its subdirectories.
// Returns a slice of PromptFile and an error if reading directories or files fails.
func (f *FSStore) loadFromDir(dirPath string) ([]PromptFile, error) {
	entries, err := fs.ReadDir(f.dirFs, dirPath)
	if err != nil {
		return nil, err
	}

	var promptFiles []PromptFile

	for _, entry := range entries {
		if entry.IsDir() {
			files, loadErr := f.loadFromDir(path.Join(dirPath, entry.Name()))
			if loadErr != nil {
				return nil, loadErr
			}
			promptFiles = append(promptFiles, files...)
		} else {
			if strings.ToLower(filepath.Ext(entry.Name())) != promptFileExtension {
				continue
			}
			file, readErr := fs.ReadFile(f.dirFs, path.Join(dirPath, entry.Name()))
			if readErr != nil {
				return nil, readErr
			}
			pf, pfErr := NewPromptFile(entry.Name(), file)
			if pfErr != nil {
				return nil, pfErr
			}
			promptFiles = append(promptFiles, *pf)
		}
	}

	return promptFiles, nil
}

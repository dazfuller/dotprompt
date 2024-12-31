package dotprompt

import (
	"io/fs"
	"path"
	"path/filepath"
	"strings"
)

type FSStore struct {
	dirFs fs.ReadDirFS
}

func NewFSStore(dirFs fs.ReadDirFS) *FSStore {
	return &FSStore{
		dirFs: dirFs,
	}
}

func (f *FSStore) Load() ([]PromptFile, error) {
	return f.loadFromDir(".")
}

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

package dotprompt

// Loader defines an interface for loading prompt files.
type Loader interface {

	// Load loads prompt files and returns a slice of PromptFile and an error.
	Load() ([]PromptFile, error)
}

// Manager is responsible for managing and storing prompt files, with mapping from their names to PromptFile instances.
type Manager struct {
	PromptFiles map[string]PromptFile
}

// GetPromptFile retrieves the prompt file with the specified name from the manager's stored prompt files.
// Returns the PromptFile and a boolean indicating success of the retrieval.
func (m *Manager) GetPromptFile(name string) (PromptFile, bool) {
	promptFile, ok := m.PromptFiles[name]
	return promptFile, ok
}

// ListPromptFileNames returns a list of all prompt file names managed by the Manager.
func (m *Manager) ListPromptFileNames() []string {
	names := make([]string, 0, len(m.PromptFiles))
	for name := range m.PromptFiles {
		names = append(names, name)
	}
	return names
}

// NewManager creates a new Manager by loading prompt files from the default file store.
// Returns a pointer to the Manager instance or an error if the loading process fails.
func NewManager() (*Manager, error) {
	loader, err := NewFileStore()
	if err != nil {
		return nil, err
	}

	return NewManagerFromLoader(loader)
}

// NewManagerFromLoader initializes and returns a Manager instance by loading prompt files using the provided Loader.
// It returns a pointer to the Manager and an error if the loading process fails.
func NewManagerFromLoader(loader Loader) (*Manager, error) {
	if loader == nil {
		return nil, &PromptError{
			Message: "loader cannot be nil",
		}
	}

	promptFiles, err := loader.Load()
	if err != nil {
		return nil, err
	}

	promptFilesMap := make(map[string]PromptFile)
	for _, promptFile := range promptFiles {
		if _, ok := promptFilesMap[promptFile.Name]; ok {
			return nil, &PromptError{
				Message: "duplicate prompt file name: " + promptFile.Name,
			}
		}
		promptFilesMap[promptFile.Name] = promptFile
	}

	return &Manager{
		PromptFiles: promptFilesMap,
	}, nil
}

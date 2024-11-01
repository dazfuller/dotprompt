package dotprompt

import (
	"fmt"
	"testing"
)

type MockLoader struct {
	PromptFiles []PromptFile
	Err         error
	LoadCount   int
}

func (m *MockLoader) Load() ([]PromptFile, error) {
	m.LoadCount++
	return m.PromptFiles, m.Err
}

func TestNewManager(t *testing.T) {
	mgr, err := NewManager()
	if err != nil {
		t.Fatal(err)
	}

	if len(mgr.PromptFiles) != 1 {
		t.Fatal("Expected 1 prompt file")
	}
}

func TestNewManagerWithLoader_WithValidLoader(t *testing.T) {
	loader := &MockLoader{
		PromptFiles: []PromptFile{
			{
				Name: "example",
			},
		},
		Err: nil,
	}

	_, err := NewManagerFromLoader(loader)
	if err != nil {
		t.Fatal(err)
	}

	if loader.LoadCount != 1 {
		t.Fatal("Expected loader to be called once")
	}
}

func TestNewManagerWithLoader_WithInvalidLoader(t *testing.T) {
	loader := &MockLoader{
		PromptFiles: nil,
		Err:         fmt.Errorf("error"),
	}

	_, err := NewManagerFromLoader(loader)
	if err == nil {
		t.Fatal("Expected error")
	}

	if loader.LoadCount != 1 {
		t.Fatal("Expected loader to be called once")
	}
}

func TestListPromptFiles(t *testing.T) {
	mgr, err := NewManager()
	if err != nil {
		t.Fatal(err)
	}

	names := mgr.ListPromptFileNames()
	if len(names) != 1 {
		t.Fatal("Expected 1 prompt file")
	}
	if names[0] != "example" {
		t.Fatal("Expected example prompt file")
	}
}

func TestGetPromptFile(t *testing.T) {
	mgr, err := NewManager()
	if err != nil {
		t.Fatal(err)
	}

	promptFile, ok := mgr.GetPromptFile("example")

	if !ok {
		t.Fatal("Expected example prompt file")
	}

	if promptFile.Name != "example" {
		t.Fatal("Expected example prompt file")
	}
}

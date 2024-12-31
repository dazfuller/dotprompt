package dotprompt

import (
	"errors"
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

func TestNewManagerWithLoader_WithEmptyLoader(t *testing.T) {
	_, err := NewManagerFromLoader(nil)
	if err == nil {
		t.Fatal("Expected error")
	}

	var promptError *PromptError
	if !errors.As(err, &promptError) {
		t.Fatal("Expected prompt error")
	}

	expectedError := "loader cannot be nil"
	if promptError.Error() != expectedError {
		t.Fatalf("Expected error %s, got %s", expectedError, promptError.Error())
	}
}

func TestNewManagerWithLoader_WithDuplicatePromptFiles(t *testing.T) {
	loader := &MockLoader{
		PromptFiles: []PromptFile{
			{
				Name: "example",
			},
			{
				Name: "example",
			},
		},
	}

	_, err := NewManagerFromLoader(loader)
	if err == nil {
		t.Fatal("Expected error")
	}

	var promptError *PromptError
	if !errors.As(err, &promptError) {
		t.Fatal("Expected prompt error")
	}

	expectedError := "duplicate prompt file name: example"
	if promptError.Error() != expectedError {
		t.Fatalf("Expected error %s, got %s", expectedError, promptError.Error())
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

	promptFile, err := mgr.GetPromptFile("example")

	if err != nil {
		t.Fatal(err)
	}

	if promptFile.Name != "example" {
		t.Fatal("Expected example prompt file")
	}
}

func TestGetPromptFile_WithInvalidPromptName(t *testing.T) {
	mgr, err := NewManager()
	if err != nil {
		t.Fatal(err)
	}

	_, err = mgr.GetPromptFile("does-not-exist")
	if err == nil {
		t.Fatal("Expected error")
	}

	var promptError *PromptError
	if !errors.As(err, &promptError) {
		t.Fatal("Expected prompt error")
	}

	expectedError := "prompt file not found: does-not-exist"
	if promptError.Error() != expectedError {
		t.Fatalf("Expected error %s, got %s", expectedError, promptError.Error())
	}
}

// ExampleNewManager demonstrates the process of creating a new Manager instance which loads from the default "prompts"
// directory, and then fetching a prompt by name from the manager.
func ExampleNewManager() {
	mgr, err := NewManager()
	if err != nil {
		panic(err)
	}

	promptFile, err := mgr.GetPromptFile("example")
	if err != nil {
		panic(err)
	}

	fmt.Println(promptFile.Prompts.System)
	// Output: You are a helpful research assistant who will provide descriptive responses for a given topic and how it impacts society
}

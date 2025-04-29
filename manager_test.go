package dotprompt

import (
	"embed"
	"errors"
	"fmt"
	"testing"
)

//go:embed prompts
var promptFiles embed.FS

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

// Example demonstrates the process of using a Manager to load and generate prompts from a specified prompt file.
func Example() {
	// Create a new manager instance, this will default to loading prompt files which are in the `prompts` folder
	// in the current working directory
	mgr, err := NewManager()
	if err != nil {
		panic(err)
	}

	// Load the prompt file from the manager
	promptFile, err := mgr.GetPromptFile("example")
	if err != nil {
		panic(err)
	}

	// Define the values for the prompt input parameters
	parameters := map[string]interface{}{
		"topic": "bluetooth",
		"style": "used car salesperson",
	}

	// Generate the system and user prompts
	systemPrompt, _ := promptFile.GetSystemPrompt(parameters)
	userPrompt, _ := promptFile.GetUserPrompt(parameters)

	fmt.Println("System prompt:")
	fmt.Println(systemPrompt)
	fmt.Println("\nUser prompt:")
	fmt.Println(userPrompt)

	// Output:
	// System prompt:
	// You are a helpful research assistant who will provide descriptive responses for a given topic and how it impacts society
	//
	// User prompt:
	// Explain the impact of bluetooth on how we engage with technology as a society
	// Can you answer in the style of a used car salesperson
}

// Example_withEmbeddedPrompts demonstrates the process of loading a file with embedded prompts and generating outputs.
// It uses a Manager to retrieve the prompt file and generate system/user prompts based on input parameters.
func Example_withEmbeddedPrompts() {
	// promptFiles is of type `embed.FS`
	// see: https://pkg.go.dev/embed
	store := NewFSStore(promptFiles)

	// Load the prompt file from the manager
	mgr, err := NewManagerFromLoader(store)
	if err != nil {
		panic(err)
	}

	// Load the prompt file from the manager
	promptFile, err := mgr.GetPromptFile("example")
	if err != nil {
		panic(err)
	}

	// Define the values for the prompt input parameters
	parameters := map[string]interface{}{
		"topic": "bluetooth",
		"style": "used car salesperson",
	}

	// Generate the system and user prompts
	systemPrompt, _ := promptFile.GetSystemPrompt(parameters)
	userPrompt, _ := promptFile.GetUserPrompt(parameters)

	fmt.Println("System prompt:")
	fmt.Println(systemPrompt)
	fmt.Println("\nUser prompt:")
	fmt.Println(userPrompt)

	// Output:
	// System prompt:
	// You are a helpful research assistant who will provide descriptive responses for a given topic and how it impacts society
	//
	// User prompt:
	// Explain the impact of bluetooth on how we engage with technology as a society
	// Can you answer in the style of a used car salesperson
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

package dotprompt

import (
	"embed"
	"errors"
	"fmt"
	"testing"
)

//go:embed file-store-tests
var validFs embed.FS

//go:embed test-data
var invalidFs embed.FS

//go:embed prompts
var promptFs embed.FS

func TestFSStore_Load(t *testing.T) {
	store := NewFSStore(validFs)

	promptFiles, err := store.Load()
	if err != nil {
		t.Fatal(err)
	}

	expectedPromptCount := 2

	if len(promptFiles) != expectedPromptCount {
		t.Fatalf("Expected %d prompt files, got %d", expectedPromptCount, len(promptFiles))
	}
}

func TestFSStore_Load_WithInvalidFiles(t *testing.T) {
	store := NewFSStore(invalidFs)

	_, err := store.Load()
	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	// The location contains invalid prompt files so we are expecting an error from loading the files
	var promptError *PromptError
	if !errors.As(err, &promptError) {
		t.Fatalf("Expected PromptError, got %T", err)
	}
}

func ExampleNewManagerFromLoader_withFSStore() {
	// Create a new FSStore instance using the embedded file system, see https://pkg.go.dev/embed for more details
	store := NewFSStore(promptFs)

	// Create a new Manager instance using the FSStore instance
	mgr, err := NewManagerFromLoader(store)
	if err != nil {
		panic(err)
	}

	// Fetch a prompt file by name from the manager
	prompt, err := mgr.GetPromptFile("example")
	if err != nil {
		panic(err)
	}

	fmt.Println(prompt.Prompts.System)
	// Output: You are a helpful research assistant who will provide descriptive responses for a given topic and how it impacts society
}

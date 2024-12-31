package dotprompt

import (
	"embed"
	"errors"
	"testing"
)

//go:embed file-store-tests
var validFs embed.FS

//go:embed test-data
var invalidFs embed.FS

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

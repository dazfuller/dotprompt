package dotprompt

import (
	"errors"
	"slices"
	"testing"
)

func TestNewFileStore(t *testing.T) {
	_, err := NewFileStore()
	if err != nil {
		t.Error(err)
	}
}

func TestNewFileStoreFromPath_WithValidPath(t *testing.T) {
	_, err := NewFileStoreFromPath("./test-data")
	if err != nil {
		t.Error(err)
	}
}

func TestNewFileStoreFromPath_WithInvalidArguments(t *testing.T) {
	tests := []struct {
		name          string
		path          string
		expectedError string
	}{
		{"empty-path", "", "The specified path is empty"},
		{"whitespace-path", " ", "The specified path is empty"},
		{"invalid-path", "./does-not-exist", "The specified path does not exist"},
		{"file-path", "./test-data/basic.prompt", "The specified path is not a directory"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			_, err := NewFileStoreFromPath(test.path)
			if err == nil {
				t.Fatal("Expected error, got nil")
			}

			var fileStoreError *FileStoreError
			if !errors.As(err, &fileStoreError) {
				t.Fatalf("Expected FileStoreError, got %T", err)
			}

			if fileStoreError.Error() != test.expectedError {
				t.Fatalf("Expected error message '%s', got '%s'", test.expectedError, fileStoreError.Error())
			}
		})
	}
}

func TestFileStore_Load(t *testing.T) {
	fileStore, err := NewFileStoreFromPath("./file-store-tests")
	if err != nil {
		t.Fatal(err)
	}

	promptFiles, err := fileStore.Load()
	if err != nil {
		t.Fatal(err)
	}

	if len(promptFiles) != 2 {
		t.Fatalf("Expected 2 prompt files, got %d", len(promptFiles))
	}

	if ok := slices.ContainsFunc(promptFiles, func(promptFile PromptFile) bool { return promptFile.Name == "basic" }); !ok {
		t.Fatal("Expected prompt file with name 'basic'")
	}

	if ok := slices.ContainsFunc(promptFiles, func(promptFile PromptFile) bool { return promptFile.Name == "another-example-with-name" }); !ok {
		t.Fatal("Expected prompt file with name 'another-example-with-name'")
	}
}

func TestFileStore_Load_WithInvalidFiles(t *testing.T) {
	fileStore, err := NewFileStoreFromPath("./test-data")
	if err != nil {
		t.Fatal(err)
	}

	_, err = fileStore.Load()
	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	// The location contains invalid prompt files so we are expecting an error from loading the files
	var promptError *PromptError
	if !errors.As(err, &promptError) {
		t.Fatalf("Expected PromptError, got %T", err)
	}
}

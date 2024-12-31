package dotprompt

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"maps"
	"os"
	"path"
	"reflect"
	"strings"
	"testing"
	"time"
)

type TestStruct struct {
	Item1 string
	Item2 int
}

func (ts TestStruct) String() string {
	return fmt.Sprintf("%s : %d", ts.Item1, ts.Item2)
}

func (of *OutputFormat) String() string {
	switch *of {
	case Text:
		return "text"
	case Json:
		return "json"
	default:
		return "unknown"
	}
}

func TestNewPromptFile_WithBasicPrompt(t *testing.T) {
	promptFile, err := NewPromptFileFromFile("test-data/basic.prompt")
	if err != nil {
		t.Fatal(err)
	}

	expectedParameters := map[string]string{
		"country": "string",
		"style?":  "string",
	}

	expectedDefaults := map[string]interface{}{
		"country": "Malta",
	}

	if promptFile.Name != "basic" {
		t.Errorf("Expected name to be 'basic', got '%s'", promptFile.Name)
	}

	if promptFile.Model != "claude-3-5-sonnet-latest" {
		t.Errorf("Expected model to be 'claude-3-5-sonnet-latest', got '%s'", promptFile.Model)
	}

	if promptFile.Config.OutputFormat != Text {
		t.Errorf("Expected output format to be text, got '%s'", promptFile.Config.OutputFormat.String())
	}

	if *promptFile.Config.MaxTokens != 500 {
		t.Errorf("Expected max tokens to be 500, got '%d'", *promptFile.Config.MaxTokens)
	}

	if *promptFile.Config.Temperature != 0.9 {
		t.Errorf("Expected temperature to be 0.9, got '%f'", *promptFile.Config.Temperature)
	}

	if !reflect.DeepEqual(promptFile.Config.Input.Parameters, expectedParameters) {
		t.Errorf("Expected parameters to be %+v, got %+v", expectedParameters, promptFile.Config.Input.Parameters)
	}

	if !maps.Equal(promptFile.Config.Input.Default, expectedDefaults) {
		t.Errorf("Expected defaults to be %+v, got %+v", expectedDefaults, promptFile.Config.Input.Default)
	}

	if !strings.HasPrefix(promptFile.Prompts.System, "You are a helpful AI assistant that enjoys making penguin related puns.") {
		t.Errorf("Expected system prompt to start with 'You are a helpful AI assistant that enjoys making penguin related puns.', got '%s'", promptFile.Prompts.System)
	}

	if !strings.HasPrefix(promptFile.Prompts.User, "I am looking at going on holiday to {{ country }}") {
		t.Errorf("Expected user prompt to start with 'I am looking at going on holiday to {{ country }}', got '%s'", promptFile.Prompts.User)
	}
}

func TestNewPromptFileFromFile_WithNameFromPromptFile(t *testing.T) {
	promptFile, err := NewPromptFileFromFile("test-data/with-name-json.prompt")
	if err != nil {
		t.Fatal(err)
	}

	if promptFile.Name != "example-with-name" {
		t.Errorf("Expected name to be 'example-with-name', got '%s'", promptFile.Name)
	}

	if promptFile.Config.OutputFormat != Json {
		t.Errorf("Expected output format to be json, got '%s'", promptFile.Config.OutputFormat.String())
	}

	if promptFile.Config.MaxTokens != nil {
		t.Errorf("Expected max tokens to be nil, got '%d'", *promptFile.Config.MaxTokens)
	}

	if promptFile.Config.Temperature != nil {
		t.Errorf("Expected temperature to be nil, got '%f'", *promptFile.Config.Temperature)
	}

	if promptFile.Prompts.System != "" {
		t.Errorf("Expected system prompt to be empty, got '%s'", promptFile.Prompts.System)
	}
}

func TestNewPromptFileFromFile_WithFewShots(t *testing.T) {
	promptFile, err := NewPromptFileFromFile("test-data/basic-fsp.prompt")
	if err != nil {
		t.Fatal(err)
	}

	if len(promptFile.FewShots) != 3 {
		t.Errorf("Expected few shots to be 3, got '%d'", len(promptFile.FewShots))
	}

	if promptFile.FewShots[0].User != "What is Bluetooth" {
		t.Errorf("Expected first few shot user prompt to be 'What is Bluetooth', got '%s'", promptFile.FewShots[0].User)
	}

	if promptFile.FewShots[2].Response != "AI is used in virtual assistants like Siri and Alexa, which understand and respond to voice commands." {
		t.Errorf("Expected third few shot response to be 'AI is used in virtual assistants like Siri and Alexa, which understand and respond to voice commands.', got '%s'", promptFile.FewShots[2].Response)
	}
}

func TestNewPromptFileFromFile_WithInvalidCharsInName(t *testing.T) {
	tests := []struct {
		name     string
		source   string
		expected string
	}{
		{"invalid-characters", "test-data/name-with-invalid-characters.prompt", "dont-use-names-like-this"},
		{"multi-line", "test-data/multiline-name.prompt", "name-which-is-over-multiple-lines"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			promptFile, err := NewPromptFileFromFile(test.source)
			if err != nil {
				t.Fatal(err)
			}

			if promptFile.Name != test.expected {
				t.Errorf("Expected name to be '%s', got '%s'", test.expected, promptFile.Name)
			}
		})
	}
}

func TestNewPromptFileFromFile_WithMissingModel(t *testing.T) {
	promptFile, err := NewPromptFileFromFile("test-data/with-name.prompt")
	if err != nil {
		t.Fatal(err)
	}

	if len(promptFile.Model) != 0 {
		t.Errorf("Expected model to be empty, got '%s'", promptFile.Model)
	}
}

func TestPromptFileFromFile_GetSystemPrompt_WithNonTemplateValues(t *testing.T) {
	tests := []struct {
		name     string
		source   string
		expected string
		exact    bool
	}{
		{"basic", "test-data/basic.prompt", "You are a helpful AI assistant that enjoys making penguin related puns. You should work as many into your response as possible", false},
		{"empty-system-prompt", "test-data/with-name.prompt", "", false},
		{"empty-system-with-json-output", "test-data/with-name-json.prompt", "Please provide the response in JSON", true},
		{"system-prompt-without-json-statement", "test-data/json-missing-messages.prompt", "You are the voice of the guide, you should be authoritative and informative and appeal to a galactic audience. Please provide the response in JSON", true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			promptFile, err := NewPromptFileFromFile(test.source)
			if err != nil {
				t.Fatal(err)
			}

			systemPrompt, err := promptFile.GetSystemPrompt(nil)
			if err != nil {
				t.Fatal(err)
			}

			if test.exact {
				if systemPrompt != test.expected {
					t.Errorf("Expected system prompt to be '%s', got '%s'", test.expected, systemPrompt)
				}
			} else {
				if !strings.Contains(systemPrompt, test.expected) {
					t.Errorf("Expected system prompt to contain '%s', got '%s'", test.expected, systemPrompt)
				}
			}
		})
	}
}

func TestPromptFileFromFile_GetSystemPrompt_WithTemplateParameters(t *testing.T) {
	promptFile, err := NewPromptFileFromFile("test-data/basic-sp-template.prompt")
	if err != nil {
		t.Fatal(err)
	}

	generatedTime := time.Date(2006, 1, 2, 3, 5, 5, 0, time.UTC)
	systemPrompt, err := promptFile.GetSystemPrompt(map[string]interface{}{
		"country":   "Italy",
		"generated": generatedTime,
	})
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(systemPrompt, "You are a helpful AI assistant that who has extensive local knowledge of Italy\nYou should append each response with the text `Generated: <date>` where `<date>` is replaced with the current date, for example:\n`Generated: Monday 02 Jan 2006`") {
		t.Errorf("Expected system prompt to contain 'You are a helpful AI assistant that who has extensive local knowledge of Italy\nYou should append each response with the text `Generated: <date>` where `<date>` is replaced with the current date, for example:\n`Generated: Monday 02 Jan 2006`', got '%s'", systemPrompt)
	}
}

func TestNewPromptFileFromFile_InvalidPath_ReturnsError(t *testing.T) {
	_, err := NewPromptFileFromFile("test-data/invalid-path.prompt")
	if err == nil {
		t.Fatal("Expected error, got none")
	}

	var pathError *fs.PathError
	if !errors.As(err, &pathError) {
		t.Fatal("Expected error to be of type fs.PathError")
	}
}

func TestNewPromptFile_WithInvalidContent_ReturnsError(t *testing.T) {
	_, err := NewPromptFile("invalid", []byte("<xml>"))
	if err == nil {
		t.Fatal("Expected error, got none")
	}

	var promptError *PromptError
	if !errors.As(err, &promptError) {
		t.Fatal("Expected error to be of type PromptError")
	}

	expectedError := "failed to parse prompt file"

	if !strings.HasPrefix(promptError.Error(), expectedError) {
		t.Errorf("Expected error to start with '%s', got '%s'", expectedError, promptError.Error())
	}
}

func TestNewPromptFileFromFile_WithInvalidYamlContent(t *testing.T) {
	tests := []struct {
		name          string
		source        string
		expectedError string
	}{
		{"invalid-yaml", "test-data/basic-broken.prompt", "failed to parse prompt file"},
		{"invalid-param-types", "test-data/invalid-params.prompt", "invalid data type for parameter oops: cat"},
		{"missing-user-prompt", "test-data/missing-user-prompt.prompt", "no user prompt template was provided in the prompt file"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			_, err := NewPromptFileFromFile(test.source)
			if err == nil {
				t.Fatal("Expected error, got none")
			}

			var promptError *PromptError
			if !errors.As(err, &promptError) {
				t.Fatal("Expected error to be of type PromptError")
			}

			if !strings.HasPrefix(promptError.Error(), test.expectedError) {
				t.Errorf("Expected error to start with '%s', got '%s'", test.expectedError, promptError.Error())
			}
		})
	}
}

func TestNewPromptFile_WithNamePart_CleansTheName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"windows-newlines", "clean\r\n\r\nthis name", "clean-this-name"},
		{"valid-name", "do-not-clean", "do-not-clean"},
		{"mixed-case", "My COOL nAMe", "my-cool-name"},
		{"invalid-characters", "this <is .pretty> un*cl()ean", "this-is-pretty-unclean"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			promptFile, err := NewPromptFile(test.input, []byte("prompts:\n  system: System prompt\n  user: User prompt"))
			if err != nil {
				t.Fatal(err)
			}

			if promptFile.Name != test.expected {
				t.Errorf("Expected name to be '%s', got '%s'", test.expected, promptFile.Name)
			}
		})
	}
}

func TestNewPromptFile_WithInvalidOutputFormat(t *testing.T) {
	_, err := NewPromptFile("invalid-format", []byte("config:\n  outputFormat: xml"))
	if err == nil {
		t.Fatal("Expected error, got none")
	}

	expected := "invalid output format: xml"
	if !strings.Contains(err.Error(), expected) {
		t.Errorf("Expected error to contain '%s', got '%s'", expected, err.Error())
	}
}

func TestNewPromptFile_WithInvalidName_ReturnsError(t *testing.T) {
	_, err := NewPromptFile("++ -- ()", []byte("prompts:\n  system: System prompt\n  user: User prompt"))
	if err == nil {
		t.Fatal("Expected error, got none")
	}

	var promptError *PromptError
	if !errors.As(err, &promptError) {
		t.Fatal("Expected error to be of type PromptError")
	}

	expectedError := "the prompt file name, once cleaned, is empty"

	if !strings.HasPrefix(promptError.Error(), expectedError) {
		t.Errorf("Expected error to start with '%s', got '%s'", expectedError, promptError.Error())
	}
}

func TestPromptFile_GetUserPrompt_WithDefaultValues(t *testing.T) {
	tests := []struct {
		name       string
		source     string
		parameters map[string]interface{}
		expected   string
	}{
		{"empty-parameters", "test-data/basic.prompt", nil, "I am looking at going on holiday to Malta and would like to know more about it, what can you tell me?\n"},
		{"required-parameters", "test-data/basic.prompt", map[string]interface{}{"country": "Antarctica"}, "I am looking at going on holiday to Antarctica and would like to know more about it, what can you tell me?\n"},
		{"unused-parameters", "test-data/basic.prompt", map[string]interface{}{"country": "Antarctica", "unused": "value"}, "I am looking at going on holiday to Antarctica and would like to know more about it, what can you tell me?\n"},
		{"with-optional-parameters", "test-data/basic.prompt", map[string]interface{}{"country": "Antarctica", "style": "pirate"}, "I am looking at going on holiday to Antarctica and would like to know more about it, what can you tell me?\nCan you answer in the style of a pirate\n"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			promptFile, err := NewPromptFileFromFile(test.source)
			if err != nil {
				t.Fatal(err)
			}

			userPrompt, err := promptFile.GetUserPrompt(test.parameters)
			if err != nil {
				t.Fatal(err)
			}

			if userPrompt != test.expected {
				t.Errorf("Expected user prompt to be '%s', got '%s'", test.expected, userPrompt)
			}
		})
	}
}

func TestPromptFile_GetUserPrompt_WithInvalidParameters_ReturnsError(t *testing.T) {
	tests := []struct {
		name       string
		source     string
		parameters map[string]interface{}
	}{
		{"empty-parameters", "test-data/required-parameters.prompt", nil},
		{"missing-parameters", "test-data/required-parameters.prompt", map[string]interface{}{"not": "this"}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			promptFile, err := NewPromptFileFromFile(test.source)
			if err != nil {
				t.Fatal(err)
			}

			_, err = promptFile.GetUserPrompt(test.parameters)
			if err == nil {
				t.Fatal("Expected error, got none")
			}

			var promptError *PromptError
			if !errors.As(err, &promptError) {
				t.Fatal("Expected error to be of type PromptError")
			}

			expectedError := "no value provided for parameter name"
			if !strings.HasPrefix(promptError.Error(), expectedError) {
				t.Errorf("Expected error to start with '%s', got '%s'", expectedError, promptError.Error())
			}
		})
	}
}

func TestPromptFile_GetUserPrompt_WithAllParameterTypes(t *testing.T) {
	promptFile, err := NewPromptFileFromFile("test-data/param-types.prompt")
	if err != nil {
		t.Fatal(err)
	}

	prompt, err := promptFile.GetUserPrompt(map[string]interface{}{
		"param1": "Arthur Dent",
		"param2": 42,
		"param3": true,
		"param4": time.Date(2024, 1, 2, 3, 4, 5, 0, time.UTC),
		"param5": struct{ SEP bool }{true},
		"param6": TestStruct{Item1: "Hello", Item2: 12},
	})
	if err != nil {
		t.Fatal(err)
	}

	expected := "Parameter 1: Arthur Dent\nParameter 2: 42\nParameter 3: true\nParameter 4: 2024-01-02 03:04:05 +0000\nParameter 5: {SEP:true}\nParameter 6: Hello : 12"
	if prompt != expected {
		t.Errorf("Expected prompt to be '%s', got '%s'", expected, prompt)
	}
}

func TestPromptFile_GetUserPrompt_WithInvalidParameterValues_ReturnsError(t *testing.T) {
	tests := []struct {
		name       string
		parameters map[string]interface{}
		expected   string
	}{
		{
			"invalid-string",
			map[string]interface{}{
				"param1": 1,
				"param2": 42,
				"param3": true,
				"param4": time.Date(2024, 1, 2, 3, 4, 5, 0, time.UTC),
				"param5": struct{ SEP bool }{true},
				"param6": TestStruct{Item1: "Hello", Item2: 12},
			},
			"parameter param1 is not a string",
		},
		{
			"invalid-number",
			map[string]interface{}{
				"param1": "Arthur Dent",
				"param2": "42",
				"param3": true,
				"param4": time.Date(2024, 1, 2, 3, 4, 5, 0, time.UTC),
				"param5": struct{ SEP bool }{true},
				"param6": TestStruct{Item1: "Hello", Item2: 12},
			},
			"parameter param2 is not a number",
		},
		{
			"invalid-bool",
			map[string]interface{}{
				"param1": "Arthur Dent",
				"param2": 42,
				"param3": "nope",
				"param4": time.Date(2024, 1, 2, 3, 4, 5, 0, time.UTC),
				"param5": struct{ SEP bool }{true},
				"param6": TestStruct{Item1: "Hello", Item2: 12},
			},
			"parameter param3 is not a bool",
		},
		{
			"invalid-datetime",
			map[string]interface{}{
				"param1": "Arthur Dent",
				"param2": 42,
				"param3": true,
				"param4": "2024-02-01",
				"param5": struct{ SEP bool }{true},
				"param6": TestStruct{Item1: "Hello", Item2: 12},
			},
			"parameter param4 is not a datetime",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			promptFile, err := NewPromptFileFromFile("test-data/param-types.prompt")
			if err != nil {
				t.Fatal(err)
			}

			_, err = promptFile.GetUserPrompt(test.parameters)
			if err == nil {
				t.Fatal("Expected error, got none")
			}

			var promptError *PromptError
			if !errors.As(err, &promptError) {
				t.Fatal("Expected error to be of type PromptError")
			}

			if promptError.Error() != test.expected {
				t.Errorf("Expected error to be '%s', got '%s'", test.expected, promptError.Error())
			}
		})
	}
}

func TestPromptFile_GetUserPrompt_WithNumericValues_ReturnsCorrectPrompt(t *testing.T) {
	tests := []struct {
		name     string
		value    interface{}
		expected string
	}{
		{"int", 8, "Pass"},
		{"int8", int8(8), "Pass"},
		{"int16", int16(8), "Pass"},
		{"int32", int32(8), "Pass"},
		{"int64", int64(8), "Pass"},
		//{"uint", uint(8), "Pass"},
		//{"uint8", uint8(8), "Pass"},
		//{"uint16", uint16(8), "Pass"},
		//{"uint32", uint32(8), "Pass"},
		//{"uint64", uint64(8), "Pass"},
		{"float32", float32(8), "Pass"},
		{"float64", float64(8), "Pass"},
		{"low-value", 2, ""},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			promptFile, err := NewPromptFileFromFile("test-data/numeric-types.prompt")
			if err != nil {
				t.Fatal(err)
			}

			prompt, err := promptFile.GetUserPrompt(map[string]interface{}{"param1": test.value})
			if err != nil {
				t.Fatal(err)
			}

			if prompt != test.expected {
				t.Errorf("Expected prompt to be '%s', got '%s'", test.expected, prompt)
			}
		})
	}
}

func TestPromptFile_Serialize(t *testing.T) {
	promptFile := PromptFile{
		Name:  "serialize-test",
		Model: "gpt-4o",
		Config: PromptConfig{
			OutputFormat: Json,
			Input: InputSchema{
				Parameters: map[string]string{
					"param1": "number",
				},
			},
		},
		Prompts: Prompts{
			System: "system",
			User:   "user",
		},
	}

	expected := []byte("name: serialize-test\nmodel: gpt-4o\nconfig:\n  outputFormat: json\n  input:\n    parameters:\n      param1: number\nprompts:\n  system: system\n  user: user\n")

	serialized, err := promptFile.Serialize()
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(serialized, expected) {
		t.Errorf("Expected serialized prompt file to be '%s', got '%s'", expected, serialized)
	}
}

func TestPromptFile_ToFile(t *testing.T) {
	promptFile := PromptFile{
		Model: "gpt-4o",
		Config: PromptConfig{
			OutputFormat: Json,
			Input: InputSchema{
				Parameters: map[string]string{
					"param1": "number",
				},
			},
		},
		Prompts: Prompts{
			System: "system",
			User:   "user",
		},
	}

	filePath := path.Join(os.TempDir(), "to-file-test.prompt")

	err := promptFile.ToFile(filePath)
	if err != nil {
		t.Fatal(err)
	}
}

// ExampleNewPromptFileFromFile demonstrates loading a prompt file from a given path and then passing in values to
// the template to generate the user prompt.
func ExampleNewPromptFileFromFile() {
	promptFile, err := NewPromptFileFromFile("test-data/basic.prompt")
	if err != nil {
		panic(err)
	}

	prompt, err := promptFile.GetUserPrompt(map[string]interface{}{"country": "Malta"})
	if err != nil {
		panic(err)
	}

	fmt.Println(prompt)
	// Output: I am looking at going on holiday to Malta and would like to know more about it, what can you tell me?
}
